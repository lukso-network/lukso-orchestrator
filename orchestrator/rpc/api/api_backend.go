package api

import (
	"context"
	"github.com/ethereum/go-ethereum/event"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/cache"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/rpc/api/events"
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
func (b *Backend) GetSlotStatus(ctx context.Context, slot uint64, requestType bool) events.Status {
	// when requested slot is greater than latest verified slot
	latestVerifiedSlot := b.VerifiedSlotInfoDB.InMemoryLatestVerifiedSlot()
	if slot > latestVerifiedSlot {
		log.WithField("slot", slot).WithField(
			"latestVerifiedSlot", latestVerifiedSlot).Debug("Requested slot is unknown")
		return events.Unknown
	}

	if requestType {
		if headerInfo, _ := b.PandoraPendingHeaderCache.Get(ctx, slot); headerInfo != nil {
			log.WithField("slot", slot).WithField(
				"api", "ConfirmPanBlockHashes").Debug("Requested slot is pending")
			return events.Pending
		}
	} else {
		if shardInfo, _ := b.VanguardPendingShardingCache.Get(ctx, slot); shardInfo != nil {
			log.WithField("slot", slot).WithField(
				"api", "ConfirmVanBlockHashes").Debug("Requested slot is pending")
			return events.Pending
		}
	}

	if slotInfo, _ := b.VerifiedSlotInfoDB.VerifiedSlotInfo(slot); slotInfo != nil {
		log.WithField("slot", slot).WithField(
			"api", "ConfirmPanBlockHashes").Debug("Requested slot is verified")
		return events.Verified
	}

	if slotInfo, _ := b.InvalidSlotInfoDB.InvalidSlotInfo(slot); slotInfo != nil {
		log.WithField("slot", slot).WithField(
			"api", "ConfirmPanBlockHashes").Debug("Requested slot is invalid")
		return events.Invalid
	}

	return events.Skipped
}
