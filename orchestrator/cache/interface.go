package cache

import (
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"time"
)

type CacheAPIs interface {
	MarkInProgress(slot uint64) error
	MarkNotInProgress(slot uint64) error
	Purge()
	ForceDelSlot(slot uint64)
	ContainsHash(hash []byte) bool
}

type VanguardCacheAPIs interface {
	CacheAPIs
	Put(slot uint64, insertParams *VanCacheInsertParams)
	Get(slot uint64) *VanguardCacheData
	RemoveByTime(timeStamp time.Time)
}

type PandoraCacheAPIs interface {
	CacheAPIs
	Put(slot uint64, insertParams *PanCacheInsertParams)
	Get(slot uint64) *PandoraCacheData
	RemoveByTime(timeStamp time.Time) []*eth1Types.Header
}
