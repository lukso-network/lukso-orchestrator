package cache

import (
	"context"
	lru "github.com/hashicorp/golang-lru"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"github.com/pkg/errors"
	"sync"
)

var (
	// maxPanHeaderCacheSize with 1024 consensus infos will be 1024 * 1.5kb.
	maxPanHeaderCacheSize = 1 << 10

	// errInvalidSlot
	errInvalidSlot = errors.New("Invalid slot")
)

// PanHeaderCache
type PanHeaderCache struct {
	cache *lru.Cache
	lock  sync.RWMutex
}

// NewPanHeaderCache initializes the map and underlying cache.
func NewPanHeaderCache() *PanHeaderCache {
	cache, err := lru.New(maxPanHeaderCacheSize)
	if err != nil {
		panic(err)
	}
	return &PanHeaderCache{
		cache: cache,
	}
}

// Put
func (c *PanHeaderCache) Put(ctx context.Context, slot uint64, header *types.PanBlockHeader) error {
	copyHeader := header.Copy()
	c.cache.Add(slot, copyHeader)
	return nil
}

// Get
func (c *PanHeaderCache) Get(ctx context.Context, slot uint64) (*types.PanBlockHeader, error) {
	item, exists := c.cache.Get(slot)

	if exists && item != nil {
		return item.(*types.PanBlockHeader).Copy(), nil
	}
	return nil, errInvalidSlot
}

// GetStatus
func (c *PanHeaderCache) GetStatus(ctx context.Context, slot uint64) (types.Status, error) {
	item, exists := c.cache.Get(slot)

	if exists && item != nil {
		return item.(*types.PanBlockHeader).Copy().Status, nil
	}
	return types.Status(-1), errInvalidSlot
}
