package cache

import (
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	lru "github.com/hashicorp/golang-lru"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/utils"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	eth2Types "github.com/prysmaticlabs/eth2-types"
	"time"
)

// NewPandoraCache creates a pandora cache
func NewPandoraCache(size int, genesisTimestamp uint64, secondsPerSlot uint64, stack *utils.Stack) *PandoraCache {
	cache, err := lru.New(size)
	if err != nil {
		panic(err)
	}

	return &PandoraCache{
		GenericCache{
			cache:            cache,
			stack:            stack,
			genesisStartTime: genesisTimestamp,
			secondsPerSlot:   secondsPerSlot,
			inProgressSlots:  make(map[uint64]bool),
		},
	}
}

func (pc *PandoraCache) MarkInProgress(slot uint64) error {
	pc.lock.Lock()
	defer pc.lock.Unlock()

	if pc.inProgressSlots[slot] {
		return errAlreadyInProgress
	}
	pc.inProgressSlots[slot] = true
	return nil
}

func (pc *PandoraCache) MarkNotInProgress(slot uint64) error {
	pc.lock.Lock()
	defer pc.lock.Unlock()

	delete(pc.inProgressSlots, slot)
	return nil
}

func (pc *PandoraCache) Put(slot uint64, insertParams *PanCacheInsertParams) error {
	if insertParams == nil || insertParams.CurrentVerifiedHeader == nil {
		return errInvalidElement
	}
	slotStartTime := utils.SlotStartTime(pc.genesisStartTime, eth2Types.Slot(slot), pc.secondsPerSlot)

	// In initial syncing mode, pandora sends previous slots which may have already exceeded 8 secs from slot start time
	if insertParams.IsSyncing {
		slotStartTime = time.Now()
	}

	panHeader := types.CopyHeader(insertParams.CurrentVerifiedHeader)
	queueData := &PandoraCacheData{
		panHeader:      panHeader,
		entryTimestamp: slotStartTime,
	}
	val, found := pc.cache.Get(slot)
	if val != nil && found {
		// slot is already in the database.
		queueData = val.(*PandoraCacheData)
		queueData.panHeader = panHeader
	}
	pc.stack.Push(panHeader.Hash().Bytes())
	pc.cache.Add(slot, queueData)
	return nil
}

func (pc *PandoraCache) Get(slot uint64) *PandoraCacheData {
	panHeader, found := pc.cache.Get(slot)
	if found {
		return panHeader.(*PandoraCacheData)
	}
	return nil
}

func (pc *PandoraCache) Purge() {
	pc.cache.Purge()
	pc.stack.Purge()
}

func (pc *PandoraCache) ContainsHash(hash []byte) bool {
	return pc.stack.Contains(hash)
}

func (pc *PandoraCache) RemoveByTime(timeStamp time.Time) []*eth1Types.Header {
	keys := pc.cache.Keys()
	var retVal []*eth1Types.Header

	for _, key := range keys {
		slot := key.(uint64)
		queueInfo := pc.Get(slot)
		if queueInfo != nil && !pc.inProgressSlots[slot] && timeStamp.Sub(queueInfo.entryTimestamp) >= cacheRemovalInterval {
			log.WithField("slot number", slot).Debug("Removing expired slot info from pandora cache")
			retVal = append(retVal, queueInfo.panHeader)
			if queueInfo.panHeader != nil {
				pc.cache.Remove(slot)
				pc.stack.Delete(queueInfo.panHeader.Hash().Bytes())
				pc.MarkNotInProgress(slot)

			}
		}
	}
	return retVal
}

func (pc *PandoraCache) ForceDelSlot(slot uint64) {
	if slotInfo := pc.Get(slot); slotInfo != nil {
		// removed all the previous slot number from cache. Now return
		if slotInfo.panHeader != nil {
			pc.cache.Remove(slot)
			pc.stack.Delete(slotInfo.panHeader.Hash().Bytes())
			pc.MarkNotInProgress(slot)
		}
	}
}

func (pc *PandoraCache) GetInProgressSlot(slot uint64) bool {
	pc.lock.Lock()
	defer pc.lock.Unlock()
	return pc.inProgressSlots[slot]
}
