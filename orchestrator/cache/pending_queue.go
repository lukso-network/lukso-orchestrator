package cache

import (
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	lru "github.com/hashicorp/golang-lru"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"sync"
	"time"
)

type PendingQueue struct {
	panHeader *eth1Types.Header // pandora header hash
	vanShardInfo *types.VanguardShardInfo // vanguard sharding info
	entryTimestamp time.Time // when this data entered into the cache
	disableDelete bool // if sync is going on dont delete
}

type PendingQueueCache struct {
	pendingCache *lru.Cache // container for PendingQueue
	inProgressSlots map[uint64]bool // which slot is now processing
	lock sync.RWMutex // real write lock to prevent data from race condition
}

// NewPendingDataContainer creates PendingQueueCache with expected size
func NewPendingDataContainer (containerSize int) *PendingQueueCache {
	cache, err := lru.New(containerSize)
	if err != nil {
		panic(err)
	}
	return &PendingQueueCache{
		pendingCache: cache,
		inProgressSlots: make(map[uint64]bool),
	}
}


func (p *PendingQueueCache) MarkInProgress (slot uint64) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.inProgressSlots[slot] {
		return errAlreadyInProgress
	}
	p.inProgressSlots[slot] = true
	return nil
}

func (p *PendingQueueCache) MarkNotInProgress (slot uint64) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	delete(p.inProgressSlots, slot)
	return nil
}

func (p *PendingQueueCache) PutPandoraHeader (slot uint64, panHeader *eth1Types.Header)  {
	val, found := p.pendingCache.Get(slot)
	if found {
		// slot is already in the database.
		queueData := val.(*PendingQueue)
		queueData.panHeader = panHeader

		// update cache
		p.pendingCache.Add(slot, queueData)
	} else {
		p.pendingCache.Add(slot, &PendingQueue{
			panHeader: panHeader,
			entryTimestamp: time.Now(),
		})
	}
}

func (p *PendingQueueCache) PutVanguardShardingInfo (slot uint64, vanShardInfo *types.VanguardShardInfo, disDel bool) {
	val, found := p.pendingCache.Get(slot)
	if found {
		// slot is already in the database.
		queueData := val.(*PendingQueue)
		queueData.vanShardInfo = vanShardInfo

		// update cache
		p.pendingCache.Add(slot, queueData)
	} else {
		p.pendingCache.Add(slot, &PendingQueue{
			vanShardInfo: vanShardInfo,
			entryTimestamp: time.Now(),
			disableDelete: disDel,
		})
	}
}

func (p *PendingQueueCache) GetSlot (slot uint64) (info *PendingQueue, found bool) {
	data, found := p.pendingCache.Get(slot)
	if found {
		return data.(*PendingQueue), found
	}
	return nil, found
}

func (p *PendingQueueCache) ForceDelSlot (slot uint64) {
	p.pendingCache.Remove(slot)
	p.MarkNotInProgress(slot)
}

func (p *PendingQueueCache) RemoveByTime (timeStamp time.Time) {
	keys := p.pendingCache.Keys()

	p.lock.Lock()
	defer p.lock.Unlock()

	for _, key := range keys {
		slot := key.(uint64)
		queueInfo, _ := p.GetSlot(slot)
		if queueInfo != nil && !p.inProgressSlots[slot] && !queueInfo.disableDelete && timeStamp.Sub(queueInfo.entryTimestamp) >= cacheRemovalInterval {
			p.pendingCache.Remove(slot)
			delete(p.inProgressSlots, slot)
		}
	}
}
