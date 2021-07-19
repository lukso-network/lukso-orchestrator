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
	GetLatestEpoch() uint64
}

// ConsensusInfoAccessDatabase
type ConsensusInfoAccessDatabase interface {
	ReadOnlyConsensusInfoDatabase

	SaveConsensusInfo(ctx context.Context, consensusInfo *types.MinimalEpochConsensusInfo) error
	SaveLatestEpoch(ctx context.Context) error
}

type ReadOnlyVerifiedSlotInfoDatabase interface {
	VerifiedSlotInfo(slot uint64) (*types.SlotInfo, error)
	LatestSavedVerifiedSlot() uint64
	InMemoryLatestVerifiedSlot() uint64
	LatestVerifiedHeaderHash() common.Hash
	InMemoryLatestVerifiedHeaderHash() common.Hash
}

type VerifiedSlotDatabase interface {
	ReadOnlyVerifiedSlotInfoDatabase

	SaveVerifiedSlotInfo(slot uint64, slotInfo *types.SlotInfo) error
	SaveLatestVerifiedSlot(ctx context.Context) error
	SaveLatestVerifiedHeaderHash() error
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

	VerifiedSlotDatabase

	InvalidSlotDatabase

	DatabasePath() string
	ClearDB() error
}
