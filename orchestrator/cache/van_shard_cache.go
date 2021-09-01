package cache

import (
	"context"
	"github.com/ethereum/go-ethereum/common/math"
	lru "github.com/hashicorp/golang-lru"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"sync"
)

// VanShardingInfoCache common struct for storing sharding info in a LRU cache
type VanShardingInfoCache struct {
	cache *lru.Cache
	lock  sync.RWMutex
}

// NewVanShardInfoCache initializes the map and underlying cache.
func NewVanShardInfoCache(cacheSize int) *VanShardingInfoCache {
	cache, err := lru.New(math.MaxInt32)
	if err != nil {
		panic(err)
	}
	return &VanShardingInfoCache{
		cache: cache,
	}
}

// Put puts sharding info into a lru cache. return error if fails.
func (vc *VanShardingInfoCache) Put(ctx context.Context, slot uint64, shardInfo *types.VanguardShardInfo) error {
	vc.cache.Add(slot, shardInfo)
	return nil
}

// Get retrieves sharding info from a cache. returns error if fails
func (vc *VanShardingInfoCache) Get(ctx context.Context, slot uint64) (*types.VanguardShardInfo, error) {
	item, exists := vc.cache.Get(slot)
	if exists && item != nil {
		shardingInfo := item.(*types.VanguardShardInfo)
		return shardingInfo, nil
	}
	return nil, errInvalidSlot
}

func (vc *VanShardingInfoCache) Remove(ctx context.Context, slot uint64) {
	for i := slot; i >= 0; i-- {
		if !vc.cache.Remove(i) {
			// removed all the previous slot number from cache. Now return
			return
		}
	}
}
