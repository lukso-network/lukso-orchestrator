package api

import (
	"context"

	"github.com/ethereum/go-ethereum/event"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/cache"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/iface"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	log "github.com/sirupsen/logrus"
)

type Backend struct {
	// feed
	ConsensusInfoFeed iface.ConsensusInfoFeed

	// db reference
	ConsensusInfoDB    db.ROnlyConsensusInfoDB
	VerifiedSlotInfoDB db.ROnlyVerifiedSlotInfoDB
	InvalidSlotInfoDB  db.ROnlyInvalidSlotInfoDB

	// cache reference
	VanguardPendingShardingCache cache.VanguardShardCache
	PandoraPendingHeaderCache    cache.PandoraHeaderCache
}

func (backend *Backend) SubscribeNewEpochEvent(ch chan<- *types.MinimalEpochConsensusInfo) event.Subscription {
	return backend.ConsensusInfoFeed.SubscribeMinConsensusInfoEvent(ch)
}

func (backend *Backend) ConsensusInfoByEpochRange(fromEpoch uint64) []*types.MinimalEpochConsensusInfo {
	consensusInfos, err := backend.ConsensusInfoDB.ConsensusInfos(fromEpoch)
	if err != nil {
		return nil
	}
	return consensusInfos
}

// GetSlotStatus
func (backend *Backend) GetSlotStatus(ctx context.Context, slot uint64, requestType bool) types.Status {

	headerInfo, _ := backend.PandoraPendingHeaderCache.Get(ctx, slot)
	shardInfo, _ := backend.VanguardPendingShardingCache.Get(ctx, slot)

	// by default if nothing is found then return skipped
	status := types.Skipped

	// if it is found in the cache then it is pending
	// maybe already verifying is going on but not settled in the database
	if shardInfo != nil || headerInfo != nil {
		status = types.Pending
	}

	//when requested slot is greater than latest verified slot
	latestVerifiedSlot := backend.VerifiedSlotInfoDB.InMemoryLatestVerifiedSlot()
	if slot > latestVerifiedSlot {
		status = types.Unknown
	}

	// finally found in the database so return immediately so that no other db call happens
	if slotInfo, _ := backend.VerifiedSlotInfoDB.VerifiedSlotInfo(slot); slotInfo != nil {
		status = types.Verified
		return status
	}

	// finally found in the database so return immediately so that no other db call happens
	if slotInfo, _ := backend.InvalidSlotInfoDB.InvalidSlotInfo(slot); slotInfo != nil {
		status = types.Invalid
		return status
	}

	defer log.WithField("slot", slot).
		WithField("last verified slot", latestVerifiedSlot).
		WithField("api", "ConfirmPanBlockHashes").
		WithField("status", status).
		Debug("GotSlotStatus")

	return status
}
