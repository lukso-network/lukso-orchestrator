package api

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/event"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/cache"
	conIface "github.com/lukso-network/lukso-orchestrator/orchestrator/consensus/iface"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/iface"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

var ErrHeaderHashMisMatch = errors.New("header hash mismatched")

type Backend struct {
	// feed
	ConsensusInfoFeed    iface.ConsensusInfoFeed
	VerifiedSlotInfoFeed conIface.VerifiedSlotInfoFeed

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

func (backend *Backend) SubscribeNewVerifiedSlotInfoEvent(ch chan<- *types.SlotInfoWithStatus) event.Subscription {
	return backend.VerifiedSlotInfoFeed.SubscribeVerifiedSlotInfoEvent(ch)
}

func (backend *Backend) ConsensusInfoByEpochRange(fromEpoch uint64) []*types.MinimalEpochConsensusInfo {
	consensusInfosV2, err := backend.ConsensusInfoDB.ConsensusInfos(fromEpoch)
	if err != nil {
		return nil
	}

	epochInfos := make([]*types.MinimalEpochConsensusInfo, len(consensusInfosV2))
	for i, epochInfo := range consensusInfosV2 {
		epochInfoV1 := epochInfo.ConvertToEpochInfoV1()
		epochInfos[i] = epochInfoV1
	}
	return epochInfos
}

func (backend *Backend) VerifiedSlotInfos(fromSlot uint64) map[uint64]*types.SlotInfo {
	slotInfos, err := backend.VerifiedSlotInfoDB.VerifiedSlotInfos(fromSlot)
	if err != nil {
		return nil
	}
	return slotInfos
}

func (backend *Backend) LatestEpoch() uint64 {
	return backend.ConsensusInfoDB.LatestSavedEpoch()
}

func (backend *Backend) LatestVerifiedSlot() uint64 {
	return backend.VerifiedSlotInfoDB.LatestSavedVerifiedSlot()
}

func (backed *Backend) PendingPandoraHeaders() []*eth1Types.Header {
	headers, err := backed.PandoraPendingHeaderCache.GetAll()
	if err != nil {
		return nil
	}
	return headers
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
