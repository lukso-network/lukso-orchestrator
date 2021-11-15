package kv

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	types "github.com/lukso-network/lukso-orchestrator/shared/types"
	"testing"
)

func TestStore_VerifiedSlotInfo(t *testing.T) {
	db := setupDB(t, true)
	slotInfosLen := 32
	slotInfos := createAndSaveEmptySlotInfos(t, slotInfosLen, db)
	retrievedSlotInfo, err := db.VerifiedSlotInfo(0)
	require.NoError(t, err)
	assert.DeepEqual(t, slotInfos[0], retrievedSlotInfo)
}

func TestStore_VerifiedSlotInfos(t *testing.T) {
	db := setupDB(t, true)
	slotInfosLen := 64
	slotInfos := createAndSaveEmptySlotInfos(t, slotInfosLen, db)
	require.NoError(t, db.SaveLatestVerifiedSlot(context.Background(), uint64(slotInfosLen)-uint64(1)))
	retrievedSlotInfos, err := db.VerifiedSlotInfos(0)
	require.NoError(t, err)
	require.Equal(t, slotInfosLen, len(retrievedSlotInfos))
	assert.DeepEqual(t, slotInfos[0], retrievedSlotInfos[0])
}

func TestStore_LatestVerifiedSuite(t *testing.T) {
	db := setupDB(t, true)
	ctx := context.Background()
	createAndSaveEmptySlotInfos(t, 64, db)
	customSlotInfoHeight := uint64(64)
	customSlotInfo := &types.SlotInfo{
		VanguardBlockHash: common.HexToHash("0x6f701e4e8b260f38a43cdc0d97cfdc7f0cd33f58ef26bbc6c327ac87d76304d2"),
		PandoraHeaderHash: common.HexToHash("0x0846da512db0a6888a59aa5f7235b741e36a9dcacc9dad33ee2a228878aefa74"),
	}

	require.NoError(t, db.SaveVerifiedSlotInfo(customSlotInfoHeight, customSlotInfo))
	require.NoError(t, db.SaveLatestVerifiedSlot(ctx, customSlotInfoHeight))
	require.Equal(t, customSlotInfoHeight, db.LatestSavedVerifiedSlot())
	// this is failing, implementation is wrong
	require.NoError(t, db.SaveLatestVerifiedHeaderHash(customSlotInfo.PandoraHeaderHash))
	require.Equal(t, customSlotInfo.PandoraHeaderHash, db.LatestVerifiedHeaderHash())
}

func TestStore_FindVerifiedSlotNumber(t *testing.T) {
	db := setupDB(t, true)
	ctx := context.Background()
	slotInfos := createAndSaveEmptySlotInfos(t, 64, db)
	customSlotInfoHeight := uint64(64)
	customSlotInfo := &types.SlotInfo{
		VanguardBlockHash: common.HexToHash("0x6f701e4e8b260f38a43cdc0d97cfdc7f0cd33f58ef26bbc6c327ac87d76304d2"),
		PandoraHeaderHash: common.HexToHash("0x0846da512db0a6888a59aa5f7235b741e36a9dcacc9dad33ee2a228878aefa74"),
	}
	require.NoError(t, db.SaveVerifiedSlotInfo(customSlotInfoHeight, customSlotInfo))
	require.NoError(t, db.SaveLatestVerifiedSlot(ctx, customSlotInfoHeight))
	// This is not working properly in my opinion, this by design should be highestVerifiedSlot, not latestVerifiedSlot
	require.Equal(t, customSlotInfoHeight, db.LatestSavedVerifiedSlot())

	t.Run("should find verified slot with matching slot", func(t *testing.T) {
		slotNumber := db.FindVerifiedSlotNumber(customSlotInfo, customSlotInfoHeight)
		require.Equal(t, customSlotInfoHeight, slotNumber)
		require.NotEqual(t, slotInfos[customSlotInfoHeight-1], customSlotInfo)
	})

	t.Run("should find verified slot with slot higher than present in db", func(t *testing.T) {
		slotNumber := db.FindVerifiedSlotNumber(customSlotInfo, customSlotInfoHeight+50)
		require.Equal(t, customSlotInfoHeight, slotNumber)
		require.NotEqual(t, slotInfos[customSlotInfoHeight-1], customSlotInfo)
	})

	t.Run("should not find verified slot", func(t *testing.T) {
		slotInfo := db.FindVerifiedSlotNumber(customSlotInfo, 4)
		require.Equal(t, uint64(0), slotInfo)
	})
}

func TestStore_RemoveRangeVerifiedInfo(t *testing.T) {
	db := setupDB(t, true)
	slotInfosLen := 128
	slotInfos := createAndSaveEmptySlotInfos(t, slotInfosLen, db)
	ctx := context.Background()
	require.NoError(t, db.SaveLatestVerifiedSlot(ctx, uint64(slotInfosLen-1)))

	t.Run("should remove range with skip", func(t *testing.T) {
		require.NoError(t, db.RemoveRangeVerifiedInfo(4, 5))
		fetchedSlotInfos, err := db.VerifiedSlotInfos(0)
		require.NoError(t, err)
		require.Equal(t, 5, len(fetchedSlotInfos))
		require.NotEqual(t, slotInfos, fetchedSlotInfos)
	})

	require.NoError(t, db.RemoveRangeVerifiedInfo(0, 0))
	slotInfos = createAndSaveEmptySlotInfos(t, 64, db)
	require.NoError(t, db.SaveLatestVerifiedSlot(ctx, 63))

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
	require.NoError(t, db.SaveLatestVerifiedSlot(ctx, 63))

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
