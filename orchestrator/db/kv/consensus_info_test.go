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

func TestStore_ConsensusInfo_Remove(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := setupDB(t, true)
	for i := 0; i < 50; i++ {
		consensusInfo := testutil.NewMinimalConsensusInfo(uint64(i))
		require.NoError(t, db.SaveConsensusInfo(ctx, consensusInfo))
	}

	err := db.removeConsensusInfoDb(40)
	require.NoError(t, err)

	_, err = db.ConsensusInfo(ctx, 40)
	require.ErrorContains(t, ErrValueNotFound.Error(), err)
}

func TestStore_GetLatestConsensusInfo(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := setupDB(t, true)

	for i := 0; i < 50; i++ {
		consensusInfo := testutil.NewMinimalConsensusInfo(uint64(i))
		require.NoError(t, db.SaveConsensusInfo(ctx, consensusInfo))
	}

	err := db.removeConsensusInfoDb(40)
	require.NoError(t, err)

	info, err := db.GetLatestConsensusInfo()
	require.NoError(t, err)
	assert.Equal(t,  uint64(49), info.Epoch)

	for i := 50; i < 100; i++ {
		consensusInfo := testutil.NewMinimalConsensusInfo(uint64(i))
		require.NoError(t, db.SaveConsensusInfo(ctx, consensusInfo))
	}

	info, err = db.GetLatestConsensusInfo()
	require.NoError(t, err)
	assert.Equal(t, uint64(99), info.Epoch)

	err = db.removeConsensusInfoDb(99)
	require.NoError(t, err)

	info, err = db.GetLatestConsensusInfo()
	require.NoError(t, err)
	assert.Equal(t, uint64(98), info.Epoch)

	for i := 0; i < 100; i++ {
		err = db.removeConsensusInfoDb(uint64(i))
		require.NoError(t, err)
	}
	info, err = db.GetLatestConsensusInfo()
	require.ErrorContains(t, ErrValueNotFound.Error(), err)
}

func TestStore_ConsensusInfos(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := setupDB(t, true)
	generatedConsensusInfo := make([]*eventTypes.MinimalEpochConsensusInfo, 50)

	for i := 0; i < 30; i++ {
		consensusInfo := testutil.NewMinimalConsensusInfo(uint64(i))
		generatedConsensusInfo[consensusInfo.Epoch] = consensusInfo
		require.NoError(t, db.SaveConsensusInfo(ctx, consensusInfo))
	}

	t.Run("search from an epoch", func(t *testing.T) {
		infos, err := db.ConsensusInfos(10)
		require.NoError(t, err)
		for _, info := range infos {
			assert.DeepEqual(t, generatedConsensusInfo[info.Epoch], info)
		}

		consensusInfo := testutil.NewMinimalConsensusInfo(uint64(40))
		generatedConsensusInfo[consensusInfo.Epoch] = consensusInfo
		require.NoError(t, db.SaveConsensusInfo(ctx, consensusInfo))

		infos, err = db.ConsensusInfos(10)
		require.NoError(t, err)
		for _, info := range infos {
			t.Log(info.Epoch)
			assert.DeepEqual(t, generatedConsensusInfo[info.Epoch], info)
		}
	})

	t.Run("search from out of boundary", func(t *testing.T) {
		infos, err := db.ConsensusInfos(200)
		require.NoError(t, err)
		assert.Equal(t, 0, len(infos))
	})

	t.Run("remove and search from the next", func(t *testing.T) {
		err := db.removeConsensusInfoDb(2)
		require.NoError(t, err)

		infos, err := db.ConsensusInfos(2)
		require.NoError(t, err)
		// it should get from the next available key
		assert.DeepEqual(t, generatedConsensusInfo[3], infos[0])
	})
}

func TestStore_ConsensusInfo_RetrieveByEpoch_FromDB(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := setupDB(t, true)
	totalConsensusInfos := make([]*eventTypes.MinimalEpochConsensusInfo, 2001)
	for i := 1; i <= 2000; i++ {
		consensusInfo := testutil.NewMinimalConsensusInfo(uint64(i))
		totalConsensusInfos[i] = consensusInfo
		require.NoError(t, db.SaveConsensusInfo(ctx, consensusInfo))
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
	require.NoError(t, db.SaveConsensusInfo(ctx, consensusInfo))
	// again try to store same consensusInfo into cache and db
	require.NoError(t, db.SaveConsensusInfo(ctx, consensusInfo))
}

func TestStore_ConsensusInfos_RetrieveByEpoch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := setupDB(t, true)
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
