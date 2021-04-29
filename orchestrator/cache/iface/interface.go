package iface

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

type PandoraHeaderCache interface {
	Put(ctx context.Context, slot uint64, header *types.PanBlockHeader) error
	Get(ctx context.Context, slot uint64) (*types.PanBlockHeader, error)
	GetStatus(ctx context.Context, slot uint64) (types.Status, error)
}
