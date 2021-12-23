package api

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/utils"
	"github.com/pkg/errors"

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
	ReorgFeed            conIface.ReorgInfoFeed

	// db reference
	ConsensusInfoDB     db.ROnlyConsensusInfoDB
	VerifiedShardInfoDB db.ROnlyVerifiedShardInfoDB
	InvalidSlotInfoDB   db.ROnlyInvalidSlotInfoDB

	// cache reference
	PanHeaderCache cache.PandoraInterface
	VanShardCache  cache.VanguardInterface
}

func (b *Backend) SubscribeNewEpochEvent(ch chan<- *types.MinimalEpochConsensusInfoV2) event.Subscription {
	return b.ConsensusInfoFeed.SubscribeMinConsensusInfoEvent(ch)
}

func (b *Backend) SubscribeNewVerifiedSlotInfoEvent(ch chan<- *types.SlotInfoWithStatus) event.Subscription {
	return b.VerifiedSlotInfoFeed.SubscribeVerifiedSlotInfoEvent(ch)
}

func (b *Backend) SubscribeNewReorgInfoEvent(ch chan<- *types.Reorg) event.Subscription {
	return b.ReorgFeed.SubscribeReorgInfoEvent(ch)
}

func (b *Backend) ConsensusInfoByEpochRange(fromEpoch uint64) ([]*types.MinimalEpochConsensusInfoV2, error) {
	consensusInfos, err := b.ConsensusInfoDB.ConsensusInfos(fromEpoch)
	if err != nil {
		return nil, err
	}

	epochInfos := make([]*types.MinimalEpochConsensusInfoV2, len(consensusInfos))
	for i, epochInfo := range consensusInfos {
		epochInfoV1 := epochInfo.ConvertToEpochInfoV2()
		epochInfos[i] = epochInfoV1
	}
	return epochInfos, nil
}

func (b *Backend) LatestEpochInfo(ctx context.Context) (*types.MinimalEpochConsensusInfoV2, error) {
	latestEpoch := b.LatestEpoch()
	latestEpochInfo, err := b.ConsensusInfoDB.ConsensusInfo(ctx, latestEpoch)
	if err != nil {
		return nil, err
	}

	if latestEpochInfo == nil {
		return nil, nil
	}

	return latestEpochInfo.ConvertToEpochInfoV2(), nil
}

func (b *Backend) StepId(slot uint64) uint64 {
	stepId, err := b.VerifiedShardInfoDB.GetStepIdBySlot(slot)
	if err != nil {
		return 0
	}
	return stepId
}

func (b *Backend) LatestStepId() uint64 {
	return b.VerifiedShardInfoDB.LatestStepID()
}

func (b *Backend) VerifiedSlotInfos(fromSlot uint64) (map[uint64]*types.BlockStatus, error) {
	// Short circuit for slot zero
	if fromSlot == 0 {
		return nil, nil
	}

	// Removing slot infos from verified slot info db
	stepId, err := b.VerifiedShardInfoDB.GetStepIdBySlot(fromSlot)
	if err != nil {
		log.WithError(err).WithField("fromSlot", fromSlot).WithField("stepId", stepId).Error("Could not found step id from DB")
		return nil, errors.Wrap(err, "Could not found step id from DB")
	}

	shardInfos, err := b.VerifiedShardInfoDB.VerifiedShardInfos(stepId)
	if err != nil {
		return nil, errors.Wrap(err, "Could not found verified shard infos from DB")
	}

	latestStepId := b.VerifiedShardInfoDB.LatestStepID()
	finalizedSlot := b.VerifiedShardInfoDB.FinalizedSlot()
	blockStatus := make(map[uint64]*types.BlockStatus)

	for si := stepId; si <= latestStepId; si++ {
		shardInfo := shardInfos[si]
		if shardInfo != nil {
			blockStatus[si] = utils.ConvertShardInfoToBlockStatus(shardInfo, types.Verified, finalizedSlot)
		}
	}

	return blockStatus, nil
}

func (b *Backend) LatestEpoch() uint64 {
	return b.ConsensusInfoDB.LatestSavedEpoch()
}

func (b *Backend) LatestFinalizedSlot() uint64 {
	return b.VerifiedShardInfoDB.FinalizedSlot()
}

// GetSlotStatus
func (b *Backend) GetSlotStatus(ctx context.Context, slot uint64, hash common.Hash, requestFrom bool) types.Status {
	logPrinter := func(stat types.Status) {
		log.WithField("slot", slot).
			WithField("status", stat).
			Debug("Verification status")
	}

	if !requestFrom {
		// if request is from vanguard, check the slot is in vanguard cache.
		// if so return pending
		if queueInfo := b.VanShardCache.Get(slot); queueInfo != nil {
			if queueInfo.IsFinalizedSlot() {
				logPrinter(types.Verified)
				return types.Verified
			}

			// data found in the queue. So it's pending
			logPrinter(types.Pending)
			return types.Pending
		}
	} else {
		// if request is from pandora, check the slot is in pandora cache
		// if so return pending
		if queueInfo := b.PanHeaderCache.Get(slot); queueInfo != nil {
			// data found in the queue. So it's pending
			logPrinter(types.Pending)
			return types.Pending
		}
	}

	// Removing slot infos from verified slot info db
	stepId, err := b.VerifiedShardInfoDB.GetStepIdBySlot(slot)
	if err != nil {
		log.WithError(err).WithField("slot", slot).WithField("stepId", stepId).Error("Could not found step id from DB")
		return types.Invalid
	}

	// finally found in the database so return immediately so that no other db call happens
	if shardInfo, _ := b.VerifiedShardInfoDB.VerifiedShardInfo(stepId); shardInfo != nil && shardInfo.NotNil() {
		panHeaderHash := shardInfo.GetPanShardRoot()
		vanHeaderHash := shardInfo.GetVanSlotRoot()

		if requestFrom && panHeaderHash == hash {
			logPrinter(types.Verified)
			return types.Verified
		}

		if !requestFrom && vanHeaderHash == hash {
			logPrinter(types.Verified)
			return types.Verified
		}
	}

	log.WithError(ErrHeaderHashMisMatch).Warn("Failed to match hash with verified slot's hash")
	logPrinter(types.Invalid)
	return types.Invalid
}

// VerifiedShardInfos retrieves shard infos from verified shard info db
func (b *Backend) VerifiedShardInfos(fromSlot uint64) (map[uint64]*types.MultiShardInfo, error) {
	// Short circuit for slot zero
	if fromSlot == 0 {
		return nil, nil
	}

	// Removing slot infos from verified slot info db
	stepId, err := b.VerifiedShardInfoDB.GetStepIdBySlot(fromSlot)
	if err != nil {
		log.WithError(err).WithField("fromSlot", fromSlot).Error("Could not found step id from DB")
		return nil, errors.Wrap(err, "Could not found step id from DB")
	}

	shardInfos, err := b.VerifiedShardInfoDB.VerifiedShardInfos(stepId)
	if err != nil {
		return nil, errors.Wrap(err, "Could not found verified shard infos from DB")
	}

	log.WithField("finalizedSlot", fromSlot).WithField("fromStepId", stepId).
		WithField("lenShardInfo", len(shardInfos)).Debug("Serving shard infos from verified db")

	return shardInfos, nil
}
