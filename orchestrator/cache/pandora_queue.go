package cache

import (
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	lru "github.com/hashicorp/golang-lru"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/utils"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	eth2Types "github.com/prysmaticlabs/eth2-types"
	log "github.com/sirupsen/logrus"
	"time"
)

// NewPandoraCache creates a pandora cache
func NewPandoraCache(size int, genesisTimestamp uint64, secondsPerSlot uint64) *PandoraCache {
	cache, err := lru.New(size)
	if err != nil {
		panic(err)
	}

	return &PandoraCache{
		GenericCache{
			cache:            cache,
			stack:            utils.NewStack(),
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

func (pc *PandoraCache) Put(slot uint64, insertParams *PanCacheInsertParams) {
	if err := pc.VerifyPandoraCache(insertParams); err != nil {
		log.WithError(err).Error("cache insertion failed in pandoraCache")
		return
	}
	panHeader := types.CopyHeader(insertParams.CurrentVerifiedHeader)
	queueData := &PandoraCacheData{
		panHeader:      panHeader,
		entryTimestamp: SlotStartTime(pc.genesisStartTime, eth2Types.Slot(slot), pc.secondsPerSlot),
	}
	val, found := pc.cache.Get(slot)
	if val != nil && found {
		// slot is already in the database.
		queueData = val.(*PandoraCacheData)
		queueData.panHeader = panHeader
	}
	pc.stack.Push(panHeader.Hash().Bytes())
	pc.cache.Add(slot, queueData)
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

	pc.lock.Lock()
	defer pc.lock.Unlock()

	for _, key := range keys {
		slot := key.(uint64)
		queueInfo := pc.Get(slot)
		if queueInfo != nil && !pc.inProgressSlots[slot] && timeStamp.Sub(queueInfo.entryTimestamp) >= cacheRemovalInterval {
			log.WithField("slot number", slot).Debug("Removing expired slot info from pandora cache")
			retVal = append(retVal, queueInfo.panHeader)
			if queueInfo.panHeader != nil {
				pc.cache.Remove(slot)
				pc.stack.Delete(queueInfo.panHeader.Hash().Bytes())
				delete(pc.inProgressSlots, slot)
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
			delete(pc.inProgressSlots, slot)
		}
	}
}
