package iface

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db/filters"
	eventTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	"io"
)

// ReadOnlyDatabase defines a struct which only has read access to database methods.
type ReadOnlyDatabase interface {
	ConsensusInfo(ctx context.Context, epoch uint64) (*eventTypes.MinimalEpochConsensusInfo, error)
	ConsensusInfos(ctx context.Context, f *filters.QueryFilter) ([]*eventTypes.MinimalEpochConsensusInfo, error)
}

// Database interface with full access.
type Database interface {
	io.Closer
	ReadOnlyDatabase

	SaveConsensusInfo(ctx context.Context, consensusInfo *eventTypes.MinimalEpochConsensusInfo) error
	DatabasePath() string
	ClearDB() error
}
