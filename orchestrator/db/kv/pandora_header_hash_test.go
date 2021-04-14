package kv

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/shared/bytesutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	"testing"
)

// TestDB_PandoraHeaderHash_Save_Retrieve
func TestDB_PandoraHeaderHash_Save_Retrieve(t *testing.T) {
	t.Parallel()
	db := setupDB(t, true)
	pandoraHeaderHashes := make([]common.Hash, 50)
	for i := 0; i < 50; i++ {
		slotByte := bytesutil.Uint64ToBytesBigEndian(uint64(i))
		hash := common.BytesToHash(slotByte)
		pandoraHeaderHashes[i] = hash
		require.NoError(t, db.SavePandoraHeaderHash(uint64(i), hash))
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
	pandoraHeaderHashes := make([]common.Hash, 50)
	for i := 0; i < 50; i++ {
		slotByte := bytesutil.Uint64ToBytesBigEndian(uint64(i))
		hash := common.BytesToHash(slotByte)
		pandoraHeaderHashes[i] = hash
		require.NoError(t, db.SavePandoraHeaderHash(uint64(i), hash))
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

	slot, err := db.LatestSavedPandoraSlot()
	require.NoError(t, err)
	assert.Equal(t, db.latestPanSlot, slot)
}
