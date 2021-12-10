package consensus

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"testing"
	"time"
)

func TestService_Start(t *testing.T) {
	hook := logTest.NewGlobal()
	ctx := context.Background()
	svc, _ := setup(ctx, t)
	defer func(svc *Service) {
		err := svc.Stop()
		assert.NoError(t, err)
	}(svc)

	svc.Start()
	time.Sleep(1 * time.Second)
	assert.LogsContain(t, hook, "Starting consensus service")

	err := svc.Status()
	assert.NoError(t, err)

	hook.Reset()
}

func TestService_GetShardingInfo(t *testing.T) {
	headerInfos, shardInfos := getHeaderInfosAndShardInfos(1, 6)
	tests := []struct {
		name              string
		vanShardInfos     []*types.VanguardShardInfo
		panHeaderInfos    []*types.PandoraHeaderInfo
		verifiedSlots     []uint64
		invalidSlots      []uint64
		expectedShardInfo *types.MultiShardInfo
	}{
		{
			name:           "getShardingInfo returns empty shardingInfo",
			vanShardInfos:  shardInfos,
			panHeaderInfos: headerInfos,
			verifiedSlots:  []uint64{1, 2, 3, 4, 5},
			invalidSlots:   []uint64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			hook := logTest.NewGlobal()
			ctx := context.Background()
			svc, mockedFeed := setup(ctx, t)
			defer func(svc *Service) {
				err := svc.Stop()
				require.NoError(t, err)
			}(svc)
			svc.Start()

			for i := 1; i < 5; i++ {
				slot := tt.vanShardInfos[i].Slot
				err := svc.vanguardPendingShardingCache.Put(ctx, slot, tt.vanShardInfos[i])
				require.NoError(t, err)
				mockedFeed.shardInfoFeed.Send(tt.vanShardInfos[i])

				err = svc.pandoraPendingHeaderCache.Put(ctx, slot, tt.panHeaderInfos[i].Header)
				require.NoError(t, err)
				mockedFeed.headerInfoFeed.Send(tt.panHeaderInfos[i])

				shardingInfo := svc.getShardingInfo(slot)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedShardInfo, shardingInfo)
			}

			hook.Reset()
		})
	}
}
