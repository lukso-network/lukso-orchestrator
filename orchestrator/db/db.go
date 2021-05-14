package db

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db/kv"
)

// Assure that Store implements Database interface
var _ Database = &kv.Store{}

// NewDB initializes a new DB.
func NewDB(ctx context.Context, dirPath string, config *kv.Config) (Database, error) {
	return kv.NewKVStore(ctx, dirPath, config)
}
