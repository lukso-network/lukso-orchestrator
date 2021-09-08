package cache

import (
	"context"
	"github.com/ethereum/go-ethereum/common/math"
	lru "github.com/hashicorp/golang-lru"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

// VanShardingInfoCache common struct for storing sharding info in a LRU cache
type VanShardingInfoCache struct {
	cache *lru.Cache
	lock  sync.RWMutex
	runCleaner sync.Once
}

type TimeBasedShardInfoContainer struct {
	shardInfo *types.VanguardShardInfo
	timestamp time.Time
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

func (vc *VanShardingInfoCache) cleanRoutine (ctx context.Context) {
	log.Debug("starting vanguard cache clean routine")
	ticker := time.NewTicker(time.Duration(cacheRemovalInterval) * time.Second)
	for {
		select {
		case <- ticker.C:
			keys := vc.cache.Keys()
			for _, key := range keys {
				slot := key.(uint64)
				item, exists := vc.cache.Get(slot)
				if exists && item != nil {
					cacheData := item.(*TimeBasedShardInfoContainer)
					timestamp := cacheData.timestamp
					timeFrame := time.Now().Add(time.Duration(cacheRemovalInterval) * time.Second * -1)
					if timestamp.Before(timeFrame) || timestamp.Equal(timeFrame) {
						vc.cache.Remove(key)
					}
				}
			}
		case <- ctx.Done():
			log.Debug("stopping pandora cache clean routine")
			return
		}
	}
}

// Put puts sharding info into a lru cache. return error if fails.
func (vc *VanShardingInfoCache) Put(ctx context.Context, slot uint64, shardInfo *types.VanguardShardInfo) error {
	cachedData := &TimeBasedShardInfoContainer{shardInfo: shardInfo, timestamp: time.Now()}
	vc.cache.Add(slot, cachedData)
	vc.runCleaner.Do(func() {
		go vc.cleanRoutine(ctx)
	})
	return nil
}

// Get retrieves sharding info from a cache. returns error if fails
func (vc *VanShardingInfoCache) Get(ctx context.Context, slot uint64) (*types.VanguardShardInfo, error) {
	item, exists := vc.cache.Get(slot)
	if exists && item != nil {
		cacheData := item.(*TimeBasedShardInfoContainer)
		shardingInfo := cacheData.shardInfo
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
