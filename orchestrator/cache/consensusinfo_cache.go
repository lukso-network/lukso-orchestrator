package cache

import (
	lru "github.com/hashicorp/golang-lru"
	eventTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	types "github.com/prysmaticlabs/eth2-types"
	"sync"
)

var (
	// TODO- Need to be decided later
	// maxConsensusInfoSize defines the max number of entries consensus info to state cache can contain.
	maxConsensusInfoSize = 100000
)

// ConsensusInfoCache
type ConsensusInfoCache struct {
	cache *lru.Cache
	lock  sync.RWMutex
}

// NewConsensusInfoCache creates a new consensus info cache for storing/accessing previous epoch.
func NewConsensusInfoCache() *ConsensusInfoCache {
	cache, err := lru.New(maxConsensusInfoSize)
	if err != nil {
		panic(err)
	}
	return &ConsensusInfoCache{
		cache: cache,
	}
}

// ConsensusInfoByEpoch fetches consensusInfo by epoch. Returns true with a
// reference to the consensusInfo, if exists. Otherwise returns false, nil.
func (c *ConsensusInfoCache) ConsensusInfoByEpoch(epoch types.Epoch) (*eventTypes.MinConsensusInfoEvent, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	item, exists := c.cache.Get(epoch)
	if exists && item != nil {
		return item.(*eventTypes.MinConsensusInfoEvent), nil
	}
	return nil, nil
}

// AddConsensusInfoCache adds ConsensusInfoCache object to the cache.
func (c *ConsensusInfoCache) AddConsensusInfoCache(
	epoch types.Epoch,
	consensusInfo *eventTypes.MinConsensusInfoEvent,
) error {

	c.lock.Lock()
	defer c.lock.Unlock()

	c.cache.Add(epoch, consensusInfo)
	log.WithField("epoch", epoch).WithField(
		"consensusInfo", consensusInfo).Debug("caching consensus info")
	return nil
}

// ConsensusInfoByEpochRange
func (c *ConsensusInfoCache) ConsensusInfoByEpochRange(
	fromEpoch, toEpoch types.Epoch) map[types.Epoch]*eventTypes.MinConsensusInfoEvent {

	log.WithField("fromEpoch", fromEpoch).WithField("toEpoch", toEpoch).Debug("method: ConsensusInfoByEpochRange")
	consensusInfoMapping := make(map[types.Epoch]*eventTypes.MinConsensusInfoEvent)
	c.lock.Lock()
	defer c.lock.Unlock()
	for epoch := fromEpoch; epoch <= toEpoch; epoch++ {
		item, exists := c.cache.Get(epoch)
		if exists && item != nil {
			consensusInfoMapping[epoch] = item.(*eventTypes.MinConsensusInfoEvent)
		} else {
			log.WithField("epoch", epoch).Debug("consensus info not found")
		}
	}
	return consensusInfoMapping
}
