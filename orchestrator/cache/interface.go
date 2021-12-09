package cache

import (
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"time"
)

type PendingQueueInterface interface {
	MarkInProgress(slot uint64) error
	MarkNotInProgress(slot uint64) error
	PutVanguardShardingInfo(slot uint64, vanShardInfo *types.VanguardShardInfo, disDel bool)
	PutPandoraHeader(slot uint64, panHeader *eth1Types.Header)
	GetSlot(slot uint64) (info *PendingQueue, found bool)
	ForceDelSlot(slot uint64)
	RemoveByTime(timeStamp time.Time) []*PendingQueue
	Purge()
}
