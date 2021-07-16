package cache

import (
	"context"
	"math/rand"
	"testing"

	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	eth "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
)

func NewPandoraShardingInfo() (*eth.PandoraShard, error) {
	retVal := new(eth.PandoraShard)
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return retVal, err
	}
	retVal.Hash = randomBytes

	if _, err := rand.Read(randomBytes); err != nil {
		return retVal, err
	}
	retVal.SealHash = randomBytes

	if _, err := rand.Read(randomBytes); err != nil {
		return retVal, err
	}
	retVal.ParentHash = randomBytes

	if _, err := rand.Read(randomBytes); err != nil {
		return retVal, err
	}
	retVal.TxHash = randomBytes

	if _, err := rand.Read(randomBytes); err != nil {
		return retVal, err
	}
	retVal.ReceiptHash = randomBytes

	if _, err := rand.Read(randomBytes); err != nil {
		return retVal, err
	}
	retVal.Signature = randomBytes

	if _, err := rand.Read(randomBytes); err != nil {
		return retVal, err
	}
	retVal.StateRoot = randomBytes

	retVal.BlockNumber = rand.Uint64()

	return retVal, nil
}

func setupShardingCache(slotNumber int) (map[uint64]*eth.PandoraShard, error) {
	slotShardMap := make(map[uint64]*eth.PandoraShard)
	for i := 0; i < slotNumber; i++ {
		tempPanShard, err := NewPandoraShardingInfo()
		if err != nil {
			return nil, err
		}
		slotShardMap[uint64(i)] = tempPanShard
	}
	return slotShardMap, nil
}

func TestVanguardShardingInfoCacheAPIs(t *testing.T) {
	vanguardCache := NewVanShardInfoCache(100)
	ctx := context.Background()
	generatedPandoraShardInfo, err := setupShardingCache(100)
	if err != nil {
		t.Error("vanguard sharding data generation failed", "error", err)
		return
	}

	for slotNumber, genInfo := range generatedPandoraShardInfo {
		err := vanguardCache.Put(ctx, slotNumber, genInfo)
		if err != nil {
			t.Error("failed while putting element vanguard cache", "slot number", slotNumber, "error", err)
		}
		receivedDataFromCache, err := vanguardCache.Get(ctx, slotNumber)
		if err != nil {
			t.Error("failed while retrieving data from the vanguard cache", "slot number", slotNumber, "error", err)
		}
		assert.DeepEqual(t, genInfo, receivedDataFromCache)
	}
}

func TestVanguardShardingInfoCacheSize(t *testing.T) {
	vanguardCache := NewVanShardInfoCache(10)
	ctx := context.Background()
	generatedPandoraShardInfo, err := setupShardingCache(100)
	if err != nil {
		t.Error("vanguard sharding data generation failed", "error", err)
		return
	}

	for slot := 0; slot < 100; slot++ {
		slotUint64 := uint64(slot)
		vanguardCache.Put(ctx, slotUint64, generatedPandoraShardInfo[slotUint64])
	}

	// Should not found slot-0 because cache size is 10
	actualHeader, err := vanguardCache.Get(ctx, 88)
	require.ErrorContains(t, "Invalid slot", err, "Should not found because cache size is 10")

	actualHeader, err = vanguardCache.Get(ctx, 90)
	require.NoError(t, err, "Should be found slot 90")
	assert.DeepEqual(t, generatedPandoraShardInfo[90], actualHeader)

}
