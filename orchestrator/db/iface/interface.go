package iface

import (
	"context"
	eventTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	"io"
)

// ReadOnlyDatabase defines a struct which only has read access to database methods.
type ReadOnlyDatabase interface {
	ConsensusInfo(ctx context.Context, epoch uint64) (*eventTypes.MinimalEpochConsensusInfo, error)
	ConsensusInfos(ctx context.Context, fromEpoch uint64) ([]*eventTypes.MinimalEpochConsensusInfo, error)
	LatestSavedEpoch(ctx context.Context) (uint64, error)
}

// Database interface with full access.
type Database interface {
	io.Closer
	ReadOnlyDatabase

	SaveConsensusInfo(ctx context.Context, consensusInfo *eventTypes.MinimalEpochConsensusInfo) error
	SaveLatestEpoch(ctx context.Context) error

	DatabasePath() string
	ClearDB() error
}
