package iface

import (
	"context"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
)

type PandoraHeaderCache interface {
	Put(ctx context.Context, slot uint64, header *eth1Types.Header) error
	Get(ctx context.Context, slot uint64) (*eth1Types.Header, error)
}
