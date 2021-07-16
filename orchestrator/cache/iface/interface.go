package iface

import (
	"context"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	eth "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
)

type PandoraHeaderCache interface {
	Put(ctx context.Context, slot uint64, header *eth1Types.Header) error
	Get(ctx context.Context, slot uint64) (*eth1Types.Header, error)
}

// VanguardShardInfoCache interface for pandora sharding info cache
type VanguardShardInfoCache interface {
	Put(ctx context.Context, slot uint64, shardInfo *eth.PandoraShard) error
	Get(ctx context.Context, slot uint64) (*eth.PandoraShard, error)
}
