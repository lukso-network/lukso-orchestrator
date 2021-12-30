package consensus

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/utils"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	logTest "github.com/sirupsen/logrus/hooks/test"
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

func TestService_TriggerReorg(t *testing.T) {
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

func TestService_EarlyExitIfAlreadyInDB(t *testing.T) {
	hook := logTest.NewGlobal()
	ctx := context.Background()
	svc, _, _, _, _ := setup(ctx, t)

	headerInfos, shardInfos := testutil.GetHeaderInfosAndShardInfos(1, 10)
	for i := 0; i < 10; i++ {
		stepId := uint64(i + 1)
		multiShardInfo := utils.PrepareMultiShardData(shardInfos[i], headerInfos[i].Header, 1, 1)
		assert.NoError(t, svc.db.SaveVerifiedShardInfo(stepId, multiShardInfo))
		assert.NoError(t, svc.db.SaveLatestStepID(stepId))
		assert.NoError(t, svc.db.SaveSlotStepIndex(multiShardInfo.SlotInfo.Slot, stepId))
	}

	newPanHeaderInfo := headerInfos[5]
	newVanShardInfo := shardInfos[5]

	assert.NoError(t, svc.processVanguardShardInfo(newVanShardInfo))
	require.LogsContain(t, hook, "Vanguard block root has already verified")

	assert.NoError(t, svc.processPandoraHeader(newPanHeaderInfo))
	//require.LogsContain(t, hook, "Pandora shard header has already verified")
}

func TestService_TriggerReorg_ResolveAndValidateNext(t *testing.T) {
	hook := logTest.NewGlobal()
	ctx := context.Background()
	svc, _, _, _, _ := setup(ctx, t)

	headerInfos, shardInfos := testutil.GetHeaderInfosAndShardInfos(1, 20)
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
	// when pandora sends invalid header after reorg
	assert.NoError(t, svc.processPandoraHeader(headerInfos[6]))
	require.LogsContain(t, hook, "Invalid pandora shard for executing reorg, discarding pandora shard!")
	assert.NoError(t, svc.processPandoraHeader(newPanHeaderInfo))

	headerInfo := headerInfos[10]
	headerInfo.Header.ParentHash = newPanHeaderInfo.Header.Hash()
	headerInfo.Header.Number = big.NewInt(7)
	vanShardInfo := testutil.NewVanguardShardInfo(11, headerInfo.Header, 0, 0)
	vanShardInfo.ParentRoot = shardInfos[9].BlockRoot
	assert.NoError(t, svc.processPandoraHeader(headerInfo))
	assert.NoError(t, svc.processVanguardShardInfo(vanShardInfo))

	assert.Equal(t, uint64(7), svc.db.LatestStepID())
	expectedShardInfo := utils.PrepareMultiShardData(vanShardInfo, headerInfo.Header, 1, 1)
	actualShardInfo, err := svc.db.VerifiedShardInfo(7)
	assert.NoError(t, err)
	assert.Equal(t, true, expectedShardInfo.DeepEqual(actualShardInfo))
}

func TestService_TriggerReorg_NotResolveAndReceiveNextSlot(t *testing.T) {
	ctx := context.Background()
	svc, _, _, _, _ := setup(ctx, t)

	headerInfos, shardInfos := testutil.GetHeaderInfosAndShardInfos(1, 20)
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
	nextVanShardInfo := shardInfos[10]
	nextVanShardInfo.ParentRoot = newVanShardInfo.BlockRoot
	assert.NoError(t, svc.processVanguardShardInfo(nextVanShardInfo))
}
