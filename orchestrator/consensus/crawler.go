package consensus

// crawler.go
// Motivation of this file is to have functional approach rather than object oriented to keep logic unaggregated
// Treat it more like gateway to commandBus or CQRS

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db/kv"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/rpc/api/events"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	log "github.com/sirupsen/logrus"
	"sync"
)

type filteredVerifiedPairs struct {
	validPairs           []*events.RealmPair
	possibleSkippedPairs []*events.RealmPair
	pandoraOrphans       map[uint64]*types.HeaderHash
	vanguardOrphans      map[uint64]*types.HeaderHash
	highestVerifiedSlot  uint64
}

// handlePreparedWork should be synchronous approach
func handlePreparedWork(
	vanguardHashDB db.VanguardHeaderHashDB,
	pandoraHeaderHashDB db.PandoraHeaderHashDB,
	realmDB db.RealmDB,
	invalidationWorkPayload *invalidationWorkPayload,
	errChan chan databaseErrors,
) {
	select {
	case <-errChan:
		return
	default:
		//vanguardBlockHashes := invalidationWorkPayload.vanguardHashes
		//pandoraHeaderHashes := invalidationWorkPayload.pandoraHashes
		//invalidationStartRealmSlot := invalidationWorkPayload.invalidationStartRealmSlot
		//fromSlot := invalidationWorkPayload.fromSlot
		//possibleSkippedPair := invalidationWorkPayload.possibleSkippedPair
		//vanguardOrphans := invalidationWorkPayload.vanguardOrphans
		//pandoraOrphans := invalidationWorkPayload.pandoraOrphans
		filteredVerifiedPairsByVanguard := resolveVerifiedPairsBasedOnVanguard(invalidationWorkPayload)

		//succeed := savePairsState(
		//	filteredVerifiedPairs.validPairs,
		//	types.Verified, vanguardHashDB,
		//	pandoraHeaderHashDB,
		//	realmDB,
		//	errChan,
		//)

		skippedPairs, pendingPairs := resolvePendingAndSkipped(possibleSkippedPairs)

		waitGroup := sync.WaitGroup{}
		waitGroup.Add(3)

		go func() {
			succeed := savePairsState(validPairs, types.Verified, vanguardHashDB, pandoraHeaderHashDB, realmDB, errChan)

			if succeed {
				waitGroup.Done()
			}
		}()

		// Save pandora orphans
		// Save vanguard orphans
		// Save invalidPairs

		go func() {
			succeed := savePairsState(validPairs, types.Skipped, vanguardHashDB, pandoraHeaderHashDB, realmDB, errChan)

			if succeed {
				waitGroup.Done()
			}
		}()

	}

	var (
		vanguardErr error
		pandoraErr  error
		realmErr    error
	)

	validPairs, possibleSkippedPairs := resolveVerifiedPairsBasedOnVanguard(invalidationWorkPayload)

	if !succeeded {
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
			possibleSkippedPair = append(possibleSkippedPair, &events.RealmPair{
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
			vanguardOrphans[slotToCheck] = vanguardHeaderHash

			continue
		}

		if nil == vanguardBlockHash {
			currentPandoraHeaderHash := &types.HeaderHash{
				HeaderHash: pandoraHeaderHash.HeaderHash,
				Status:     types.Pending,
			}
			pandoraOrphans[slotToCheck] = currentPandoraHeaderHash

			continue
		}

		log.WithField("slot", slotToCheck).
			WithField("hash", vanguardBlockHash.HeaderHash.String()).
			Debug("I am inserting verified vanguardBlockHash")

		vanguardErr = vanguardHashDB.SaveVanguardHeaderHash(slotToCheck, &types.HeaderHash{
			HeaderHash: vanguardBlockHash.HeaderHash,
			Status:     types.Verified,
		})

		log.WithField("slot", slotToCheck).
			WithField("hash", pandoraHeaderHash.HeaderHash.String()).
			Debug("I am inserting verified pandoraHeaderHash")

		pandoraErr = pandoraHeaderHashDB.SavePandoraHeaderHash(slotToCheck, &types.HeaderHash{
			HeaderHash: pandoraHeaderHash.HeaderHash,
			Status:     types.Verified,
		})

		if nil != vanguardErr || nil != pandoraErr {
			break
		}

		realmErr = realmDB.SaveLatestVerifiedRealmSlot(slotToCheck)
	}

	if nil != vanguardErr || nil != pandoraErr || nil != realmErr {
		log.WithField("vanguardErr", vanguardErr).
			WithField("pandoraErr", pandoraErr).
			WithField("realmErr", realmErr).
			Error("Got error during invalidation of pending queue")
		errChan <- databaseErrors{
			vanguardErr: vanguardErr,
			pandoraErr:  pandoraErr,
			realmErr:    realmErr,
		}

		return
	}

	// Resolve state of possible invalid pairs
	latestSavedVerifiedRealmSlot := realmDB.LatestVerifiedRealmSlot()
	log.WithField("possibleInvalidPairs", len(possibleSkippedPair)).
		WithField("latestVerifiedRealmSlot", latestSavedVerifiedRealmSlot).
		WithField("invalidationStartRealmSlot", invalidationStartRealmSlot).
		Info("Requeue possible invalid pairs")

	invalidationRange := latestSavedVerifiedRealmSlot - invalidationStartRealmSlot

	// All of orphans and possibleSkipped are still pending
	if 0 == invalidationRange {
		log.WithField("invalidationStartRealmSlot", invalidationStartRealmSlot).
			WithField("latestVerifiedRealmSlot", latestSavedVerifiedRealmSlot).
			Warn("I did not progress any slot")

		return
	}

	if invalidationRange < 0 {
		log.Fatal("Got wrong invalidation range. This is a fatal bug that should never happen.")

		return
	}

	// Mark all pandora Orphans as skipped
	for slot := range pandoraOrphans {
		pandoraErr = pandoraHeaderHashDB.SavePandoraHeaderHash(slot, &types.HeaderHash{
			HeaderHash: common.Hash{},
			Status:     types.Skipped,
		})
	}

	// Mark all vanguard orphans as skipped
	for slot := range vanguardOrphans {
		vanguardErr = vanguardHashDB.SaveVanguardHeaderHash(slot, &types.HeaderHash{
			HeaderHash: common.Hash{},
			Status:     types.Skipped,
		})
	}

	if nil != vanguardErr || nil != pandoraErr {
		log.WithField("vanguardErr", vanguardErr).
			WithField("pandoraErr", pandoraErr).
			Error("Got error during invalidation of possible orphans")
		errChan <- databaseErrors{
			vanguardErr: vanguardErr,
			pandoraErr:  pandoraErr,
			realmErr:    realmErr,
		}

		return
	}

	pendingPairs := make([]*events.RealmPair, 0)

	for _, pair := range possibleSkippedPair {
		if nil == pair {
			continue
		}

		if pair.Slot > latestSavedVerifiedRealmSlot {
			pendingPairs = append(pendingPairs, pair)

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

	if nil != vanguardErr || nil != pandoraErr {
		log.WithField("vanguardErr", vanguardErr).
			WithField("pandoraErr", pandoraErr).
			Error("Got error during invalidation of possible skipped pairs")
		errChan <- databaseErrors{
			vanguardErr: vanguardErr,
			pandoraErr:  pandoraErr,
			realmErr:    realmErr,
		}

		return
	}

	for _, pair := range pendingPairs {
		if nil != pair.VanguardHash {
			vanguardErr = vanguardHashDB.SaveVanguardHeaderHash(pair.Slot, &types.HeaderHash{
				Status:     types.Skipped,
				HeaderHash: pair.VanguardHash.HeaderHash,
			})
		}

		// TODO: when more shard will come we will need to maintain this information
		if len(pair.PandoraHashes) > 0 && nil != pair.PandoraHashes[0] {
			pandoraErr = pandoraHeaderHashDB.SavePandoraHeaderHash(pair.Slot, &types.HeaderHash{
				Status:     types.Skipped,
				HeaderHash: pair.PandoraHashes[0].HeaderHash,
			})
		}
	}

	if nil != vanguardErr || nil != pandoraErr {
		log.WithField("vanguardErr", vanguardErr).
			WithField("pandoraErr", pandoraErr).
			Error("Got error during invalidation of pendingPairs")
		errChan <- databaseErrors{
			vanguardErr: vanguardErr,
			pandoraErr:  pandoraErr,
			realmErr:    realmErr,
		}

		return
	}

	var (
		finalVanguardBatch []*types.HeaderHash
		finalPandoraBatch  []*types.HeaderHash
	)

	// At the very end fill all vanguard and pandora nil entries as skipped ones
	// Do not fetch any higher records
	finalVanguardBatch, vanguardErr = vanguardHashDB.VanguardHeaderHashes(
		invalidationStartRealmSlot,
		invalidationRange,
	)

	finalPandoraBatch, pandoraErr = pandoraHeaderHashDB.PandoraHeaderHashes(
		invalidationStartRealmSlot,
		invalidationRange,
	)

	if nil != vanguardErr || nil != pandoraErr {
		log.WithField("vanguardErr", vanguardErr).
			WithField("pandoraErr", pandoraErr).
			WithField("realmErr", realmErr).
			Error("Got error during invalidation of pending queue")

		errChan <- databaseErrors{
			vanguardErr: vanguardErr,
			pandoraErr:  pandoraErr,
			realmErr:    realmErr,
		}

		return
	}

	for index, headerHash := range finalVanguardBatch {
		if nil != headerHash {
			continue
		}

		slotToCheck := fromSlot + uint64(index)

		if slotToCheck > latestSavedVerifiedRealmSlot {
			continue
		}

		headerHash = &types.HeaderHash{
			HeaderHash: kv.EmptyHash,
			Status:     types.Skipped,
		}

		vanguardErr = vanguardHashDB.SaveVanguardHeaderHash(slotToCheck, headerHash)
	}

	for index, headerHash := range finalPandoraBatch {
		if nil != headerHash {
			continue
		}

		slotToCheck := fromSlot + uint64(index)

		if slotToCheck > latestSavedVerifiedRealmSlot {
			continue
		}

		headerHash = &types.HeaderHash{
			HeaderHash: kv.EmptyHash,
			Status:     types.Skipped,
		}
		pandoraErr = pandoraHeaderHashDB.SavePandoraHeaderHash(slotToCheck, headerHash)
	}

	if nil != vanguardErr || nil != pandoraErr {
		log.WithField("vanguardErr", vanguardErr).
			WithField("pandoraErr", pandoraErr).
			WithField("realmErr", realmErr).
			Error("Got error during invalidation of final Vanguard or Pandora batch")

		errChan <- databaseErrors{
			vanguardErr: vanguardErr,
			pandoraErr:  pandoraErr,
			realmErr:    realmErr,
		}

		return
	}

	log.WithField("highestCheckedSlot", latestSavedVerifiedRealmSlot).
		Info("I have resolved Canonicalize")
}

// resolveVerifiedPairsBasedOnVanguard
// First of all we get verified, then we get possibleSkipped
// We keep track of the orphans
// After this loop we sort possible skipped
func resolveVerifiedPairsBasedOnVanguard(
	invalidationWorkPayloadStruct *invalidationWorkPayload,
) (
	filtered *filteredVerifiedPairs,
) {
	vanguardBlockHashes := invalidationWorkPayloadStruct.vanguardHashes
	pandoraHeaderHashes := invalidationWorkPayloadStruct.pandoraHashes
	fromSlot := invalidationWorkPayloadStruct.fromSlot
	vanguardOrphans := map[uint64]*types.HeaderHash{}
	pandoraOrphans := map[uint64]*types.HeaderHash{}
	validPairs := make([]*events.RealmPair, 0)
	possibleSkippedPairs := make([]*events.RealmPair, 0)
	highestVerifiedSlot := uint64(0)

	// This is quite naive, but should work
	for index, vanguardBlockHash := range vanguardBlockHashes {
		slotToCheck := fromSlot + uint64(index)

		if len(pandoraHeaderHashes) <= index {
			break
		}

		pandoraHeaderHash := pandoraHeaderHashes[index]

		// Potentially skipped slot
		if nil == pandoraHeaderHash && nil == vanguardBlockHash {
			possibleSkippedPairs = append(possibleSkippedPairs, &events.RealmPair{
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
			vanguardOrphans[slotToCheck] = vanguardHeaderHash

			continue
		}

		if nil == vanguardBlockHash {
			currentPandoraHeaderHash := &types.HeaderHash{
				HeaderHash: pandoraHeaderHash.HeaderHash,
				Status:     types.Pending,
			}
			pandoraOrphans[slotToCheck] = currentPandoraHeaderHash

			continue
		}

		// Here lays multiple shards handling logic
		pandoraHashes := make([]*types.HeaderHash, 0)
		pandoraHashes = append(pandoraHashes, &types.HeaderHash{
			HeaderHash: pandoraHeaderHash.HeaderHash,
			Status:     types.Verified,
		})

		validPairs = append(validPairs, &events.RealmPair{
			Slot: slotToCheck,
			VanguardHash: &types.HeaderHash{
				HeaderHash: vanguardBlockHash.HeaderHash,
				Status:     types.Verified,
			},
			PandoraHashes: pandoraHashes,
		})

		// Keep in-memory track of highest verified pair
		if highestVerifiedSlot < slotToCheck {
			highestVerifiedSlot = slotToCheck
		}

		log.WithField("slot", slotToCheck).
			WithField("hash", vanguardBlockHash.HeaderHash.String()).
			Debug("I am inserting verified vanguardBlockHash")
	}

	filtered = &filteredVerifiedPairs{
		validPairs:           validPairs,
		possibleSkippedPairs: possibleSkippedPairs,
		pandoraOrphans:       pandoraOrphans,
		vanguardOrphans:      vanguardOrphans,
		highestVerifiedSlot:  highestVerifiedSlot,
	}

	return
}

func resolvePendingAndSkipped(
	possibleSkippedPair []*events.RealmPair,
	latestSavedVerifiedRealmSlot uint64,
) (
	skippedPairs []*events.RealmPair,
	pendingPairs []*events.RealmPair,
) {
	skippedPairs = make([]*events.RealmPair, 0)
	pendingPairs = make([]*events.RealmPair, 0)

	for _, pair := range possibleSkippedPair {
		if nil == pair {
			continue
		}

		if pair.Slot > latestSavedVerifiedRealmSlot {
			pendingPairs = append(pendingPairs, pair)

			continue
		}

		skippedPairs = append(skippedPairs, pair)
	}

	return
}

func savePairsState(
	pairs []*events.RealmPair,
	status types.Status,
	vanguardHashDB db.VanguardHeaderHashDB,
	pandoraHeaderHashDB db.PandoraHeaderHashDB,
	realmDB db.RealmDB,
	errChan chan databaseErrors,
) (success bool) {
	var (
		vanguardErr error
		pandoraErr  error
		realmErr    error
	)

	for _, pair := range pairs {
		slotToCheck := pair.Slot
		vanguardBlockHash := pair.VanguardHash
		pandoraHashes := pair.PandoraHashes

		log.WithField("slot", slotToCheck).
			WithField("hash", vanguardBlockHash.HeaderHash.String()).
			Debug("I am inserting verified vanguardBlockHash")

		vanguardErr = vanguardHashDB.SaveVanguardHeaderHash(slotToCheck, &types.HeaderHash{
			HeaderHash: vanguardBlockHash.HeaderHash,
			Status:     status,
		})

		// TODO: handle multiple shards
		for _, pandoraHeaderHash := range pandoraHashes {
			log.WithField("slot", slotToCheck).
				WithField("hash", pandoraHeaderHash.HeaderHash.String()).
				Debug("I am inserting verified pandoraHeaderHash")

			pandoraErr = pandoraHeaderHashDB.SavePandoraHeaderHash(slotToCheck, &types.HeaderHash{
				HeaderHash: pandoraHeaderHash.HeaderHash,
				Status:     status,
			})

			if nil != pandoraErr {
				break
			}
		}

		if nil != vanguardErr || nil != pandoraErr {
			break
		}

		realmErr = realmDB.SaveLatestVerifiedRealmSlot(slotToCheck)

		if nil != realmErr {
			break
		}
	}

	if nil != vanguardErr || nil != pandoraErr || nil != realmErr {
		errChan <- databaseErrors{
			vanguardErr: vanguardErr,
			pandoraErr:  pandoraErr,
			realmErr:    realmErr,
		}

		success = false

		return
	}

	success = true

	return
}
