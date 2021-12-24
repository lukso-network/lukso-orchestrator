package cache

import (
	"github.com/lukso-network/lukso-orchestrator/orchestrator/utils"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"testing"
)

func TestVanguardCache_Put(t *testing.T) {
	pc := NewVanguardCache(1024, 0, 6, utils.NewStack())
	_, vanShardInfos := testutil.GetHeaderInfosAndShardInfos(1, 25)
	for _, vanShard := range vanShardInfos {
		pc.Put(vanShard.Slot, &VanCacheInsertParams{
			CurrentShardInfo:      vanShard,
			LastVerifiedShardInfo: nil,
			DisableDelete:         true,
		})
	}
	assert.Equal(t, 25, len(pc.cache.Keys()))
}

func TestVanguardCache_PutInConsecutiveShard(t *testing.T) {
	pc := NewVanguardCache(1024, 0, 6, utils.NewStack())
	_, vanShardInfos := testutil.GetHeaderInfosAndShardInfos(1, 25)
	curShardInfo := vanShardInfos[24]

}
