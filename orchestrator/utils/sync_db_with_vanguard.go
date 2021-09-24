package utils

import (
	"context"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/client"
	log "github.com/sirupsen/logrus"
	"time"
)

const vanguardConnectionRetryLimit = 10

func EstablishConnectionWithVanguard(ctx context.Context, endpoint string) (vanClient client.VanguardClient, err error) {
	ticker := time.NewTicker(5 * time.Second)
	retryLimit := vanguardConnectionRetryLimit
	for retryLimit > 0 {
		select {
		case <- ticker.C:
			vanClient, err = client.Dial(ctx, endpoint, time.Minute*6, 32, math.MaxInt32)
			if err == nil {
				// connected. so return immediately
				return vanClient, err
			}
		}
		retryLimit--
	}
	return
}

func SyncDatabase(ctx context.Context, localDB db.Database, vanguardEndpoint string) error {

	if localDB.LatestSavedEpoch() == 0 && localDB.LatestSavedVerifiedSlot() == 0 {
		// the instance is just started freshly. No need to sync the database.
		return nil
	}

	// some data is present in the orchestrator databse. So these data should be synced with vanguard.
	// first connect to vanguard
	vanClient, err := EstablishConnectionWithVanguard(ctx, vanguardEndpoint)
	if err != nil {
		// connection establishment failed. so dB can't be synced.
		return err
	}

	// fetch latest finalized epoch
	finalizedEpoch := vanClient.GetFinalizedEpoch()
	latestEpochInOrchestrator := localDB.LatestSavedEpoch()

	log.WithField("finalizedEpoch", finalizedEpoch).WithField("local latest Epoch", latestEpochInOrchestrator).Info("SyncDatabase fetched info")

	if uint64(finalizedEpoch) < latestEpochInOrchestrator {
		// finalized epoch is lower than our saved epoch.
		// so maybe orchestrator is in the wrong chain and as soon as vanguard syncs orchestrator will hold invalid info
		// so just remove everything from finalized epoch to our latest epoch
		// we will subscribe from or before finalize epoch so info won't be reverted.

	}
	return nil
}
