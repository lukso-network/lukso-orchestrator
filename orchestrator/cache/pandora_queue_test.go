package cache

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/utils"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	eth2Types "github.com/prysmaticlabs/eth2-types"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestPandoraCache_RemoveByTime(t *testing.T) {
	pc := NewPandoraCache(1024, 0, 6, utils.NewStack())
	headerInfos, _ := testutil.GetHeaderInfosAndShardInfos(1, 25)
	for i := 0; i < 25; i++ {
		slot := uint64(i + 1)
		queueData := &PandoraCacheData{
			panHeader:      headerInfos[i].Header,
			entryTimestamp: utils.SlotStartTime(pc.genesisStartTime, eth2Types.Slot(slot), pc.secondsPerSlot),
		}
		pc.stack.Push(headerInfos[i].Header.Hash().Bytes())
		pc.cache.Add(slot, queueData)
	}

	removedHeaders := pc.RemoveByTime(time.Now())
	assert.Equal(t, 25, len(removedHeaders))
}

func TestPandoraCache_Put(t *testing.T) {
	t.Parallel()

	t.Run("insert empty data into cache", func(t *testing.T) {
		tempPanCache := NewPandoraCache(1, 0, 6, utils.NewStack())
		err := tempPanCache.Put(0, nil)
		require.Error(t, err)
	})
	t.Run("insert one data into cache", func(t *testing.T) {
		tempPanCache := NewPandoraCache(1, 0, 6, utils.NewStack())
		err := tempPanCache.Put(0, &PanCacheInsertParams{CurrentVerifiedHeader: new(types.Header)})
		require.NoError(t, err)
	})
	t.Run("insert more element than the size", func(t *testing.T) {
		tempPanCache := NewPandoraCache(1, 0, 6, utils.NewStack())
		err := tempPanCache.Put(0, &PanCacheInsertParams{CurrentVerifiedHeader: new(types.Header)})
		require.NoError(t, err)
		for i := 0; i < 10; i++ {
			err := tempPanCache.Put(uint64(i), &PanCacheInsertParams{CurrentVerifiedHeader: getTestHeader(int64(i))})
			require.NoError(t, err)
		}
		require.Nil(t, tempPanCache.Get(0))
		require.NotNil(t, tempPanCache.Get(9))
	})
	t.Run("check if data is updated", func(t *testing.T) {
		tempPanCache := NewPandoraCache(1, 0, 6, utils.NewStack())
		tempHeader := getTestHeader(0)
		err := tempPanCache.Put(0, &PanCacheInsertParams{CurrentVerifiedHeader: tempHeader})
		require.NoError(t, err)
		assert.DeepEqual(t, tempHeader, tempPanCache.Get(0).GetPanHeader())

		tempHeader = getTestHeader(23)
		err = tempPanCache.Put(0, &PanCacheInsertParams{CurrentVerifiedHeader: tempHeader})
		require.NoError(t, err)
		assert.DeepEqual(t, tempHeader, tempPanCache.Get(0).GetPanHeader())
	})
	t.Run("current header is nil", func(t *testing.T) {
		tempPanCache := NewPandoraCache(1, 0, 6, utils.NewStack())
		err := tempPanCache.Put(0, &PanCacheInsertParams{IsSyncing: true})
		require.Error(t, err)
	})
	t.Run("set sync status", func(t *testing.T) {
		tempPanCache := NewPandoraCache(1, 0, 6, utils.NewStack())
		tempHeader := getTestHeader(0)
		currentTime := time.Now()
		err := tempPanCache.Put(0, &PanCacheInsertParams{CurrentVerifiedHeader: tempHeader, IsSyncing: true})
		require.NoError(t, err)
		require.GreaterOrEqual(t, tempPanCache.Get(0).entryTimestamp.Unix(), currentTime.Unix())
	})
}
