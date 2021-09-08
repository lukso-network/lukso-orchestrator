package cache

import (
	"context"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"

	eth1Types "github.com/ethereum/go-ethereum/core/types"
	lru "github.com/hashicorp/golang-lru"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

// PanHeaderCache
type PanHeaderCache struct {
	cache *lru.Cache
	lock  sync.RWMutex
	runCleaner sync.Once
}

type TimeBasedContainer struct {
	headerInfo *eth1Types.Header
	timestamp time.Time
}

// NewPanHeaderCache initializes the map and underlying cache.
func NewPanHeaderCache() *PanHeaderCache {
	cache, err := lru.New(maxCacheSize)
	if err != nil {
		panic(err)
	}
	return &PanHeaderCache{
		cache: cache,
	}
}

func (c *PanHeaderCache) cleanRoutine(ctx context.Context) {
	log.Debug("starting pandora cache clean routine")
	ticker := time.NewTicker(time.Duration(cacheRemovalInterval) * time.Second)
	for {
		select {
		case <- ticker.C:
			keys := c.cache.Keys()
			for _, key := range keys {
				slot := key.(uint64)
				item, exists := c.cache.Get(slot)
				if exists && item != nil {
					cacheData := item.(*TimeBasedContainer)
					timestamp := cacheData.timestamp
					timeFrame := time.Now().Add(time.Duration(cacheRemovalInterval) * time.Second * -1)
					if timestamp.Before(timeFrame) || timestamp.Equal(timeFrame) {
						c.cache.Remove(key)
					}
				}
			}
		case <- ctx.Done():
			log.Debug("stopping pandora cache clean routine")
			return
		}
	}
}
// Put
func (c *PanHeaderCache) Put(ctx context.Context, slot uint64, header *eth1Types.Header) error {
	copyHeader := types.CopyHeader(header)
	cacheData := &TimeBasedContainer{headerInfo: copyHeader, timestamp: time.Now()}
	c.cache.Add(slot, cacheData)
	c.runCleaner.Do(func() {
		go c.cleanRoutine(ctx)
	})
	return nil
}

// Get
func (c *PanHeaderCache) Get(ctx context.Context, slot uint64) (*eth1Types.Header, error) {
	item, exists := c.cache.Get(slot)
	if exists && item != nil {
		cacheData := item.(*TimeBasedContainer)
		header := cacheData.headerInfo
		copiedHeader := types.CopyHeader(header)
		return copiedHeader, nil
	}
	return nil, errInvalidSlot
}

func (c *PanHeaderCache) Remove(ctx context.Context, slot uint64) {
	for i := slot; i >= 0; i-- {
		if !c.cache.Remove(i) {
			// removed all the previous slot number from cache. Now return
			return
		}
	}
}

func (c *PanHeaderCache) GetAll() ([]*eth1Types.Header, error) {
	keys := c.cache.Keys()
	pendingHeaders := make([]*eth1Types.Header, 0)

	for _, key := range keys {
		slot := key.(uint64)
		item, exists := c.cache.Get(slot)
		if exists && item != nil {
			cacheData := item.(*TimeBasedContainer)
			header := cacheData.headerInfo
			copiedHeader := types.CopyHeader(header)
			pendingHeaders = append(pendingHeaders, copiedHeader)
		}
	}
	return pendingHeaders, nil
}
