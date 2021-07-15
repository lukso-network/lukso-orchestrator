package consensus

import (
	"context"
	"fmt"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db/iface"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/rpc/api/events"
	"github.com/lukso-network/lukso-orchestrator/shared"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	log "github.com/sirupsen/logrus"
	"sync"
)

type databaseErrors struct {
	vanguardErr error
	pandoraErr  error
	realmErr    error
}

type invalidationWorkPayload struct {
	invalidationStartRealmSlot uint64
	fromSlot                   uint64
	possibleSkippedPair        []*events.RealmPair
	pandoraHashes              []*types.HeaderHash
	vanguardHashes             []*types.HeaderHash
	pandoraOrphans             map[uint64]*types.HeaderHash
	vanguardOrphans            map[uint64]*types.HeaderHash
}

// Service This part could be moved to other place during refactor, might be registered as a service
type Service struct {
	VanguardHeaderHashDB        iface.VanHeaderAccessDatabase
	PandoraHeaderHashDB         iface.PanHeaderAccessDatabase
	RealmDB                     iface.RealmAccessDatabase
	VanguardHeadersChan         chan *types.HeaderHash
	VanguardConsensusInfoChan   chan *types.MinimalEpochConsensusInfo
	PandoraHeadersChan          chan *types.HeaderHash
	stopChan                    chan struct{}
	canonicalizeChan            chan uint64
	isWorking                   bool
	canonicalizeLock            *sync.Mutex
	invalidationWorkPayloadChan chan *invalidationWorkPayload
	errChan                     chan databaseErrors
	started                     bool
	ctx                         context.Context
}

// Start service should be registered only after Pandora and Vanguard notified about:
// - consensus info (Vanguard)
// - pendingHeaders (Vanguard)
// - pendingHeaders (Pandora)
// In current implementation we use debounce to determine state of syncing
func (service *Service) Start() {
	// There might be multiple scenarios that will trigger different slot required to trigger the canonicalize
	service.workLoop()

	return
}

func (service *Service) Stop() error {
	service.stopChan <- struct{}{}
	close(service.stopChan)

	return nil
}

func (service *Service) Status() error {
	return nil
}

var _ shared.Service = &Service{}

func New(
	ctx context.Context,
	database db.Database,
	vanguardHeadersChan chan *types.HeaderHash,
	vanguardConsensusInfoChan chan *types.MinimalEpochConsensusInfo,
	pandoraHeadersChan chan *types.HeaderHash,
) (service *Service) {
	stopChan := make(chan struct{})
	canonicalizeChain := make(chan uint64)
	invalidationWorkPayloadChan := make(chan *invalidationWorkPayload, 10000)
	errChan := make(chan databaseErrors, 10000)

	return &Service{
		VanguardHeaderHashDB:      database,
		PandoraHeaderHashDB:       database,
		RealmDB:                   database,
		VanguardHeadersChan:       vanguardHeadersChan,
		VanguardConsensusInfoChan: vanguardConsensusInfoChan,
		PandoraHeadersChan:        pandoraHeadersChan,
		stopChan:                  stopChan,
		canonicalizeChan:          canonicalizeChain,
		canonicalizeLock:          &sync.Mutex{},
		// invalidationWorkPayloadChan is internal so it is created on the fly
		invalidationWorkPayloadChan: invalidationWorkPayloadChan,
		// errChan is internal so it is created on the fly
		errChan: errChan,
		ctx:     ctx,
	}
}

// Canonicalize must be called numerous of times with different from slot
// new slots may arrive after canonicalization, so Canonicalize must be invoked again
// function must be working only on started service
func (service *Service) Canonicalize(
	fromSlot uint64,
	batchLimit uint64,
) (
	vanguardErr error,
	pandoraErr error,
	realmErr error,
) {
	if nil == service {
		realmErr = fmt.Errorf("cannot start canonicalization without service")

		return
	}

	if !service.started {
		log.WithField("tip", "use service.Start() before using Canonicalize").
			Fatal("I cannot Canonicalize on not started service")

		return
	}

	vanguardHashDB := service.VanguardHeaderHashDB
	pandoraHeaderHashDB := service.PandoraHeaderHashDB
	realmDB := service.RealmDB
	errChan := service.errChan

	// Short circuit, do not invalidate when databases are not present.
	if nil == vanguardHashDB || nil == pandoraHeaderHashDB || nil == realmDB {
		log.WithField("vanguardHashDB", vanguardHashDB).
			WithField("pandoraHeaderHashDB", pandoraHeaderHashDB).
			WithField("realmDB", realmDB).Warn("Databases are not present")
		return
	}

	log.Info("I am starting to Canonicalize in batches")
	select {
	case <-service.stopChan:
		service.isWorking = false
		log.Info("I stop Invalidation")

		return
	case databaseErrorList := <-errChan:
		vanguardErr = databaseErrorList.vanguardErr
		pandoraErr = databaseErrorList.pandoraErr
		realmErr = databaseErrorList.realmErr

		return
	default:
		// If higher slot was found and is valid all the gaps below must me treated as skipped
		// Any other should be treated as pending
		// When Sharding info comes we can determine slashing and Invalid state
		// SIDE NOTE: This is invalid, when a lot of blocks were just simply not present yet due to the network traffic
		err := invokeInvalidation(
			vanguardHashDB,
			pandoraHeaderHashDB,
			realmDB,
			fromSlot,
			batchLimit,
			errChan,
			service.invalidationWorkPayloadChan,
		)

		if nil != err {
			return
		}

		work := <-service.invalidationWorkPayloadChan
		handlePreparedWork(vanguardHashDB, pandoraHeaderHashDB, realmDB, work, errChan)
	}

	return
}

// workLoop should be responsible of handling multiple events and resolving them
// Assumption is that if you want to validate pending queue you should receive information from Vanguard and Pandora
// TODO: handle reorgs
// TODO: consider working on MinimalConsensusInfo
func (service *Service) workLoop() {
	var onceAtTheTime = sync.Once{}
	service.started = true
	verifiedSlotWorkLoopStart := service.RealmDB.LatestVerifiedRealmSlot()
	log.WithField("verifiedSlotWorkLoopStart", verifiedSlotWorkLoopStart).
		Info("I am starting the work loop")

	possiblePendingWork := make([]*types.HeaderHash, 0)

	// This is arbitrary, it may be less or more. Depends on the approach
	// Create merged channel
	mergedChannel := merge(service.VanguardHeadersChan, service.PandoraHeadersChan)
	// Provide handlers for debounce
	mergedChannelHandler := func(slot uint64) {
		possiblePendingWork = make([]*types.HeaderHash, 0)

		if !service.isWorking {
			onceAtTheTime = sync.Once{}
		}

		onceAtTheTime.Do(func() {
			defer func() {
				service.isWorking = false
			}()
			service.isWorking = true
			log.WithField("latestVerifiedSlot", slot).
				Info("I am starting canonicalization")

			vanguardErr, pandoraErr, realmErr := service.Canonicalize(slot, 50000)

			log.WithField("latestVerifiedSlot", slot).
				Info("After canonicalization")

			if nil != vanguardErr {
				log.WithField("canonicalize", "vanguardErr").Debug(vanguardErr)
			}

			if nil != pandoraErr {
				log.WithField("canonicalize", "pandoraErr").Debug(pandoraErr)
			}

			if nil != realmErr {
				log.WithField("canonicalize", "realmErr").Debug(realmErr)
			}
		})
	}

	go func() {
		for {
			select {
			case header := <-mergedChannel:
				possiblePendingWork = append(possiblePendingWork, header)
				log.WithField("cause", "mergedChannel").
					Debug("I am pushing header to merged channel")
				mergedChannelHandler(service.RealmDB.LatestVerifiedRealmSlot())
			case slot := <-service.canonicalizeChan:
				log.WithField("cause", "canonicalizeChan").
					Debug("I am pushing header to merged channel")
				mergedChannelHandler(slot)
			case <-service.stopChan:
				log.WithField("canonicalize", "stop").Info("Received stop signal")

				return
			}
		}
	}()
}

// invokeInvalidation will prepare payload for crawler and will push it through the channel
// if any error occurs it will be pushed back to databaseErrorsChan
// if any skip must happen it will be send information via skipChan
func invokeInvalidation(
	vanguardHashDB db.VanguardHeaderHashDB,
	pandoraHeaderHashDB db.PandoraHeaderHashDB,
	realmDB db.RealmDB,
	fromSlot uint64,
	batchLimit uint64,
	databaseErrorsChan chan databaseErrors,
	invalidationWorkPayloadChan chan *invalidationWorkPayload,
) (err error) {
	possibleSkippedPair := make([]*events.RealmPair, 0)
	latestSavedVerifiedRealmSlot := realmDB.LatestVerifiedRealmSlot()
	invalidationStartRealmSlot := latestSavedVerifiedRealmSlot

	if fromSlot > latestSavedVerifiedRealmSlot {
		databaseErrorsChan <- databaseErrors{realmErr: fmt.Errorf("I cannot start invalidation without root")}

		return
	}

	log.WithField("latestSavedVerifiedRealmSlot", latestSavedVerifiedRealmSlot).
		WithField("slot", fromSlot).
		Info("Invalidation starts")

	pandoraHeaderHashes, err := pandoraHeaderHashDB.PandoraHeaderHashes(fromSlot, batchLimit)

	if nil != err {
		log.WithField("cause", "Failed to invalidate pending queue").Error(err)
		databaseErrorsChan <- databaseErrors{pandoraErr: err}

		return
	}

	vanguardBlockHashes, err := vanguardHashDB.VanguardHeaderHashes(fromSlot, batchLimit)

	if nil != err {
		log.WithField("cause", "Failed to invalidate pending queue").Error(err)
		databaseErrorsChan <- databaseErrors{vanguardErr: err}

		return
	}

	pandoraRange := len(pandoraHeaderHashes)
	vanguardRange := len(vanguardBlockHashes)

	pandoraOrphans := map[uint64]*types.HeaderHash{}
	vanguardOrphans := map[uint64]*types.HeaderHash{}

	// You wont match anything, so short circuit
	if pandoraRange < 1 || vanguardRange < 1 {
		log.WithField("pandoraRange", pandoraRange).WithField("vanguardRange", vanguardRange).
			Trace("Not enough blocks to start invalidation")

		err = fmt.Errorf("not enough blocks to start invalidation")
		databaseErrorsChan <- databaseErrors{realmErr: fmt.Errorf("not enough blocks to start invalidation")}

		return
	}

	log.WithField("pandoraRange", pandoraRange).WithField("vanguardRange", vanguardRange).
		Trace("Invalidation with range of blocks")

	invalidationWorkPayloadChan <- &invalidationWorkPayload{
		invalidationStartRealmSlot: invalidationStartRealmSlot,
		fromSlot:                   fromSlot,
		possibleSkippedPair:        possibleSkippedPair,
		pandoraHashes:              pandoraHeaderHashes,
		vanguardHashes:             vanguardBlockHashes,
		pandoraOrphans:             pandoraOrphans,
		vanguardOrphans:            vanguardOrphans,
	}

	return
}

func merge(cs ...<-chan *types.HeaderHash) <-chan *types.HeaderHash {
	out := make(chan *types.HeaderHash)
	var wg sync.WaitGroup
	wg.Add(len(cs))
	for _, c := range cs {
		go func(c <-chan *types.HeaderHash) {
			for v := range c {
				out <- v
			}
			wg.Done()
		}(c)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
