package cache

import (
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	lru "github.com/hashicorp/golang-lru"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"sync"
	"time"
)

type PendingQueue struct {
	panHeader      *eth1Types.Header        // pandora header hash
	vanShardInfo   *types.VanguardShardInfo // vanguard sharding info
	entryTimestamp time.Time                // when this data entered into the cache
	disableDelete  bool                     // if sync is going on dont delete
}

type PendingQueueCache struct {
	pendingCache    *lru.Cache      // container for PendingQueue
	inProgressSlots map[uint64]bool // which slot is now processing
	lock            sync.RWMutex    // real write lock to prevent data from race condition
}

func (q *PendingQueue) GetPanHeader() *eth1Types.Header {
	return q.panHeader
}

func (q *PendingQueue) GetVanShard() *types.VanguardShardInfo {
	return q.vanShardInfo
}

func (q *PendingQueue) GetVanShardSlotNumber() uint64 {
	if q.vanShardInfo != nil {
		return q.vanShardInfo.Slot
	}
	return 0
}
