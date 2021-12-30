package consensus

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/utils"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"math/big"
	"testing"
)

func TestService_CheckReorg(t *testing.T) {
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

	lastStepId := svc.db.LatestStepID()
	lastVerifiedShardInfo, err := svc.db.VerifiedShardInfo(lastStepId)
	assert.NoError(t, err)

	_, parentStepId, err := svc.checkReorg(newVanShardInfo, lastVerifiedShardInfo, lastStepId)
	assert.NoError(t, err)
	assert.Equal(t, uint64(5), parentStepId)
}
