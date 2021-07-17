package cache

import (
	"context"
	lru "github.com/hashicorp/golang-lru"
	eth "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"sync"
)

// VanShardingInfoCache common struct for storing sharding info in a LRU cache
type VanShardingInfoCache struct {
	cache *lru.Cache
	lock  sync.RWMutex
}

// NewVanShardInfoCache initializes the map and underlying cache.
func NewVanShardInfoCache(cacheSize int) *VanShardingInfoCache {
	cache, err := lru.New(cacheSize)
	if err != nil {
		panic(err)
	}
	return &VanShardingInfoCache{
		cache: cache,
	}
}

// Put puts sharding info into a lru cache. return error if fails.
func (vc *VanShardingInfoCache) Put(ctx context.Context, slot uint64, shardInfo *eth.PandoraShard) error {
	if vc.cache.Add(slot, shardInfo) {
		return errAddingCache
	}
	return nil
}

// Get retrieves sharding info from a cache. returns error if fails
func (vc *VanShardingInfoCache) Get(ctx context.Context, slot uint64) (*eth.PandoraShard, error) {
	item, exists := vc.cache.Get(slot)
	if exists && item != nil {
		shardingInfo := item.(*eth.PandoraShard)
		return shardingInfo, nil
	}
	return nil, errInvalidSlot
}
