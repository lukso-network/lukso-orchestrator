package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/event"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/cache"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/iface"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

var ErrHeaderHashMisMatch = errors.New("Header hash mis-matched")

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
func (backend *Backend) GetSlotStatus(ctx context.Context, slot uint64, hash common.Hash, requestFrom bool) types.Status {
	// by default if nothing is found then return skipped
	status := types.Pending

	//when requested slot is greater than latest verified slot
	latestVerifiedSlot := backend.VerifiedSlotInfoDB.InMemoryLatestVerifiedSlot()
	var slotInfo *types.SlotInfo

	logPrinter := func(stat types.Status) {
		log.WithField("slot", slot).
			WithField("latestVerifiedSlot", latestVerifiedSlot).
			WithField("slotInfo", fmt.Sprintf("%+v", slotInfo)).
			WithField("status", stat).
			Debug("Verification status")
	}
	// finally found in the database so return immediately so that no other db call happens
	if slotInfo, _ = backend.VerifiedSlotInfoDB.VerifiedSlotInfo(slot); slotInfo != nil {
		panHeaderHash := slotInfo.PandoraHeaderHash
		vanHeaderHash := slotInfo.VanguardBlockHash

		if requestFrom && panHeaderHash != hash {
			log.WithError(ErrHeaderHashMisMatch).
				Warn("Failed to match header hash with requested header hash from pandora node")
			logPrinter(types.Invalid)
			return types.Invalid
		}

		if !requestFrom && vanHeaderHash != hash {
			log.WithError(ErrHeaderHashMisMatch).
				Warn("Failed to match header hash with requested header hash from vanguard node")
			logPrinter(types.Invalid)
			return types.Invalid
		}

		status = types.Verified
		logPrinter(types.Verified)
		return status
	}

	// finally found in the database so return immediately so that no other db call happens
	if slotInfo, _ = backend.InvalidSlotInfoDB.InvalidSlotInfo(slot); slotInfo != nil {
		status = types.Invalid
		logPrinter(types.Invalid)
		return status
	}
	logPrinter(status)
	return status
}
