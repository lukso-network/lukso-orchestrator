package types

import (
	"github.com/ethereum/go-ethereum/common"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"math/big"
	"testing"
)

func TestTypes_DeepEqual(t *testing.T) {
	panHeader1 := &eth1Types.Header{
		Number: big.NewInt(1),
	}
	panHeader2 := &eth1Types.Header{
		Number: big.NewInt(2),
	}
	vanShardInfo := &VanguardShardInfo{
		Slot:      1,
		BlockHash: []byte{'A'},
	}

	shardInfo1 := prepareMultiShardData(vanShardInfo, panHeader1, 1, 1)
	shardInfo2 := prepareMultiShardData(vanShardInfo, panHeader2, 1, 1)
	assert.Equal(t, false, shardInfo1.DeepEqual(shardInfo2))
}

func prepareMultiShardData(
	vanShardInfo *VanguardShardInfo,
	panHeader *eth1Types.Header,
	totalExecShardCnt, shardsPerVanBlock uint64) *MultiShardInfo {
	// In future there will be multiple execution sharding chain so orchestrator needs to have support
	shards := make([]*Shard, totalExecShardCnt)
	shardData := make([]*ShardData, shardsPerVanBlock)

	// assign sharding data
	// currently vanguard block contains one sharding info per block,
	// so shardData contains only one pandora sharding header
	shardData[0] = &ShardData{
		Number:     panHeader.Number.Uint64(),
		HeaderRoot: panHeader.Hash(),
	}

	// assign shard with shard id.
	// currently we are running only one execution shard so shard id is zero
	shards[0] = &Shard{
		Id:     0,
		Blocks: shardData,
	}

	shardInfo := new(MultiShardInfo)
	shardInfo.SlotInfo = &NewSlotInfo{
		Slot:      vanShardInfo.Slot,
		BlockRoot: common.BytesToHash(vanShardInfo.BlockHash),
	}
	shardInfo.Shards = shards

	return shardInfo
}
