package kv

import (
	"context"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/lukso-network/lukso-orchestrator/shared/fork"
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
}

func TestStore_DeleteVerifiedSlotInfo(t *testing.T) {
	db := setupDB(t, true)
	slotInfos := make([]*types.SlotInfo, 50)

	for i := 0; i < len(slotInfos); i++ {
		slotInfo := new(types.SlotInfo)
		slotInfo.VanguardBlockHash = eth1Types.EmptyRootHash
		slotInfo.PandoraHeaderHash = eth1Types.EmptyRootHash
		slotInfos[i] = slotInfo

		require.NoError(t, db.SaveVerifiedSlotInfo(uint64(i), slotInfo))
	}

	for i := 0; i < len(slotInfos); i++ {
		presentSlot, err := db.VerifiedSlotInfo(uint64(i))
		require.NoError(t, err)
		assert.DeepEqual(t, slotInfos[i], presentSlot)
		require.NoError(t, db.DeleteVerifiedSlotInfo(uint64(i)))

		slotAfterDelete, err := db.VerifiedSlotInfo(uint64(i))
		require.NoError(t, err)
		assert.Equal(t, (*types.SlotInfo)(nil), slotAfterDelete)
	}
}

func TestStore_VerifiedSlotInfo_ForkRestriction(t *testing.T) {
	db := setupDB(t, true)
	slotInfos := make([]*types.SlotInfo, 2)

	for i := 0; i < len(slotInfos); i++ {
		slotInfo := new(types.SlotInfo)
		slotInfo.VanguardBlockHash = eth1Types.EmptyRootHash
		slotInfo.PandoraHeaderHash = eth1Types.EmptyRootHash
		slotInfos[i] = slotInfo

		require.NoError(t, db.SaveVerifiedSlotInfo(uint64(i), slotInfo))
	}

	assert.Equal(t, true, len(fork.UnsupportedForkL15PandoraProd) > 0)

	var chosenForkSlotNumber uint64

	for slotNumber, pandoraHash := range fork.UnsupportedForkL15PandoraProd {
		slotInfo := &types.SlotInfo{
			VanguardBlockHash: eth1Types.EmptyRootHash,
			PandoraHeaderHash: pandoraHash,
		}

		require.NoError(t, db.SaveVerifiedSlotInfo(slotNumber, slotInfo))
		chosenForkSlotNumber = slotNumber

		break
	}

	t.Run("should restrict slot", func(t *testing.T) {
		slotInfo, err := db.VerifiedSlotInfo(chosenForkSlotNumber)
		require.Equal(t, (*types.SlotInfo)(nil), slotInfo)
		require.ErrorContains(t, "", err)
	})

	t.Run("should not return restricted slot", func(t *testing.T) {
		db.latestVerifiedSlot = chosenForkSlotNumber
		require.NoError(t, db.SaveLatestVerifiedSlot(context.Background()))
		fetchedSlotInfos, err := db.VerifiedSlotInfos(0)
		require.NoError(t, err)
		assert.Equal(t, 2, len(fetchedSlotInfos))
	})
}
