package iface

import (
	"context"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

type PandoraHeaderCache interface {
	Put(ctx context.Context, slot uint64, header *eth1Types.Header) error
	Get(ctx context.Context, slot uint64) (*eth1Types.Header, error)
	GetAll() ([]*eth1Types.Header, error)
	Remove(ctx context.Context, slot uint64)
	Purge()
}

// VanguardShardInfoCache interface for pandora sharding info cache
type VanguardShardInfoCache interface {
	Put(ctx context.Context, slot uint64, shardInfo *types.VanguardShardInfo) error
	Get(ctx context.Context, slot uint64) (*types.VanguardShardInfo, error)
	Remove(ctx context.Context, slot uint64)
	Purge()
}
