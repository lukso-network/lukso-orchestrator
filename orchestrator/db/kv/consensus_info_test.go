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
	totalConsensusInfos := make([]*eventTypes.MinimalEpochConsensusInfoV2, 50)
	for i := 0; i < 50; i++ {
		consensusInfo := testutil.NewMinimalConsensusInfo(uint64(i))
		epochInfoV2 := consensusInfo.ConvertToEpochInfoV2()
		totalConsensusInfos[i] = epochInfoV2
		require.NoError(t, db.SaveConsensusInfo(ctx, epochInfoV2))
	}

	retrievedConsensusInfo, err := db.ConsensusInfo(ctx, 49)
	require.NoError(t, err)
	assert.DeepEqual(t, totalConsensusInfos[49], retrievedConsensusInfo)
}

func TestStore_ConsensusInfo_RetrieveByEpoch_FromDB(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := setupDB(t, true)
	totalConsensusInfos := make([]*eventTypes.MinimalEpochConsensusInfoV2, 2001)
	for i := 1; i <= 2000; i++ {
		consensusInfo := testutil.NewMinimalConsensusInfo(uint64(i))
		epochInfoV2 := consensusInfo.ConvertToEpochInfoV2()
		totalConsensusInfos[i] = epochInfoV2
		require.NoError(t, db.SaveConsensusInfo(ctx, epochInfoV2))
	}

	retrievedConsensusInfo, err := db.ConsensusInfo(ctx, 1)
	require.NoError(t, err)
	assert.DeepEqual(t, totalConsensusInfos[1], retrievedConsensusInfo)
}

func TestStore_SaveConsensusInfo_AlreadyExist(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := setupDB(t, true)

	consensusInfo := testutil.NewMinimalConsensusInfo(0)
	epochInfoV2 := consensusInfo.ConvertToEpochInfoV2()
	require.NoError(t, db.SaveConsensusInfo(ctx, epochInfoV2))
	// again try to store same consensusInfo into cache and db
	require.NoError(t, db.SaveConsensusInfo(ctx, epochInfoV2))
}

func TestStore_ConsensusInfos_RetrieveByEpoch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := setupDB(t, true)
	db.latestEpoch = 199
	db.SaveLatestEpoch(ctx)
	totalConsensusInfos := make([]*eventTypes.MinimalEpochConsensusInfoV2, 200)

	for i := 0; i < 200; i++ {
		consensusInfo := testutil.NewMinimalConsensusInfo(uint64(i))
		epochInfoV2 := consensusInfo.ConvertToEpochInfoV2()
		totalConsensusInfos[i] = epochInfoV2
		require.NoError(t, db.SaveConsensusInfo(ctx, epochInfoV2))
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

	if db.isRunning {
		require.NoError(t, db.Close())
	}
}

func TestStore_LatestEpoch_ClosingDB_OpeningDB(t *testing.T) {
	t.Parallel()
	db := setupDB(t, false)
	prevLatestEpoch := rand.Uint64()
	db.latestEpoch = prevLatestEpoch

	if !db.isRunning {
		t.SkipNow()

		return
	}

	require.NoError(t, db.Close())
	restartedDB := setupDB(t, false)

	// LatestSavedEpoch is called when db is going up
	latestEpoch := restartedDB.LatestSavedEpoch()
	assert.Equal(t, prevLatestEpoch, latestEpoch)
}
