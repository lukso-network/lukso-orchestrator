package kv

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"testing"
)

func TestStore_ReorgDB(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := setupReorgDB(t, ctx)
	defer db.ClearDB()
	reorgEpochInfo := &types.MinimalEpochConsensusInfoV2{
		Epoch: 3,
		ReorgInfo: &types.Reorg{
			VanParentHash: []byte{uint8(50)},
			PanParentHash: []byte{uint8(100)},
		},
	}
	require.NoError(t, db.RevertConsensusInfo(reorgEpochInfo))
	expectedSlotInfo := (*types.SlotInfo)(nil)
	for i := 51; i < 100; i++ {
		actualSlotInfo, err := db.VerifiedSlotInfo(uint64(i))
		require.NoError(t, err)
		require.DeepEqual(t, expectedSlotInfo, actualSlotInfo)
	}

	for i := 0; i <= 50; i++ {
		expectedSlotInfo := new(types.SlotInfo)
		expectedSlotInfo.VanguardBlockHash = common.BytesToHash([]byte{uint8(i)})
		expectedSlotInfo.PandoraHeaderHash = common.BytesToHash([]byte{uint8(i + 50)})

		actualSlotInfo, err := db.VerifiedSlotInfo(uint64(i))
		require.NoError(t, err)
		require.DeepEqual(t, expectedSlotInfo, actualSlotInfo)
	}
}

func setupReorgDB(t *testing.T, ctx context.Context) *Store {
	db := setupDB(t, true)
	for i := 0; i < 5; i++ {
		consensusInfo := testutil.NewMinimalConsensusInfo(uint64(i))
		epochInfoV2 := consensusInfo.ConvertToEpochInfo()
		require.NoError(t, db.SaveConsensusInfo(ctx, epochInfoV2))
	}
	for i := 0; i <= 100; i++ {
		slotInfo := new(types.SlotInfo)
		slotInfo.VanguardBlockHash = common.BytesToHash([]byte{uint8(i)})
		slotInfo.PandoraHeaderHash = common.BytesToHash([]byte{uint8(i + 50)})
		require.NoError(t, db.SaveVerifiedSlotInfo(uint64(i), slotInfo))
	}

	require.NoError(t, db.SaveLatestEpoch(ctx))
	require.NoError(t, db.SaveLatestVerifiedSlot(ctx, 100))
	require.NoError(t, db.SaveLatestVerifiedHeaderHash())
	return db
}
