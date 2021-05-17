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

// ReadOnlyPanHeaderAccessDatabase
type ReadOnlyPanHeaderAccessDatabase interface {
	PandoraHeaderHash(slot uint64) (*types.PanHeaderHash, error)
	PandoraHeaderHashes(fromSlot uint64) ([]*types.PanHeaderHash, error)
	LatestSavedPandoraSlot() uint64
	LatestSavedPandoraHeaderHash() common.Hash
	GetLatestHeaderHash() common.Hash
}

// PanHeaderAccessDatabase
type PanHeaderAccessDatabase interface {
	ReadOnlyPanHeaderAccessDatabase

	SavePandoraHeaderHash(slot uint64, headerHash *types.PanHeaderHash) error
	SaveLatestPandoraSlot() error
	SaveLatestPandoraHeaderHash() error
}

// Database interface with full access.
type Database interface {
	io.Closer

	ConsensusInfoAccessDatabase

	PanHeaderAccessDatabase

	DatabasePath() string
	ClearDB() error
}
