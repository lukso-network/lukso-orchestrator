package kv

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"testing"
)

func setupReorgDB(t *testing.T, ctx context.Context) *Store {
	db := setupDB(t, true)
	for i := 0; i < 5; i++ {
		consensusInfo := testutil.NewMinimalConsensusInfo(uint64(i))
		epochInfoV2 := consensusInfo.ConvertToEpochInfo()
		require.NoError(t, db.SaveConsensusInfo(ctx, epochInfoV2))
	}
	slotInfo := new(types.SlotInfo)
	for i := 0; i <= 100; i++ {
		slotInfo.VanguardBlockHash = common.BytesToHash([]byte{uint8(i)})
		slotInfo.PandoraHeaderHash = common.BytesToHash([]byte{uint8(i + 50)})
		require.NoError(t, db.SaveVerifiedSlotInfo(uint64(i), slotInfo))
	}

	require.NoError(t, db.SaveLatestEpoch(ctx, 4))
	require.NoError(t, db.SaveLatestVerifiedSlot(ctx, 100))
	require.NoError(t, db.SaveLatestVerifiedHeaderHash(slotInfo.PandoraHeaderHash))
	return db
}
