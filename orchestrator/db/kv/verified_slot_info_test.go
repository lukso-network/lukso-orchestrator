package kv

import (
	"github.com/ethereum/go-ethereum/common"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	types "github.com/lukso-network/lukso-orchestrator/shared/types"
	"testing"
)

func TestStore_VerifiedSlotInfo(t *testing.T) {
	t.Parallel()
	db := setupDB(t, true)
	slotInfos := make([]*types.SlotInfo, 2001)
	for i := 0; i <= 2000; i++ {
		slotInfo := new(types.SlotInfo)
		slotInfo.VanguardBlockHash = eth1Types.EmptyRootHash
		slotInfo.PandoraHeaderHash = eth1Types.EmptyRootHash
		slotInfos[i] = slotInfo

		require.NoError(t, db.SaveVerifiedSlotInfo(uint64(i), slotInfo))
	}

	retrievedSlotInfo, err := db.VerifiedSlotInfo(0)
	require.NoError(t, err)
	assert.DeepEqual(t, slotInfos[0], retrievedSlotInfo)

	t.Run("get deleted slot info", func(t *testing.T) {
		err := db.removeSlotInfoFromVerifiedDB(50)
		require.NoError(t, err)

		_, err = db.VerifiedSlotInfo(50)
		require.ErrorContains(t, ErrValueNotFound.Error(), err)
	})

	t.Run("get out of boundary slot info", func(t *testing.T) {
		_, err := db.VerifiedSlotInfo(5000)
		require.ErrorContains(t, ErrValueNotFound.Error(), err)
	})
}

func TestStore_VerifiedSlotInfos(t *testing.T) {
	t.Parallel()
	db := setupDB(t, true)
	slotInfos := make([]*types.SlotInfo, 2001)
	for i := 0; i <= 2000; i++ {
		slotInfo := new(types.SlotInfo)
		slotInfo.VanguardBlockHash = eth1Types.EmptyRootHash
		slotInfo.PandoraHeaderHash = eth1Types.EmptyRootHash
		slotInfos[i] = slotInfo

		require.NoError(t, db.SaveVerifiedSlotInfo(uint64(i), slotInfo))
	}

	t.Run("get verified slot infos", func(t *testing.T) {
		infos, err := db.VerifiedSlotInfos(20)
		require.NoError(t, err)
		for i := 20; i <= 2000; i++ {
			assert.DeepEqual(t, slotInfos[i],infos[uint64(i)])
		}
	})

	t.Run("change slot info and get slot infos", func(t *testing.T) {
		modifiedSlotInfo := &types.SlotInfo{PandoraHeaderHash: EmptyHash, VanguardBlockHash: common.HexToHash("0x12345")}
		infos, err := db.VerifiedSlotInfos(30)
		require.NoError(t, err)

		err = db.SaveVerifiedSlotInfo(20, modifiedSlotInfo)
		require.NoError(t, err)

		for k, v := range infos {
			assert.DeepEqual(t, slotInfos[k], v)
		}

		slotInfos[20] = modifiedSlotInfo
	})

	t.Run("remove head and get slot infos", func(t *testing.T) {
		err := db.removeSlotInfoFromVerifiedDB(21)
		require.NoError(t, err)

		infos, err := db.VerifiedSlotInfos(21)
		require.NoError(t, err)

		for _, val := range infos {
			assert.NotEqual(t, slotInfos[21], val)
		}
	})
}

func TestStore_LatestSavedVerifiedSlot(t *testing.T) {
	t.Parallel()
	db := setupDB(t, true)
	slotInfos := make([]*types.SlotInfo, 2001)
	for i := 0; i <= 2000; i++ {
		slotInfo := new(types.SlotInfo)
		slotInfo.VanguardBlockHash = eth1Types.EmptyRootHash
		slotInfo.PandoraHeaderHash = eth1Types.EmptyRootHash
		slotInfos[i] = slotInfo

		require.NoError(t, db.SaveVerifiedSlotInfo(uint64(i), slotInfo))
	}

	t.Run("get latest verified slot", func(t *testing.T) {
		assert.Equal(t, uint64(2000), db.LatestSavedVerifiedSlot())
	})

	t.Run("delete and check latest verified slot", func(t *testing.T) {
		err := db.removeSlotInfoFromVerifiedDB(20)
		require.NoError(t, err)
		assert.Equal(t, uint64(2000), db.LatestSavedVerifiedSlot())

		err = db.removeSlotInfoFromVerifiedDB(2000)
		require.NoError(t, err)
		assert.Equal(t, uint64(1999), db.LatestSavedVerifiedSlot())
	})

	t.Run("change and check latest verified slot", func(t *testing.T) {
		modifiedSlotInfo := &types.SlotInfo{PandoraHeaderHash: EmptyHash, VanguardBlockHash: common.HexToHash("0x12345")}
		err := db.SaveVerifiedSlotInfo(20, modifiedSlotInfo)
		require.NoError(t, err)

		assert.Equal(t, uint64(1999), db.LatestSavedVerifiedSlot())
		err = db.SaveVerifiedSlotInfo(2000, modifiedSlotInfo)
		require.NoError(t, err)
		assert.Equal(t, uint64(2000), db.LatestSavedVerifiedSlot())
	})
}

func TestStore_GetFirstVerifiedSlotNumber(t *testing.T) {
	t.Parallel()
	db := setupDB(t, true)
	slotInfos := make([]*types.SlotInfo, 2001)
	for i := 50; i <= 2000; i++ {
		slotInfo := new(types.SlotInfo)
		slotInfo.VanguardBlockHash = eth1Types.EmptyRootHash
		slotInfo.PandoraHeaderHash = eth1Types.EmptyRootHash
		slotInfos[i] = slotInfo

		require.NoError(t, db.SaveVerifiedSlotInfo(uint64(i), slotInfo))
	}

	t.Run("get first slot number (before slot boundary)", func(t *testing.T) {
		number, err := db.GetFirstVerifiedSlotNumber(2)
		require.NoError(t, err)
		assert.Equal(t, uint64(50), number)
	})

	t.Run("get first slot number (after slot boundary)", func(t *testing.T) {
		_, err := db.GetFirstVerifiedSlotNumber(2001)
		require.ErrorContains(t, ErrValueNotFound.Error(), err)
	})

	t.Run("get first slot number (inside boundary)", func(t *testing.T) {
		number, err := db.GetFirstVerifiedSlotNumber(2000)
		require.NoError(t, err)
		assert.Equal(t, uint64(2000), number)
	})
}
