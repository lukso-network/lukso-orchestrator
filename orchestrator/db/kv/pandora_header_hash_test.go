package kv

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/shared/bytesutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"testing"
)

// TestDB_PandoraHeaderHash_Save_Retrieve
func TestDB_PandoraHeaderHash_Save_Retrieve(t *testing.T) {
	t.Parallel()
	db := setupDB(t, true)

	pandoraHeaderHashes := make([]*types.HeaderHash, 50)
	for i := 0; i < 50; i++ {
		slotByte := bytesutil.Uint64ToBytesBigEndian(uint64(i))
		hash := common.BytesToHash(slotByte)
		pandoraHeaderHashes[i] = &types.HeaderHash{HeaderHash: hash, Status: types.Pending}
		require.NoError(t, db.SavePandoraHeaderHash(uint64(i), pandoraHeaderHashes[i]))
	}

	// checking retrieval from cache
	retrievedPanHeaderHash, err := db.PandoraHeaderHash(49)
	require.NoError(t, err)
	assert.DeepEqual(t, pandoraHeaderHashes[49], retrievedPanHeaderHash)

	// checking retrieval from db
	db.panHeaderCache.Clear()
	retrievedPanHeaderHash, err = db.PandoraHeaderHash(49)
	require.NoError(t, err)
	assert.DeepEqual(t, pandoraHeaderHashes[49], retrievedPanHeaderHash)
}

// TestDB_PandoraHeaderHashes_Retrieve
func TestDB_PandoraHeaderHashes_Retrieve(t *testing.T) {
	t.Parallel()
	db := setupDB(t, true)
	pandoraHeaderHashes := make([]*types.HeaderHash, 50)
	for i := 0; i < 50; i++ {
		slotByte := bytesutil.Uint64ToBytesBigEndian(uint64(i))
		hash := common.BytesToHash(slotByte)
		pandoraHeaderHashes[i] = &types.HeaderHash{HeaderHash: hash, Status: types.Pending}
		require.NoError(t, db.SavePandoraHeaderHash(uint64(i), pandoraHeaderHashes[i]))
	}

	// checking retrieval from cache
	retrievedPanHeaderHashes, err := db.PandoraHeaderHashes(40)
	require.NoError(t, err)
	assert.DeepEqual(t, pandoraHeaderHashes[40:], retrievedPanHeaderHashes)
}

// TestDB_LatestSavedPandoraSlot
func TestDB_LatestSavedPandoraSlot(t *testing.T) {
	t.Parallel()
	db := setupDB(t, true)
	db.latestPanSlot = 1233
	require.NoError(t, db.SaveLatestPandoraSlot())

	slot := db.LatestSavedPandoraSlot()
	assert.Equal(t, db.latestPanSlot, slot)
}

// TestDB_LatestSavedPandoraHeaderHash
func TestDB_LatestSavedPandoraHeaderHash(t *testing.T) {
	t.Parallel()
	db := setupDB(t, true)
	db.latestPanHeaderHash = common.BytesToHash(bytesutil.Uint64ToBytesBigEndian(uint64(100)))
	require.NoError(t, db.SaveLatestPandoraHeaderHash())

	latestHeaderHash := db.LatestSavedPandoraHeaderHash()
	assert.Equal(t, db.latestPanHeaderHash, latestHeaderHash)
}
