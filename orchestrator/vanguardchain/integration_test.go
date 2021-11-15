package vanguardchain

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"testing"
)

func TestDBSetup(t *testing.T) {
	ctx := context.Background()
	vanDB := dbSetup(ctx, t, 200)
	assert.Equal(t, uint64(199), vanDB.LatestSavedVerifiedSlot())
	assert.Equal(t, common.BytesToHash([]byte{199}), vanDB.LatestVerifiedHeaderHash())
	slotInfo, err := vanDB.VerifiedSlotInfo(200)
	assert.NoError(t, err)
	var expectedSlotInfo *types.SlotInfo
	assert.Equal(t, expectedSlotInfo, slotInfo)
	expectedSlotInfo = &types.SlotInfo{
		PandoraHeaderHash: common.BytesToHash([]byte{0}),
		VanguardBlockHash: common.BytesToHash([]byte{0}),
	}
	slotInfo, err = vanDB.VerifiedSlotInfo(0)
	assert.NoError(t, err)
	assert.DeepEqual(t, expectedSlotInfo, slotInfo)

	expectedEpochInfo := testutil.NewMinimalConsensusInfo(5).ConvertToEpochInfo()
	assert.Equal(t, uint64(199), vanDB.LatestSavedEpoch())

	consensusInfo, err := vanDB.ConsensusInfo(ctx, 5)
	assert.NoError(t, err)
	assert.DeepEqual(t, expectedEpochInfo, consensusInfo)
}
