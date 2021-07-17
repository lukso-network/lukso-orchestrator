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
	defer svc.Stop()

	svc.Start()
	time.Sleep(1 * time.Second)
	assert.LogsContain(t, hook, "Starting consensus service")
	hook.Reset()
}

func TestService(t *testing.T) {
	headerInfos, shardInfos := getHeaderInfosAndShardInfos(1, 6)
	tests := []struct {
		name              string
		vanShardInfos     []*types.VanguardShardInfo
		panHeaderInfos    []*types.PandoraHeaderInfo
		verifiedSlots     []uint64
		invalidSlots      []uint64
		expectedOutputMsg string
	}{
		{
			name:              "Test subscription process",
			vanShardInfos:     shardInfos,
			panHeaderInfos:    headerInfos,
			verifiedSlots:     []uint64{1, 2, 3, 4, 5},
			invalidSlots:      []uint64{},
			expectedOutputMsg: "Successfully verified sharding info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			hook := logTest.NewGlobal()
			ctx := context.Background()
			svc, mockedFeed := setup(ctx, t)
			defer svc.Stop()
			svc.Start()

			for i := 0; i < 5; i++ {
				slot := tt.vanShardInfos[i].Slot
				svc.vanguardPendingShardingCache.Put(ctx, slot, tt.vanShardInfos[i])
				mockedFeed.shardInfoFeed.Send(tt.vanShardInfos[i])

				time.Sleep(5 * time.Millisecond)

				svc.pandoraPendingHeaderCache.Put(ctx, slot, tt.panHeaderInfos[i].Header)
				mockedFeed.headerInfoFeed.Send(tt.panHeaderInfos[i])

				time.Sleep(100 * time.Millisecond)
				slotInfo, err := svc.verifiedSlotInfoDB.VerifiedSlotInfo(slot)
				require.NoError(t, err)
				assert.NotNil(t, slotInfo)
			}

			time.Sleep(2 * time.Second)
			assert.LogsContain(t, hook, tt.expectedOutputMsg)
			hook.Reset()
		})
	}
}
