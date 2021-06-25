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
		filteredVerifiedPairsByVanguard := resolveVerifiedPairsBasedOnVanguard(invalidationWorkPayload)
		validPairs := filteredVerifiedPairsByVanguard.validPairs
		highestSlot := filteredVerifiedPairsByVanguard.highestVerifiedSlot
		invalidationStartRealmSlot := invalidationWorkPayload.invalidationStartRealmSlot
		invalidationRange := highestSlot - invalidationStartRealmSlot

		if 0 == invalidationRange {
			log.WithField("invalidationStartRealmSlot", invalidationStartRealmSlot).
				WithField("highestSlot", highestSlot).
				Warn("I did not progress any slot")

			return
		}

		if invalidationRange < 0 {
			log.Fatal("Got wrong invalidation range. This is a fatal bug that should never happen.")

			return
		}

		incompleteSkipped, pendingPairs := resolvePendingAndSkippedByHighestSlot(filteredVerifiedPairsByVanguard)
		// Resolve completeness of skipped by orphans
		skippedPairs := resolveSkippedByOrphans(&filteredVerifiedPairs{
			validPairs:           filteredVerifiedPairsByVanguard.validPairs,
			possibleSkippedPairs: incompleteSkipped,
			pandoraOrphans:       filteredVerifiedPairsByVanguard.pandoraOrphans,
			vanguardOrphans:      filteredVerifiedPairsByVanguard.vanguardOrphans,
			highestVerifiedSlot:  highestSlot,
		})

		waitGroup := &sync.WaitGroup{}
		waitGroup.Add(3)
		savePairState := func(
			pairs []*events.RealmPair,
			group *sync.WaitGroup,
		) {
			succeed := savePairsState(validPairs, vanguardHashDB, pandoraHeaderHashDB, realmDB, errChan)

			if succeed {
				waitGroup.Done()
			}
		}

		bulkAsyncSave := func() {
			go savePairState(validPairs, waitGroup)
			go savePairState(skippedPairs, waitGroup)
			go savePairState(pendingPairs, waitGroup)
		}

		go bulkAsyncSave()
		waitGroup.Wait()

		// Handle all possible nils in this batches
		finalVanguardBatch, finalPandoraBatch := fetchFinalVanguardAndPandoraBatch(
			vanguardHashDB,
			pandoraHeaderHashDB,
			invalidationStartRealmSlot,
			invalidationRange,
			errChan,
		)

		skippedSlots := resolveSkippedBasedOnNilRecordsInBatch(
			finalVanguardBatch,
			finalPandoraBatch,
			invalidationStartRealmSlot,
			highestSlot,
		)

		saved := savePairsState(skippedSlots, vanguardHashDB, pandoraHeaderHashDB, realmDB, errChan)

		if !saved {
			log.WithField("cause", "batch based on nil records").
				Error("failure during save of batch based on nil records")

			return
		}

		log.WithField("highestCheckedSlot", highestSlot).
			Info("I have resolved Canonicalize")

		saveLatestVerifiedSlot(realmDB, highestSlot, errChan)
	}
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

func resolvePendingAndSkippedByHighestSlot(
	filtered *filteredVerifiedPairs,
) (
	skippedPairs []*events.RealmPair,
	pendingPairs []*events.RealmPair,
) {
	possibleSkippedPair := filtered.possibleSkippedPairs
	latestSavedVerifiedRealmSlot := filtered.highestVerifiedSlot

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

func resolveSkippedByOrphans(filtered *filteredVerifiedPairs) (skipped []*events.RealmPair) {
	fullSkipped := map[uint64]*events.RealmPair{}
	skipped = make([]*events.RealmPair, 0)

	for _, pair := range filtered.possibleSkippedPairs {
		fullSkipped[pair.Slot] = pair
	}

	insertSkippedSlot := func(slot uint64) {
		if slot <= filtered.highestVerifiedSlot {
			fullSkipped[slot] = &events.RealmPair{
				Slot: slot,
				VanguardHash: &types.HeaderHash{
					HeaderHash: common.Hash{},
					Status:     types.Skipped,
				},
				PandoraHashes: make([]*types.HeaderHash, 1),
			}

			fullSkipped[slot].PandoraHashes[0] = &types.HeaderHash{
				HeaderHash: common.Hash{},
				Status:     types.Skipped,
			}
		}
	}

	// Override map. This will be way complicated when sharding comes
	for slot := range filtered.pandoraOrphans {
		insertSkippedSlot(slot)
	}

	// Mark all vanguard orphans as skipped
	for slot := range filtered.vanguardOrphans {
		insertSkippedSlot(slot)
	}

	// For now it doesn't matter is it sorted ascending or random
	for _, pair := range fullSkipped {
		skipped = append(skipped, pair)
	}

	return
}

func fetchFinalVanguardAndPandoraBatch(
	vanguardHashDB db.VanguardHeaderHashDB,
	pandoraHeaderHashDB db.PandoraHeaderHashDB,
	invalidationStartRealmSlot uint64,
	invalidationRange uint64,
	errChan chan databaseErrors,
) (finalVanguardBatch []*types.HeaderHash, finalPandoraBatch []*types.HeaderHash) {
	finalVanguardBatch, vanguardErr := vanguardHashDB.VanguardHeaderHashes(
		invalidationStartRealmSlot,
		invalidationRange,
	)

	finalPandoraBatch, pandoraErr := pandoraHeaderHashDB.PandoraHeaderHashes(
		invalidationStartRealmSlot,
		invalidationRange,
	)

	if nil != vanguardErr || nil != pandoraErr {
		errChan <- databaseErrors{
			vanguardErr: vanguardErr,
			pandoraErr:  pandoraErr,
			realmErr:    nil,
		}

		finalVanguardBatch = nil
		finalPandoraBatch = nil
	}

	return
}

func resolveSkippedBasedOnNilRecordsInBatch(
	finalVanguardBatch []*types.HeaderHash,
	finalPandoraBatch []*types.HeaderHash,
	fromSlot uint64,
	latestSavedVerifiedRealmSlot uint64,
) (skippedPairs []*events.RealmPair) {
	skippedSlots := map[uint64]*events.RealmPair{}
	skippedPairs = make([]*events.RealmPair, 0)

	insertSkippedSlot := func(slotToCheck uint64) {
		headerHash := &types.HeaderHash{
			HeaderHash: kv.EmptyHash,
			Status:     types.Skipped,
		}

		skippedSlots[slotToCheck] = &events.RealmPair{
			Slot:          slotToCheck,
			VanguardHash:  headerHash,
			PandoraHashes: make([]*types.HeaderHash, 1),
		}

		skippedSlots[slotToCheck].PandoraHashes[0] = headerHash
	}

	for index, headerHash := range finalVanguardBatch {
		if nil != headerHash {
			continue
		}

		slotToCheck := fromSlot + uint64(index)

		if slotToCheck > latestSavedVerifiedRealmSlot {
			continue
		}

		insertSkippedSlot(slotToCheck)
	}

	for index, headerHash := range finalPandoraBatch {
		if nil != headerHash {
			continue
		}

		slotToCheck := fromSlot + uint64(index)

		if slotToCheck > latestSavedVerifiedRealmSlot {
			continue
		}

		insertSkippedSlot(slotToCheck)
	}

	for _, skippedSlot := range skippedSlots {
		skippedPairs = append(skippedPairs, skippedSlot)
	}

	return
}

func savePairsState(
	pairs []*events.RealmPair,
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
			Status:     vanguardBlockHash.Status,
		})

		// TODO: handle multiple shards
		for _, pandoraHeaderHash := range pandoraHashes {
			log.WithField("slot", slotToCheck).
				WithField("hash", pandoraHeaderHash.HeaderHash.String()).
				Debug("I am inserting verified pandoraHeaderHash")

			pandoraErr = pandoraHeaderHashDB.SavePandoraHeaderHash(slotToCheck, &types.HeaderHash{
				HeaderHash: pandoraHeaderHash.HeaderHash,
				Status:     pandoraHeaderHash.Status,
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

func saveLatestVerifiedSlot(realmDB db.RealmDB, slot uint64, errChan chan databaseErrors) {
	err := realmDB.SaveLatestVerifiedRealmSlot(slot)

	if nil != err {
		errChan <- databaseErrors{realmErr: err}
	}
}
