package consensus

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/utils"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"testing"
	"time"
)

func TestService_Start(t *testing.T) {
	hook := logTest.NewGlobal()
	ctx := context.Background()
	svc, _, _, _, _ := setup(ctx, t)
	defer svc.Stop()

	svc.Start()
	time.Sleep(1 * time.Second)
	assert.LogsContain(t, hook, "Starting consensus service")
	hook.Reset()
}

// TestService_CheckingFirstStepVerification checks consensus service for verifying the first entry
func TestService_CheckingFirstStepVerification(t *testing.T) {
	ctx := context.Background()
	svc, _, db, _, _ := setup(ctx, t)
	defer svc.Stop()

	time.Sleep(1 * time.Second)
	headerInfos, shardInfos := getHeaderInfosAndShardInfos(1, 1)
	assert.NoError(t, svc.processVanguardShardInfo(shardInfos[0]))
	time.Sleep(1 * time.Second)
	assert.NoError(t, svc.processPandoraHeader(headerInfos[0]))

	assert.Equal(t, uint64(1), db.LatestStepID())
	assert.Equal(t, uint64(1), db.FinalizedSlot())

	expectedShardInfo := utils.PrepareMultiShardData(shardInfos[0], headerInfos[0].Header, 1, 1)
	actualShardInfo, err := db.VerifiedShardInfo(1)
	assert.NoError(t, err)
	assert.Equal(t, true, expectedShardInfo.DeepEqual(actualShardInfo))
}

// TestService_CheckingWithoutReorg checks for verifying the second entry
func TestService_CheckingWithoutReorg(t *testing.T) {
	ctx := context.Background()
	svc, _, db, _, _ := setup(ctx, t)
	defer svc.Stop()

	headerInfos, shardInfos := getHeaderInfosAndShardInfos(1, 2)
	multiShardInfo := utils.PrepareMultiShardData(shardInfos[0], headerInfos[0].Header, 1, 1)
	db.SaveLatestStepID(1)

	assert.NoError(t, svc.db.SaveVerifiedShardInfo(1, multiShardInfo))
	assert.NoError(t, svc.db.SaveLatestStepID(1))
	assert.NoError(t, svc.db.SaveSlotStepIndex(multiShardInfo.SlotInfo.Slot, 1))

	time.Sleep(1 * time.Second)

	assert.NoError(t, svc.processVanguardShardInfo(shardInfos[1]))
	time.Sleep(2 * time.Second)
	assert.NoError(t, svc.processPandoraHeader(headerInfos[1]))

	assert.Equal(t, uint64(2), db.LatestStepID())
	assert.Equal(t, uint64(0), db.FinalizedSlot())

	expectedShardInfo := utils.PrepareMultiShardData(shardInfos[1], headerInfos[1].Header, 1, 1)
	actualShardInfo, err := db.VerifiedShardInfo(2)
	assert.NoError(t, err)
	assert.Equal(t, true, expectedShardInfo.DeepEqual(actualShardInfo))
}
