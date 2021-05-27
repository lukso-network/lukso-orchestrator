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
	PandoraHeaderHash(slot uint64) (*types.HeaderHash, error)
	PandoraHeaderHashes(fromSlot uint64) ([]*types.HeaderHash, error)
	LatestSavedPandoraSlot() uint64
	LatestSavedPandoraHeaderHash() common.Hash
	GetLatestHeaderHash() common.Hash
}

// ReadOnlyVanHeaderAccessDatabase
type ReadOnlyVanHeaderAccessDatabase interface {
	VanguardHeaderHash(slot uint64) (*types.HeaderHash, error)
	VanguardHeaderHashes(fromSlot uint64) ([]*types.HeaderHash, error)
	LatestSavedVanguardSlot() uint64
	LatestSavedVanguardHeaderHash() common.Hash
	GetLatestHeaderHash() common.Hash
}

// PanHeaderAccessDatabase
type PanHeaderAccessDatabase interface {
	ReadOnlyPanHeaderAccessDatabase

	SavePandoraHeaderHash(slot uint64, headerHash *types.HeaderHash) error
	SaveLatestPandoraSlot() error
	SaveLatestPandoraHeaderHash() error
}

type VanHeaderAccessDatabase interface {
	ReadOnlyVanHeaderAccessDatabase

	SaveVanguardHeaderHash(slot uint64, headerHash *types.HeaderHash) error
	SaveLatestVanguardSlot() error
	SaveLatestVanguardHeaderHash() error
}

type RealmReadOnlyAccessDatabase interface {
	LatestVerifiedRealmSlot() (slot uint64)
}

type RealmAccessDatabase interface {
	RealmReadOnlyAccessDatabase

	SaveLatestVerifiedRealmSlot(slot uint64) (err error)
}

// Database interface with full access.
type Database interface {
	io.Closer

	ConsensusInfoAccessDatabase

	PanHeaderAccessDatabase

	VanHeaderAccessDatabase

	RealmAccessDatabase

	DatabasePath() string
	ClearDB() error
}
