package kv

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	eventTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	"math/rand"
	"testing"
)

func TestStore_ConsensusInfo_RetrieveByEpoch_FromCache(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := setupDB(t, true)
	totalConsensusInfos := make([]*eventTypes.MinimalEpochConsensusInfo, 50)
	for i := 0; i < 50; i++ {
		consensusInfo := testutil.NewMinimalConsensusInfo(uint64(i))
		totalConsensusInfos[i] = consensusInfo
		require.NoError(t, db.SaveConsensusInfo(ctx, consensusInfo))
	}

	retrievedConsensusInfo, err := db.ConsensusInfo(ctx, 49)
	require.NoError(t, err)
	assert.DeepEqual(t, totalConsensusInfos[49], retrievedConsensusInfo)
}

func TestStore_ConsensusInfo_RetrieveByEpoch_FromDB(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := setupDB(t, true)
	totalConsensusInfos := make([]*eventTypes.MinimalEpochConsensusInfo, 2000)
	for i := 0; i < 2000; i++ {
		consensusInfo := testutil.NewMinimalConsensusInfo(uint64(i))
		totalConsensusInfos[i] = consensusInfo
		require.NoError(t, db.SaveConsensusInfo(ctx, consensusInfo))
	}

	retrievedConsensusInfo, err := db.ConsensusInfo(ctx, 0)
	require.NoError(t, err)
	assert.DeepEqual(t, totalConsensusInfos[0], retrievedConsensusInfo)
}

func TestStore_SaveConsensusInfo_AlreadyExist(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := setupDB(t, true)

	consensusInfo := testutil.NewMinimalConsensusInfo(0)
	require.NoError(t, db.SaveConsensusInfo(ctx, consensusInfo))
	// again try to store same consensusInfo into cache and db
	require.NoError(t, db.SaveConsensusInfo(ctx, consensusInfo))
}

func TestStore_ConsensusInfos_RetrieveByEpoch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := setupDB(t, true)
	db.latestEpoch = 199
	totalConsensusInfos := make([]*eventTypes.MinimalEpochConsensusInfo, 200)

	for i := 0; i < 200; i++ {
		consensusInfo := testutil.NewMinimalConsensusInfo(uint64(i))
		totalConsensusInfos[i] = consensusInfo
		require.NoError(t, db.SaveConsensusInfo(ctx, consensusInfo))
	}

	retrievedConsensusInfos, err := db.ConsensusInfos(10)
	require.NoError(t, err)
	assert.DeepEqual(t, totalConsensusInfos[10:], retrievedConsensusInfos)
}

// TestStore_LatestSavedEpoch_ForFirstTime
func TestStore_LatestSavedEpoch_ForFirstTime(t *testing.T) {
	t.Parallel()
	db := setupDB(t, true)

	latestEpoch := db.LatestSavedEpoch()
	assert.Equal(t, db.latestEpoch, latestEpoch)
}

// TestStore_LatestSavedEpoch
func TestStore_SaveLatestSavedEpoch_RetrieveLatestEpoch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := setupDB(t, true)
	db.latestEpoch = uint64(1000)

	// SaveLatestEpoch is called when db is going to close
	require.NoError(t, db.SaveLatestEpoch(ctx))

	// LatestSavedEpoch is called when db is going up
	latestEpoch := db.LatestSavedEpoch()
	assert.Equal(t, db.latestEpoch, latestEpoch)
}

// TestDB_Close_Success
func TestDB_Close_Success(t *testing.T) {
	t.Parallel()
	db := setupDB(t, true)
	db.latestEpoch = uint64(1000)
	require.NoError(t, db.Close())
}

func TestStore_LatestEpoch_ClosingDB_OpeningDB(t *testing.T) {
	t.Parallel()
	db := setupDB(t, false)
	prevLatestEpoch := rand.Uint64()
	db.latestEpoch = prevLatestEpoch

	require.NoError(t, db.Close())
	restartedDB := setupDB(t, false)

	// LatestSavedEpoch is called when db is going up
	latestEpoch := restartedDB.LatestSavedEpoch()
	assert.Equal(t, prevLatestEpoch, latestEpoch)
}
