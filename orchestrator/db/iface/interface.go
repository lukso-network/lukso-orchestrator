package iface

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"io"
)

// ReadOnlyDatabase defines a struct which only has read access to database methods.
type ReadOnlyConsensusInfoDatabase interface {
	ConsensusInfo(ctx context.Context, epoch uint64) (*types.MinimalEpochConsensusInfo, error)
	ConsensusInfos(fromEpoch uint64) ([]*types.MinimalEpochConsensusInfo, error)
	LatestSavedEpoch() uint64
}

// ConsensusInfoAccessDatabase
type ConsensusInfoAccessDatabase interface {
	ReadOnlyConsensusInfoDatabase

	SaveConsensusInfo(ctx context.Context, consensusInfo *types.MinimalEpochConsensusInfo) error
	SaveLatestEpoch(ctx context.Context, epoch uint64) error
}

type ReadOnlyVerifiedShardInfoDatabase interface {
	VerifiedShardInfo(stepId uint64) (*types.MultiShardInfo, error)
	VerifiedShardInfos(fromStepId uint64) (map[uint64]*types.MultiShardInfo, error)
	LatestStepID() uint64
	GetStepIdBySlot(slot uint64) (uint64, error)
	FinalizedSlot() uint64
	FinalizedEpoch() uint64
	FindAncestor(fromStepId, toStepId uint64, blockHash common.Hash) (*types.MultiShardInfo, uint64, error)
}

type VerifiedShardInfoDatabase interface {
	ReadOnlyVerifiedShardInfoDatabase

	SaveVerifiedShardInfo(stepId uint64, shardInfo *types.MultiShardInfo) error
	SaveLatestStepID(stepID uint64) error
	RemoveShardingInfos(fromStepId uint64) error
	SaveSlotStepIndex(slot, stepId uint64) error
	SaveFinalizedSlot(latestFinalizedSlot uint64) error
	SaveFinalizedEpoch(latestFinalizedEpoch uint64) error
}

type ReadOnlyInvalidSlotInfoDatabase interface {
	InvalidSlotInfo(slots uint64) (*types.SlotInfo, error)
}

type InvalidSlotDatabase interface {
	ReadOnlyInvalidSlotInfoDatabase

	SaveInvalidSlotInfo(slot uint64, slotInfo *types.SlotInfo) error
}

// Database interface with full access.
type Database interface {
	io.Closer

	ConsensusInfoAccessDatabase

	InvalidSlotDatabase

	VerifiedShardInfoDatabase

	DatabasePath() string
	ClearDB() error
}
