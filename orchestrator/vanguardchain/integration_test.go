package vanguardchain

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	testDB "github.com/lukso-network/lukso-orchestrator/orchestrator/db/testing"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"testing"
)

func db_setup(ctx context.Context, t *testing.T, numberOfElements byte) db.Database {
	vanguardDb := testDB.SetupDB(t)
	var slotInfo *types.SlotInfo
	for i := byte(0); i < numberOfElements; i++ {
		slotInfo = &types.SlotInfo{
			PandoraHeaderHash: common.BytesToHash([]byte{i}),
			VanguardBlockHash: common.BytesToHash([]byte{i}),
		}

		err := vanguardDb.SaveVerifiedSlotInfo(uint64(i), slotInfo)
		assert.NoError(t, err)
	}
	err := vanguardDb.SaveLatestVerifiedSlot(ctx, uint64(numberOfElements - 1))
	assert.NoError(t, err)
	assert.NotNil(t, slotInfo)
	assert.NotNil(t, slotInfo.PandoraHeaderHash)
	err = vanguardDb.SaveLatestVerifiedHeaderHash(slotInfo.PandoraHeaderHash)
	assert.NoError(t, err)

	err = vanguardDb.SaveLatestFinalizedSlot(32)
	assert.NoError(t, err)

	err = vanguardDb.SaveLatestFinalizedEpoch(1)
	assert.NoError(t, err)

	// Save Epoch info
	var totalConsensusInfos []*types.MinimalEpochConsensusInfo
	for i := byte(0); i < numberOfElements; i++ {
		consensusInfo := testutil.NewMinimalConsensusInfo(uint64(i))
		epochInfoV2 := consensusInfo.ConvertToEpochInfo()
		totalConsensusInfos = append(totalConsensusInfos, epochInfoV2)
		err = vanguardDb.SaveConsensusInfo(ctx, epochInfoV2)
		assert.NoError(t, err)
	}
	err = vanguardDb.SaveLatestEpoch(ctx, uint64(numberOfElements - 1))
	assert.NoError(t, err)

	return vanguardDb
}

func TestDBSetup(t *testing.T) {
	ctx := context.Background()
	vanDB := db_setup(ctx, t, 200)
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


