package kv

import (
	"context"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	types "github.com/lukso-network/lukso-orchestrator/shared/types"
	"testing"
)

func TestStore_VerifiedSlotInfo(t *testing.T) {
	db := setupDB(t, true)
	slotInfos := createAndSaveEmptySlotInfos(t, 256, db)
	retrievedSlotInfo, err := db.VerifiedSlotInfo(0)
	require.NoError(t, err)
	assert.DeepEqual(t, slotInfos[0], retrievedSlotInfo)
}

func TestStore_RemoveRangeVerifiedInfo(t *testing.T) {
	db := setupDB(t, true)
	slotInfosLen := 128
	slotInfos := createAndSaveEmptySlotInfos(t, slotInfosLen, db)
	ctx := context.Background()
	require.NoError(t, db.SaveLatestVerifiedSlot(ctx))

	t.Run("should remove range with skip", func(t *testing.T) {
		require.NoError(t, db.RemoveRangeVerifiedInfo(4, 5))
		fetchedSlotInfos, err := db.VerifiedSlotInfos(0)
		require.NoError(t, err)
		require.Equal(t, 5, len(fetchedSlotInfos))
		require.NotEqual(t, slotInfos, fetchedSlotInfos)
	})

	require.NoError(t, db.RemoveRangeVerifiedInfo(0, 0))
	slotInfos = createAndSaveEmptySlotInfos(t, 64, db)
	require.NoError(t, db.SaveLatestVerifiedSlot(ctx))

	t.Run("should remove range without skip", func(t *testing.T) {
		fetchedSlots, err := db.VerifiedSlotInfos(0)
		require.NoError(t, err)
		require.Equal(t, len(slotInfos), len(fetchedSlots))
		require.NoError(t, db.RemoveRangeVerifiedInfo(4, 0))
		fetchedSlots, err = db.VerifiedSlotInfos(0)
		require.NoError(t, err)
		require.Equal(t, 4, len(fetchedSlots))
	})

	require.NoError(t, db.RemoveRangeVerifiedInfo(0, 0))
	slotInfos = createAndSaveEmptySlotInfos(t, 64, db)
	require.NoError(t, db.SaveLatestVerifiedSlot(ctx))

	t.Run("should remove all slots", func(t *testing.T) {
		fetchedSlots, err := db.VerifiedSlotInfos(0)
		require.NoError(t, err)
		require.Equal(t, len(slotInfos), len(fetchedSlots))
		require.NoError(t, db.RemoveRangeVerifiedInfo(0, 0))
		fetchedSlots, err = db.VerifiedSlotInfos(0)
		require.NoError(t, err)
		require.Equal(t, 1, len(fetchedSlots))
	})
}

func createAndSaveEmptySlotInfos(t *testing.T, slotsLen int, db *Store) (slotInfos []*types.SlotInfo) {
	slotInfos = make([]*types.SlotInfo, slotsLen)

	for i := 0; i < slotsLen; i++ {
		slotInfo := new(types.SlotInfo)
		slotInfo.VanguardBlockHash = eth1Types.EmptyRootHash
		slotInfo.PandoraHeaderHash = eth1Types.EmptyRootHash
		slotInfos[i] = slotInfo

		require.NoError(t, db.SaveVerifiedSlotInfo(uint64(i), slotInfo))
	}

	return
}
