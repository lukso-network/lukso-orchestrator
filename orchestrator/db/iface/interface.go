package iface

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"io"
)

// ReadOnlyDatabase defines a struct which only has read access to database methods.
type ReadOnlyDatabase interface {
	ConsensusInfo(ctx context.Context, epoch uint64) (*types.MinimalEpochConsensusInfo, error)
	ConsensusInfos(fromEpoch uint64) ([]*types.MinimalEpochConsensusInfo, error)
	LatestSavedEpoch() (uint64, error)
}

// PanHeaderAccessDatabase
type PanHeaderAccessDatabase interface {
	PanHeader(slot uint64) (*types.PanBlockHeader, error)
	PanHeaders(fromSlot uint64) ([]*types.PanBlockHeader, error)
	LatestSavedPanBlockNum() (uint64, error)
	SavePanHeader(header *types.PanBlockHeader) error
	SaveLatestPanSlot() error
	SaveLatestPanBlockNum() error
}

// Database interface with full access.
type Database interface {
	io.Closer
	ReadOnlyDatabase

	SaveConsensusInfo(ctx context.Context, consensusInfo *types.MinimalEpochConsensusInfo) error
	SaveLatestEpoch(ctx context.Context) error

	DatabasePath() string
	ClearDB() error
}
