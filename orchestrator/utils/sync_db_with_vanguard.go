package utils

import (
	"context"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db/kv"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/client"
	"github.com/pkg/errors"
	types "github.com/prysmaticlabs/eth2-types"
	log "github.com/sirupsen/logrus"
	"time"
)

const vanguardConnectionRetryLimit = 10

func EstablishConnectionWithVanguard(ctx context.Context, endpoint string) (vanClient client.VanguardClient, err error) {
	ticker := time.NewTicker(5 * time.Second)
	retryLimit := vanguardConnectionRetryLimit
	for retryLimit > 0 {
		select {
		case <- ctx.Done():
			return nil, errors.New("context deadline exited. Closing...")
		case <- ticker.C:
			vanClient, err = client.Dial(ctx, endpoint, time.Minute*6, 32, math.MaxInt32)
			if err == nil {
				// connected. so return immediately
				return vanClient, err
			}
		}
		retryLimit--
	}
	return nil, errors.New("connection establishment failed with vanguard")
}

func SyncDatabase(ctx context.Context, localDB db.Database, vanguardEndpoint string) error {

	if localDB.LatestSavedEpoch() == 0 && localDB.LatestSavedVerifiedSlot() == 0 {
		// the instance is just started freshly. No need to sync the database.
		log.Info("fresh start. No need to sync database")
		return nil
	}

	max := func(a, b uint64) uint64{
		if a > b {
			return a
		}
		return b
	}

	// some data is present in the orchestrator databse. So these data should be synced with vanguard.
	// first connect to vanguard
	vanClient, err := EstablishConnectionWithVanguard(ctx, vanguardEndpoint)
	if err != nil {
		// connection establishment failed. so dB can't be synced.
		return err
	}
	// while exiting sync, just close the connection
	defer vanClient.Close()

	// fetch latest finalized epoch
	finalizedEpoch, err := vanClient.GetFinalizedEpoch()
	if err != nil {
		return err
	}
	latestEpochInOrchestrator := localDB.LatestSavedEpoch()

	log.WithField("finalizedEpoch", finalizedEpoch).WithField("local latest Epoch", latestEpochInOrchestrator).Info("SyncDatabase fetched info")

	if uint64(finalizedEpoch) < latestEpochInOrchestrator {
		// finalized epoch is lower than our saved epoch.
		// so maybe orchestrator is in the wrong chain and as soon as vanguard syncs orchestrator will hold invalid info
		// so just remove everything from finalized epoch to our latest epoch
		// we will subscribe from or before finalize epoch so info won't be reverted.
		err := localDB.RemoveInfoFromAllDb(uint64(finalizedEpoch), latestEpochInOrchestrator)
		if err != nil {
			return err
		}
	}
	latestEpochInOrchestrator = localDB.LatestSavedEpoch()
	validationIterator := latestEpochInOrchestrator
	for validationIterator > 0 {
		retrievedSlotInfo, err := localDB.GetFirstVerifiedSlotInAnEpoch(validationIterator)
		if err != nil {
			log.WithField("error", err).WithField("epoch", validationIterator).Error("First verified slot info not found")
		}
		if retrievedSlotInfo != nil {
			inCanonicalChain, err := vanClient.IsValidBlock(0, retrievedSlotInfo.VanguardBlockHash.Bytes())
			if err != nil {
				return err
			}
			if inCanonicalChain {
				// hash is in canonical chain. so from here it is a valid epoch
				break
			}
		}
		log.WithField("epoch", validationIterator).Debug("now validating")
		validationIterator--
	}

	// delete everything till finalized epoch or maximum invalid epoch
	if validationIterator < max(latestEpochInOrchestrator, uint64(finalizedEpoch)) {
		log.WithField("from epoch", validationIterator).WithField("to epoch", latestEpochInOrchestrator).Debug("removing info from all dB")
		err := localDB.RemoveInfoFromAllDb(validationIterator, max(latestEpochInOrchestrator, uint64(finalizedEpoch)))
		if err != nil {
			return err
		}
	}
	return nil
}

func PreviousEpochReceived (vanguardClient client.VanguardClient, epoch uint64, localDb db.Database) error {
	startSlot := kv.StartSlot(epoch)
	infos, err := localDb.VerifiedSlotInfos(startSlot)
	if err != nil {
		return err
	}
	for slot, info := range infos {
		isInCanonical, err := vanguardClient.IsValidBlock(types.Slot(slot), info.VanguardBlockHash.Bytes())
		if err != nil {
			return err
		}
		if !isInCanonical {
			log.WithField("slot", slot).WithField("hash", info.VanguardBlockHash).Debug("slot is not in canonical chain")
			err := localDb.RemoveSlotInfo(slot)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
