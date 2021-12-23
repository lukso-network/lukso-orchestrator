package cache

import (
	"github.com/lukso-network/lukso-orchestrator/orchestrator/utils"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	eth2Types "github.com/prysmaticlabs/eth2-types"
	"testing"
	"time"
)

func TestPandoraCache_RemoveByTime(t *testing.T) {
	pc := NewPandoraCache(1024, 0, 6, utils.NewStack())
	headerInfos, _ := testutil.GetHeaderInfosAndShardInfos(1, 25)
	for i := 0; i < 25; i++ {
		slot := uint64(i + 1)
		queueData := &PandoraCacheData{
			panHeader:      headerInfos[i].Header,
			entryTimestamp: utils.SlotStartTime(pc.genesisStartTime, eth2Types.Slot(slot), pc.secondsPerSlot),
		}
		pc.stack.Push(headerInfos[i].Header.Hash().Bytes())
		pc.cache.Add(slot, queueData)
	}

	removedHeaders := pc.RemoveByTime(time.Now())
	assert.Equal(t, 25, len(removedHeaders))
}
