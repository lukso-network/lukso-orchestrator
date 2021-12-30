package cache

import (
	lru "github.com/hashicorp/golang-lru"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/utils"
	eth2Types "github.com/prysmaticlabs/eth2-types"
	"time"
)

func NewVanguardCache(size int, genesisTimestamp uint64, secondsPerSlot uint64, stack *utils.Stack) *VanguardCache {
	cache, err := lru.New(size)
	if err != nil {
		panic(err)
	}

	return &VanguardCache{
		GenericCache{
			cache:            cache,
			stack:            stack,
			genesisStartTime: genesisTimestamp,
			secondsPerSlot:   secondsPerSlot,
			inProgressSlots:  make(map[uint64]bool),
		},
	}
}

func (vc *VanguardCache) MarkInProgress(slot uint64) error {
	vc.lock.Lock()
	defer vc.lock.Unlock()

	if vc.inProgressSlots[slot] {
		return errAlreadyInProgress
	}
	vc.inProgressSlots[slot] = true
	return nil
}

func (vc *VanguardCache) MarkNotInProgress(slot uint64) error {
	vc.lock.Lock()
	defer vc.lock.Unlock()

	delete(vc.inProgressSlots, slot)
	return nil
}

func (vc *VanguardCache) Put(slot uint64, insertParams *VanCacheInsertParams) error {
	if err := vc.VerifyVanguardCache(insertParams); err != nil {
		log.WithError(err).Error("cache insertion failed in vanguardCache")
		return err
	}

	val, found := vc.cache.Get(slot)
	queueData := &VanguardCacheData{
		vanShardInfo:   insertParams.CurrentShardInfo,
		entryTimestamp: utils.SlotStartTime(vc.genesisStartTime, eth2Types.Slot(slot), vc.secondsPerSlot),
		disableDelete:  insertParams.DisableDelete,
	}
	if val != nil && found {
		// slot is already in the database.
		queueData = val.(*VanguardCacheData)
		queueData.vanShardInfo = insertParams.CurrentShardInfo
	}
	vc.stack.Push(insertParams.CurrentShardInfo.BlockRoot[:])
	vc.cache.Add(slot, queueData)
	return nil
}

func (vc *VanguardCache) Get(slot uint64) *VanguardCacheData {
	vanShard, found := vc.cache.Get(slot)
	if found {
		return vanShard.(*VanguardCacheData).Copy()
	}
	return nil
}

func (vc *VanguardCache) Purge() {
	vc.cache.Purge()
	vc.stack.Purge()
}

func (vc *VanguardCache) ContainsHash(hash []byte) bool {
	return vc.stack.Contains(hash)
}

func (vc *VanguardCache) RemoveByTime(timeStamp time.Time) {
	keys := vc.cache.Keys()

	for _, key := range keys {
		slot := key.(uint64)
		if err := vc.MarkInProgress(slot); err != nil {
			continue
		}

		queueInfo := vc.Get(slot)
		if queueInfo != nil && !queueInfo.disableDelete && timeStamp.Sub(queueInfo.entryTimestamp) >= cacheRemovalInterval {
			if queueInfo.vanShardInfo != nil {
				log.WithField("slot number", slot).Debug("Removing expired slot info from vanguard cache")
				vc.cache.Remove(slot)
				vc.stack.Delete(queueInfo.vanShardInfo.BlockRoot[:])
			}
		}
		vc.MarkNotInProgress(slot)
	}
}

func (vc *VanguardCache) ForceDelSlot(slot uint64) {
	if slotInfo := vc.Get(slot); slotInfo != nil {
		// removed all the previous slot number from cache. Now return
		if slotInfo.vanShardInfo != nil {
			vc.cache.Remove(slot)
			vc.stack.Delete(slotInfo.vanShardInfo.BlockRoot[:])
			delete(vc.inProgressSlots, slot)
		}
	}
}
