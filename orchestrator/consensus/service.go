package consensus

import (
	"context"
	"fmt"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db/iface"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/rpc/api/events"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/utils"
	"github.com/lukso-network/lukso-orchestrator/shared"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

// This part could be moved to other place during refactor, might be registered as a service
type Service struct {
	VanguardHeaderHashDB      iface.VanHeaderAccessDatabase
	PandoraHeaderHashDB       iface.PanHeaderAccessDatabase
	RealmDB                   iface.RealmAccessDatabase
	VanguardHeadersChan       chan *types.HeaderHash
	VanguardConsensusInfoChan chan *types.MinimalEpochConsensusInfo
	PandoraHeadersChan        chan *types.HeaderHash
	stopChan                  chan bool
	canonicalizeChan          chan uint64
}

// This service should be registered only after Pandora and Vanguard notified about:
// - consensus info (Vanguard)
// - pendingHeaders (Vanguard)
// - pendingHeaders (Pandora)
// In current implementation we use debounce to determine state of syncing
func (service *Service) Start() {
	go func() {
		for {
			select {
			case slot := <-service.canonicalizeChan:
				log.WithField("latestVerifiedSlot", slot).
					Info("I am starting canonicalization")
				vanguardErr, pandoraErr, realmErr := service.Canonicalize(slot, 50000)

				if nil != vanguardErr {
					log.WithField("canonicalize", "vanguardErr").Debug(vanguardErr)
				}

				if nil != pandoraErr {
					log.WithField("canonicalize", "pandoraErr").Debug(pandoraErr)
				}

				if nil != realmErr {
					log.WithField("canonicalize", "realmErr").Debug(realmErr)
				}
			case stop := <-service.stopChan:
				if stop {
					log.WithField("canonicalize", "stop").Info("Received stop signal")
					return
				}
			}
		}
	}()

	// There might be multiple scenarios that will trigger different slot required to trigger the canonicalize
	service.workLoop()

	return
}

func (service *Service) Stop() error {
	service.stopChan <- true

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
	stopChan := make(chan bool)
	canonicalizeChain := make(chan uint64)

	return &Service{
		VanguardHeaderHashDB:      database,
		PandoraHeaderHashDB:       database,
		RealmDB:                   database,
		VanguardHeadersChan:       vanguardHeadersChan,
		VanguardConsensusInfoChan: vanguardConsensusInfoChan,
		PandoraHeadersChan:        pandoraHeadersChan,
		stopChan:                  stopChan,
		canonicalizeChan:          canonicalizeChain,
	}
}

// Canonicalize must be called numerous of times with different from slot
// new slots may arrive after canonicalization, so Canonicalize must be invoked again
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

	vanguardHashDB := service.VanguardHeaderHashDB
	pandoraHeaderHashDB := service.PandoraHeaderHashDB
	realmDB := service.RealmDB

	// Short circuit, do not invalidate when databases are not present.
	if nil == vanguardHashDB || nil == pandoraHeaderHashDB || nil == realmDB {
		return
	}

	log.Info("I am starting to Canonicalize in batches")
	select {
	case stop := <-service.stopChan:
		if stop {
			log.Info("I stop Invalidation")
			return
		}
	default:
		// If higher slot was found and is valid all the gaps between must me treated as invalid and discarded
		// SIDE NOTE: This is invalid, when a lot of blocks were just simply not present yet due to the network traffic
		possibleInvalidPair := make([]*events.RealmPair, 0)
		latestSavedVerifiedRealmSlot := realmDB.LatestVerifiedRealmSlot()

		if fromSlot > latestSavedVerifiedRealmSlot {
			realmErr = fmt.Errorf("I cannot start invalidation without root")

			return
		}

		log.WithField("latestSavedVerifiedRealmSlot", latestSavedVerifiedRealmSlot).
			WithField("slot", fromSlot).
			Info("Invalidation starts")

		pandoraHeaderHashes, err := pandoraHeaderHashDB.PandoraHeaderHashes(fromSlot, batchLimit)

		if nil != err {
			log.WithField("cause", "Failed to invalidate pending queue").Error(err)
			return
		}

		vanguardBlockHashes, err := vanguardHashDB.VanguardHeaderHashes(fromSlot, batchLimit)

		if nil != err {
			log.WithField("cause", "Failed to invalidate pending queue").Error(err)
			realmErr = err

			return
		}

		pandoraRange := len(pandoraHeaderHashes)
		vanguardRange := len(vanguardBlockHashes)

		log.WithField("pandoraRange", pandoraRange).WithField("vanguardRange", vanguardRange).
			Trace("Invalidation with range of blocks")

		// You wont match anything, so short circuit
		if pandoraRange < 1 || vanguardRange < 1 {
			return
		}

		// TODO: move it to memory, and save in batch
		// This is quite naive, but should work
		for index, vanguardBlockHash := range vanguardBlockHashes {
			slotToCheck := fromSlot + uint64(index)

			if len(pandoraHeaderHashes) <= index {
				break
			}

			pandoraHeaderHash := pandoraHeaderHashes[index]

			// Potentially skipped slot
			if nil == pandoraHeaderHash && nil == vanguardBlockHash {
				possibleInvalidPair = append(possibleInvalidPair, &events.RealmPair{
					Slot:          slotToCheck,
					VanguardHash:  nil,
					PandoraHashes: nil,
				})

				continue
			}

			// I dont know yet, if it is true.
			// In my opinion INVALID state is 100% accurate only with blockShard verification approach
			// TODO: add additional Sharding info check VanguardBlock -> PandoraHeaderHash when implementation on vanguard side will be ready
			if nil == pandoraHeaderHash {
				vanguardHeaderHash := &types.HeaderHash{
					HeaderHash: vanguardBlockHash.HeaderHash,
					Status:     types.Pending,
				}
				vanguardErr = vanguardHashDB.SaveVanguardHeaderHash(slotToCheck, vanguardHeaderHash)

				possibleInvalidPair = append(possibleInvalidPair, &events.RealmPair{
					Slot:          slotToCheck,
					VanguardHash:  vanguardHeaderHash,
					PandoraHashes: nil,
				})

				continue
			}

			if nil == vanguardBlockHash {
				currentPandoraHeaderHash := &types.HeaderHash{
					HeaderHash: pandoraHeaderHash.HeaderHash,
					Status:     types.Pending,
				}
				currentPandoraHeaderHashes := make([]*types.HeaderHash, 1)
				currentPandoraHeaderHashes[0] = currentPandoraHeaderHash
				pandoraErr = pandoraHeaderHashDB.SavePandoraHeaderHash(slotToCheck, currentPandoraHeaderHash)

				possibleInvalidPair = append(possibleInvalidPair, &events.RealmPair{
					Slot:          slotToCheck,
					VanguardHash:  nil,
					PandoraHashes: currentPandoraHeaderHashes,
				})

				continue
			}

			if types.Verified != vanguardBlockHash.Status {
				vanguardErr = vanguardHashDB.SaveVanguardHeaderHash(slotToCheck, &types.HeaderHash{
					HeaderHash: vanguardBlockHash.HeaderHash,
					Status:     types.Verified,
				})
			}

			if types.Verified != pandoraHeaderHash.Status {
				pandoraErr = pandoraHeaderHashDB.SavePandoraHeaderHash(slotToCheck, &types.HeaderHash{
					HeaderHash: pandoraHeaderHash.HeaderHash,
					Status:     types.Verified,
				})
			}

			if nil != vanguardErr || nil != pandoraErr {
				break
			}

			realmErr = realmDB.SaveLatestVerifiedRealmSlot(slotToCheck)
			pandoraErr = pandoraHeaderHashDB.SaveLatestPandoraSlot()
			vanguardErr = vanguardHashDB.SaveLatestVanguardSlot()

			if nil != realmErr || nil != pandoraErr || nil != vanguardErr {
				log.WithField("vanguardErr", vanguardErr).
					WithField("pandoraErr", pandoraErr).
					WithField("realmErr", realmErr).
					Error("Got error during compare of VanguardHashes against PandoraHashes")
				break
			}

			vanguardErr = vanguardHashDB.SaveLatestVanguardHeaderHash()
			pandoraErr = pandoraHeaderHashDB.SaveLatestPandoraHeaderHash()

			if nil != vanguardErr || nil != pandoraErr {
				break
			}
		}

		if nil != vanguardErr || nil != pandoraErr || nil != realmErr {
			log.WithField("vanguardErr", vanguardErr).
				WithField("pandoraErr", pandoraErr).
				WithField("realmErr", realmErr).
				Error("Got error during invalidation of pending queue")
			return
		}

		// Resolve state of possible invalid pairs
		latestSavedVerifiedRealmSlot = realmDB.LatestVerifiedRealmSlot()
		log.WithField("possibleInvalidPairs", len(possibleInvalidPair)).
			WithField("latestVerifiedRealmSlot", latestSavedVerifiedRealmSlot).
			Info("Requeue possible invalid pairs")

		slotCounter := latestSavedVerifiedRealmSlot

		for _, pair := range possibleInvalidPair {
			if nil == pair {
				continue
			}

			if pair.Slot > latestSavedVerifiedRealmSlot {
				log.WithField("slot", pair.Slot).
					WithField("vanguardHash", pair.VanguardHash).
					WithField("pandoraHashes", pair.PandoraHashes).
					Warn("pair slot higher than latestSavedVerifiedRealmSlot")
				continue
			}

			vanguardErr = vanguardHashDB.SaveVanguardHeaderHash(pair.Slot, &types.HeaderHash{
				Status: types.Skipped,
			})

			// TODO: when more shard will come we will need to maintain this information
			pandoraErr = pandoraHeaderHashDB.SavePandoraHeaderHash(pair.Slot, &types.HeaderHash{
				Status: types.Skipped,
			})

			if nil != vanguardErr || nil != pandoraErr {
				log.WithField("vanguardErr", vanguardErr).
					WithField("pandoraErr", pandoraErr).
					WithField("realmErr", realmErr).
					Error("Got error during invalidation of pending queue")
				break
			}
		}

		realmErr = realmDB.SaveLatestVerifiedRealmSlot(slotCounter)

		log.WithField("highestCheckedSlot", slotCounter).
			Info("I have resolved InvalidatePendingQueue")

		return
	}

	return
}

// workLoop should be responsible of handling multiple events and resolving them
// Assumption is that if you want to validate pending queue you should receive information from Vanguard and Pandora
// TODO: handle reorgs
// TODO: consider working on MinimalConsensusInfo
func (service *Service) workLoop() {
	verifiedSlotWorkLoopStart := service.RealmDB.LatestVerifiedRealmSlot()
	log.WithField("verifiedSlotWorkLoopStart", verifiedSlotWorkLoopStart).
		Info("I am pushing work to canonicalizeChan")
	//service.canonicalizeChan <- verifiedSlotWorkLoopStart
	realmDB := service.RealmDB
	vanguardDB := service.VanguardHeaderHashDB
	pandoraDB := service.PandoraHeaderHashDB
	batchLimit := 50000
	possiblePendingWork := make([]*types.HeaderHash, 0)

	// This is arbitrary, it may be less or more. Depends on the approach
	debounceDuration := time.Second
	// Create merged channel
	mergedChannel := merge(service.VanguardHeadersChan, service.PandoraHeadersChan)

	// Create bridge for debounce
	mergedHeadersChanBridge := make(chan interface{})
	// Provide handlers for debounce
	mergedChannelHandler := func(work interface{}) {
		header, isHeaderHash := work.(*types.HeaderHash)

		if !isHeaderHash {
			log.WithField("cause", "mergedChannelHandler").Warn("invalid header hash")

			return
		}

		if nil == header {
			log.WithField("cause", "mergedChannelHandler").Warn("empty header hash")
			return
		}

		latestVerifiedRealmSlot := realmDB.LatestVerifiedRealmSlot()
		vanguardHashes, err := vanguardDB.VanguardHeaderHashes(latestVerifiedRealmSlot, uint64(batchLimit))

		if nil != err {
			log.WithField("cause", "mergedChannelHandler").Warn(err)

			return
		}

		pandoraHashes, err := pandoraDB.PandoraHeaderHashes(latestVerifiedRealmSlot, uint64(batchLimit))

		// This is naive, but might work
		// We need to have at least one pair to start invalidation.
		// It might lead to 2 pairs on one side, or invalidation stall,
		// But ATM I do not have quicker and better idea
		if len(possiblePendingWork) < 2 {
			log.WithField("cause", "mergedChannelHandler").Debug("not enough pending pairs")
			return
		}

		// If hash will be found above verified slot than it means that its pending
		// Otherwise we have an reorg (Not supported yet)
		// The most tricky part is to resolve what is the slot number for particular hash
		// TODO: consider pushing also slot number in channel or in headerHash struct
		for index, hash := range vanguardHashes {
			if nil == hash {
				continue
			}

			if hash.HeaderHash.String() != header.HeaderHash.String() {
				continue
			}

			currentSlot := uint64(index) + latestVerifiedRealmSlot
			service.canonicalizeChan <- currentSlot

			return
		}

		for index, hash := range pandoraHashes {
			if nil == hash {
				continue
			}

			if hash.HeaderHash.String() != header.HeaderHash.String() {
				continue
			}

			currentSlot := uint64(index) + latestVerifiedRealmSlot
			service.canonicalizeChan <- currentSlot

			return
		}
	}

	go func() {
		for {
			select {
			case header := <-mergedChannel:
				possiblePendingWork = append(possiblePendingWork, header)
				mergedHeadersChanBridge <- header
			case <-service.canonicalizeChan:
				possiblePendingWork = make([]*types.HeaderHash, 0)
			case stop := <-service.stopChan:
				if stop {
					log.WithField("canonicalize", "stop").Info("Received stop signal")

					return
				}
			}
		}
	}()

	// Debounce (aggregate) calls and invoke invalidation of pending queue only when needed
	go utils.Debounce(
		context.Background(),
		debounceDuration,
		mergedHeadersChanBridge,
		mergedChannelHandler,
	)
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
