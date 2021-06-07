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
)

// This part could be moved to other place during refactor, might be registered as a service
type Service struct {
	VanguardHeaderHashDB iface.VanHeaderAccessDatabase
	PandoraHeaderHashDB  iface.PanHeaderAccessDatabase
	RealmDB              iface.RealmAccessDatabase
	stopChan             chan bool
	canonicalizeChan     chan uint64
}

func (service *Service) Start() {
	go func() {
		for {
			select {
			case slot := <-service.canonicalizeChan:
				vanguardErr, pandoraErr, realmErr := service.Canonicalize(slot, 500)

				if nil != vanguardErr {
					log.WithField("canonicalize", "vanguardErr").Error(vanguardErr)
				}

				if nil != pandoraErr {
					log.WithField("canonicalize", "pandoraErr").Error(pandoraErr)
				}

				if nil != realmErr {
					log.WithField("canonicalize", "realmErr").Error(realmErr)
				}
			case stop := <-service.stopChan:
				if stop {
					log.WithField("canonicalize", "stop").Info("Received stop signal")
					return
				}
			}
		}
	}()

	latestVerifiedSlot := service.RealmDB.LatestVerifiedRealmSlot()
	service.canonicalizeChan <- latestVerifiedSlot

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

func New(ctx context.Context, database db.Database) (service *Service) {
	stopChan := make(chan bool)
	canonicalizeChain := make(chan uint64)

	return &Service{
		VanguardHeaderHashDB: database,
		PandoraHeaderHashDB:  database,
		RealmDB:              database,
		stopChan:             stopChan,
		canonicalizeChan:     canonicalizeChain,
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

	log.Info("I am starting to InvalidatePendingQueue in batches")

	// If higher slot was found and is valid all the gaps between must me treated as invalid and discarded
	// SIDE NOTE: This is invalid, when a lot of blocks were just simply not present yet due to the network traffic
	possibleInvalidPair := make([]*events.RealmPair, 0)
	latestSavedVerifiedRealmSlot := realmDB.LatestVerifiedRealmSlot()

	if fromSlot > latestSavedVerifiedRealmSlot {
		realmErr = fmt.Errorf("I cannot start invalidation without root")

		return
	}

	log.WithField("latestSavedVerifiedRealmSlot", latestSavedVerifiedRealmSlot).
		WithField("from slot", fromSlot).
		Info("Invalidation starts")

	pandoraHeaderHashes, err := pandoraHeaderHashDB.PandoraHeaderHashes(fromSlot)

	if nil != err {
		log.WithField("cause", "Failed to invalidate pending queue").Error(err)
		return
	}

	vanguardBlockHashes, err := vanguardHashDB.VanguardHeaderHashes(fromSlot, batchLimit)

	log.WithField("pandoraHeaderHashes", len(pandoraHeaderHashes)).
		WithField("vanguardHeaderHashes", len(vanguardBlockHashes)).
		Trace("Got header hashes")

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

		slotCounter = realmDB.LatestVerifiedRealmSlot()

		if slotCounter > pair.Slot {
			continue
		}

		slotCounter = pair.Slot
	}

	realmErr = realmDB.SaveLatestVerifiedRealmSlot(slotCounter)

	log.WithField("highestCheckedSlot", slotCounter).
		Info("I have resolved InvalidatePendingQueue")

	return
}
