package consensus

import (
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"testing"
)

func TestService_CompareShardingInfo(t *testing.T) {
	headerInfos, shardInfos := testutil.GetHeaderInfosAndShardInfos(1, 10)
	for i := 0; i < 10; i++ {
		ph := headerInfos[i]
		vs := shardInfos[i]
		assert.Equal(t, true, compareShardingInfo(ph.Header, vs.ShardInfo))
	}
}
