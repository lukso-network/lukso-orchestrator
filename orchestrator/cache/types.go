package cache

import (
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	lru "github.com/hashicorp/golang-lru"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/utils"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"sync"
	"time"
)

type PandoraCacheData struct {
	panHeader      *eth1Types.Header // pandora header hash
	entryTimestamp time.Time         // when this data entered into the cache
}

type VanguardCacheData struct {
	vanShardInfo   *types.VanguardShardInfo // vanguard sharding info
	entryTimestamp time.Time                // when this data entered into the cache
	disableDelete  bool                     // if sync is going on dont delete
}

type GenericCache struct {
	lock             sync.Mutex
	cache            *lru.Cache
	stack            *utils.Stack
	inProgressSlots  map[uint64]bool // which slot is now processing
	genesisStartTime uint64
	secondsPerSlot   uint64
}

type PandoraCache struct {
	GenericCache
}

type VanguardCache struct {
	GenericCache
}

type PanCacheInsertParams struct {
	CurrentVerifiedHeader *eth1Types.Header
	LastVerifiedShardInfo *types.MultiShardInfo
}

type VanCacheInsertParams struct {
	CurrentShardInfo      *types.VanguardShardInfo
	LastVerifiedShardInfo *types.MultiShardInfo
	DisableDelete         bool
}

func (q *PandoraCacheData) GetPanHeader() *eth1Types.Header {
	return q.panHeader
}

func (q *VanguardCacheData) GetVanShard() *types.VanguardShardInfo {
	return q.vanShardInfo
}

func (q *VanguardCacheData) GetVanShardSlotNumber() uint64 {
	if q.vanShardInfo != nil {
		return q.vanShardInfo.Slot
	}
	return 0
}

func (q *VanguardCacheData) IsFinalizedSlot() bool {
	return q.disableDelete
}