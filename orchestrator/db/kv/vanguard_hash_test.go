package kv

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/shared/bytesutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"testing"
)

func TestStore_LatestSavedVanguardHeaderHash(t *testing.T) {
	t.Parallel()
	db := setupDB(t, true)

	vanguardHeaderHashes := make([]*types.HeaderHash, 50)

	for i := 0; i < 50; i++ {
		slotByte := bytesutil.Uint64ToBytesBigEndian(uint64(i))
		hash := common.BytesToHash(slotByte)
		vanguardHeaderHashes[i] = &types.HeaderHash{HeaderHash: hash, Status: types.Pending}
		require.NoError(t, db.SaveVanguardHeaderHash(uint64(i), vanguardHeaderHashes[i]))
	}

	retrievedVanHeaderHash, err := db.VanguardHeaderHash(49)
	require.NoError(t, err)
	assert.DeepEqual(t, vanguardHeaderHashes[49], retrievedVanHeaderHash)

	latestVanHeaderHash := db.LatestSavedVanguardHeaderHash()
	require.DeepEqual(t, vanguardHeaderHashes[len(vanguardHeaderHashes)-1].HeaderHash, latestVanHeaderHash)

	db.vanHeaderCache.Clear()
	retrievedVanHeaderHash, err = db.VanguardHeaderHash(49)
	require.NoError(t, err)
	assert.DeepEqual(t, vanguardHeaderHashes[49], retrievedVanHeaderHash)
}

func TestStore_LatestSavedVanguardSlot(t *testing.T) {
	t.Parallel()
	db := setupDB(t, true)
	db.latestVanSlot = 1233
	require.NoError(t, db.SaveLatestVanguardSlot())

	slot := db.LatestSavedVanguardSlot()
	assert.Equal(t, db.latestVanSlot, slot)
}

func TestStore_VanguardHeaderHash(t *testing.T) {
	t.Parallel()
	db := setupDB(t, true)

	slot := 1
	slotByte := bytesutil.Uint64ToBytesBigEndian(uint64(slot))
	hash := common.BytesToHash(slotByte)
	headerHash := &types.HeaderHash{HeaderHash: hash, Status: types.Pending}
	require.NoError(t, db.SaveVanguardHeaderHash(uint64(slot), headerHash))

	// checking retrieval from cache
	vanHeaderHash, err := db.VanguardHeaderHash(uint64(slot))
	require.NoError(t, err)
	assert.DeepEqual(t, headerHash, vanHeaderHash)
}

func TestStore_VanguardHeaderHashes(t *testing.T) {
	t.Parallel()
	db := setupDB(t, true)
	vanguardHeaderHashes := make([]*types.HeaderHash, 50)
	for i := 0; i < 50; i++ {
		slotByte := bytesutil.Uint64ToBytesBigEndian(uint64(i))
		hash := common.BytesToHash(slotByte)
		vanguardHeaderHashes[i] = &types.HeaderHash{HeaderHash: hash, Status: types.Pending}
		require.NoError(t, db.SaveVanguardHeaderHash(uint64(i), vanguardHeaderHashes[i]))
	}

	// checking retrieval from cache
	retrievedVanHeaderHashes, err := db.VanguardHeaderHashes(40, 50)
	require.NoError(t, err)
	assert.DeepEqual(t, vanguardHeaderHashes[40:], retrievedVanHeaderHashes)
}

func TestStore_SaveVanguardHeaderHash(t *testing.T) {
	t.Parallel()
	db := setupDB(t, true)

	slot := 1
	slotByte := bytesutil.Uint64ToBytesBigEndian(uint64(slot))
	hash := common.BytesToHash(slotByte)
	headerHash := &types.HeaderHash{HeaderHash: hash, Status: types.Pending}
	require.NoError(t, db.SaveVanguardHeaderHash(uint64(slot), headerHash))
}

func TestStore_SaveLatestVanguardSlot(t *testing.T) {
	t.Parallel()
	db := setupDB(t, true)
	db.latestVanSlot = 1233
	require.NoError(t, db.SaveLatestVanguardSlot())
}

func TestStore_SaveLatestVanguardHeaderHash(t *testing.T) {
	t.Parallel()
	db := setupDB(t, true)
	db.latestVanHash = common.BytesToHash(bytesutil.Uint64ToBytesBigEndian(uint64(100)))
	require.NoError(t, db.SaveLatestVanguardHeaderHash())

	latestHeaderHash := db.LatestSavedVanguardHeaderHash()
	assert.Equal(t, db.latestVanHash, latestHeaderHash)
}
