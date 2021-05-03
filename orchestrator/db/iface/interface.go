package iface

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"io"
)

// ReadOnlyDatabase defines a struct which only has read access to database methods.
type ReadOnlyDatabase interface {
	ConsensusInfo(ctx context.Context, epoch uint64) (*types.MinimalEpochConsensusInfo, error)
	ConsensusInfos(fromEpoch uint64) ([]*types.MinimalEpochConsensusInfo, error)
	LatestSavedEpoch() uint64
}

// PanHeaderAccessDatabase
type PanHeaderAccessDatabase interface {
	PandoraHeaderHash(slot uint64) (*types.PanHeaderHash, error)
	PandoraHeaderHashes(fromSlot uint64) ([]*types.PanHeaderHash, error)
	SavePandoraHeaderHash(slot uint64, headerHash *types.PanHeaderHash) error
	SaveLatestPandoraSlot() error
	LatestSavedPandoraSlot() uint64
	SaveLatestPandoraHeaderHash() error
	LatestSavedPandoraHeaderHash() common.Hash
	GetLatestHeaderHash() common.Hash
}

// Database interface with full access.
type Database interface {
	io.Closer
	ReadOnlyDatabase
	PanHeaderAccessDatabase

	SaveConsensusInfo(ctx context.Context, consensusInfo *types.MinimalEpochConsensusInfo) error
	SaveLatestEpoch(ctx context.Context) error

	DatabasePath() string
	ClearDB() error
}
