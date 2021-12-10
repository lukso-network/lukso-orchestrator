package cache

import (
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	lru "github.com/hashicorp/golang-lru"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	log "github.com/sirupsen/logrus"
	"time"
)

// NewPendingDataContainer creates PendingQueueCache with expected size
func NewPendingDataContainer(containerSize int) *PendingQueueCache {
	cache, err := lru.New(containerSize)
	if err != nil {
		panic(err)
	}
	return &PendingQueueCache{
		pendingCache:    cache,
		inProgressSlots: make(map[uint64]bool),
	}
}

func (p *PendingQueueCache) MarkInProgress(slot uint64) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.inProgressSlots[slot] {
		return errAlreadyInProgress
	}
	p.inProgressSlots[slot] = true
	return nil
}

func (p *PendingQueueCache) MarkNotInProgress(slot uint64) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	delete(p.inProgressSlots, slot)
	return nil
}

func (p *PendingQueueCache) PutPandoraHeader(slot uint64, header *eth1Types.Header) {
	panHeader := types.CopyHeader(header)
	val, found := p.pendingCache.Get(slot)
	if val != nil && found {
		// slot is already in the database.
		queueData := val.(*PendingQueue)
		queueData.panHeader = panHeader

		// update cache
		p.pendingCache.Add(slot, queueData)
	} else {
		p.pendingCache.Add(slot, &PendingQueue{
			panHeader:      panHeader,
			entryTimestamp: time.Now(),
		})
	}
}

func (p *PendingQueueCache) PutVanguardShardingInfo(slot uint64, vanShardInfo *types.VanguardShardInfo, disDel bool) {
	val, found := p.pendingCache.Get(slot)
	if val != nil && found {
		// slot is already in the database.
		queueData := val.(*PendingQueue)
		queueData.vanShardInfo = vanShardInfo

		// update cache
		p.pendingCache.Add(slot, queueData)
	} else {
		p.pendingCache.Add(slot, &PendingQueue{
			vanShardInfo:   vanShardInfo,
			entryTimestamp: time.Now(),
			disableDelete:  disDel,
		})
	}
}

func (p *PendingQueueCache) GetSlot(slot uint64) (info *PendingQueue, found bool) {
	data, found := p.pendingCache.Get(slot)
	if found {
		return data.(*PendingQueue), found
	}
	return nil, found
}

func (p *PendingQueueCache) ForceDelSlot(slot uint64) {
	for i := slot; i > 0; i-- {
		if p.pendingCache.Contains(i) {
			// removed all the previous slot number from cache. Now return
			p.pendingCache.Remove(i)
		}
	}
}

func (p *PendingQueueCache) Purge() {
	p.pendingCache.Purge()
}

func (p *PendingQueueCache) RemoveByTime(timeStamp time.Time) []*PendingQueue {
	keys := p.pendingCache.Keys()
	var retVal []*PendingQueue

	p.lock.Lock()
	defer p.lock.Unlock()

	for _, key := range keys {
		slot := key.(uint64)
		queueInfo, _ := p.GetSlot(slot)
		if queueInfo != nil && !p.inProgressSlots[slot] && !queueInfo.disableDelete && timeStamp.Sub(queueInfo.entryTimestamp) >= cacheRemovalInterval {
			log.WithField("slot number", slot).Debug("Removing expired slot info from cache")
			retVal = append(retVal, queueInfo)
			p.pendingCache.Remove(slot)
			delete(p.inProgressSlots, slot)
		}
	}
	return retVal
}
