package consensus

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/utils"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"math/big"
	"testing"
	"time"
)

// TestService_CheckingFirstStepVerification checks consensus service for verifying the first entry
func TestService_CheckingFirstStepVerification(t *testing.T) {
	ctx := context.Background()
	svc, _, db, _, _ := setup(ctx, t)
	defer svc.Stop()

	headerInfos, shardInfos := testutil.GetHeaderInfosAndShardInfos(1, 1)
	assert.NoError(t, svc.processVanguardShardInfo(shardInfos[0]))
	time.Sleep(1 * time.Second)
	assert.NoError(t, svc.processPandoraHeader(headerInfos[0]))
	assert.Equal(t, uint64(1), db.LatestStepID())

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

	headerInfos, shardInfos := testutil.GetHeaderInfosAndShardInfos(1, 2)
	multiShardInfo := utils.PrepareMultiShardData(shardInfos[0], headerInfos[0].Header, 1, 1)

	assert.NoError(t, svc.db.SaveVerifiedShardInfo(1, multiShardInfo))
	assert.NoError(t, svc.db.SaveLatestStepID(1))
	assert.NoError(t, svc.db.SaveSlotStepIndex(multiShardInfo.SlotInfo.Slot, 1))

	assert.NoError(t, svc.processVanguardShardInfo(shardInfos[1]))
	assert.NoError(t, svc.processPandoraHeader(headerInfos[1]))
	assert.Equal(t, uint64(2), db.LatestStepID())
	assert.Equal(t, uint64(0), db.FinalizedSlot())

	expectedShardInfo := utils.PrepareMultiShardData(shardInfos[1], headerInfos[1].Header, 1, 1)
	actualShardInfo, err := db.VerifiedShardInfo(2)
	assert.NoError(t, err)
	assert.Equal(t, true, expectedShardInfo.DeepEqual(actualShardInfo))
}

func TestService_TriggerReorgAndCheckDBConsistency(t *testing.T) {
	ctx := context.Background()
	svc, _, _, _, _ := setup(ctx, t)
	defer svc.Stop()

	headerInfos, shardInfos := testutil.GetHeaderInfosAndShardInfos(1, 10)
	for i := 0; i < 9; i++ {
		stepId := uint64(i + 1)
		multiShardInfo := utils.PrepareMultiShardData(shardInfos[i], headerInfos[i].Header, 1, 1)
		assert.NoError(t, svc.db.SaveVerifiedShardInfo(stepId, multiShardInfo))
		assert.NoError(t, svc.db.SaveLatestStepID(stepId))
		assert.NoError(t, svc.db.SaveSlotStepIndex(multiShardInfo.SlotInfo.Slot, stepId))
	}

	newVanShardInfo := shardInfos[9]
	newPanHeaderInfo := headerInfos[9]
	newVanShardInfo.ParentRoot = shardInfos[4].BlockRoot
	newPanHeaderInfo.Header.ParentHash = headerInfos[4].Header.Hash()

	assert.NoError(t, svc.processVanguardShardInfo(newVanShardInfo))

	latestStepId := svc.db.LatestStepID()
	assert.Equal(t, uint64(5), latestStepId)

	actualHeadShardInfoInDB, err := svc.db.VerifiedShardInfo(latestStepId)
	assert.NoError(t, err)
	expectedHeadShardInfoInDB := utils.PrepareMultiShardData(shardInfos[4], headerInfos[4].Header, 1, 1)
	assert.Equal(t, true, expectedHeadShardInfoInDB.DeepEqual(actualHeadShardInfoInDB))
}

func TestService_TriggerReorgAndVerifyNextShard(t *testing.T) {
	ctx := context.Background()
	svc, _, _, _, _ := setup(ctx, t)
	defer svc.Stop()

	headerInfos, shardInfos := testutil.GetHeaderInfosAndShardInfos(1, 10)
	for i := 0; i < 9; i++ {
		stepId := uint64(i + 1)
		multiShardInfo := utils.PrepareMultiShardData(shardInfos[i], headerInfos[i].Header, 1, 1)
		assert.NoError(t, svc.db.SaveVerifiedShardInfo(stepId, multiShardInfo))
		assert.NoError(t, svc.db.SaveLatestStepID(stepId))
		assert.NoError(t, svc.db.SaveSlotStepIndex(multiShardInfo.SlotInfo.Slot, stepId))
	}

	newPanHeaderInfo := headerInfos[9]
	newPanHeaderInfo.Header.ParentHash = headerInfos[4].Header.Hash()
	newPanHeaderInfo.Header.Number = big.NewInt(6)
	newVanShardInfo := testutil.NewVanguardShardInfo(10, newPanHeaderInfo.Header, 0, 0)
	newVanShardInfo.ParentRoot = shardInfos[4].BlockRoot

	assert.NoError(t, svc.processVanguardShardInfo(newVanShardInfo))
	assert.NoError(t, svc.processPandoraHeader(newPanHeaderInfo))

	assert.Equal(t, uint64(6), svc.db.LatestStepID())
	expectedShardInfo := utils.PrepareMultiShardData(newVanShardInfo, newPanHeaderInfo.Header, 1, 1)
	actualShardInfo, err := svc.db.VerifiedShardInfo(6)
	assert.NoError(t, err)
	assert.Equal(t, true, expectedShardInfo.DeepEqual(actualShardInfo))
}
