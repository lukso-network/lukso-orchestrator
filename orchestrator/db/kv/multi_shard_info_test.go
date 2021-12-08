package kv

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"testing"
)

func TestStore_SaveVerifiedShardInfo(t *testing.T) {
	db := setupDB(t, true)
	shardInfos := NewMultiShardInfos(11)
	for i := 1; i <= 11; i++ {
		require.NoError(t, db.SaveVerifiedShardInfo(uint64(i), shardInfos[i]))
	}

	for i := 1; i <= 11; i++ {
		shardInfo, err := db.VerifiedShardInfo(uint64(i))
		require.NoError(t, err)
		require.DeepEqual(t, shardInfo, shardInfos[i])
	}
}

func NewMultiShardInfos(size int) []*types.MultiShardInfo {
	shardInfos := make([]*types.MultiShardInfo, size+1)
	for i := 1; i <= size; i++ {
		slotInfo := &types.NewSlotInfo{
			Slot:      uint64(i),
			BlockRoot: common.BytesToHash([]byte{uint8(i)}),
		}

		shards := make([]*types.Shard, 2)
		blocks := make([]*types.ShardData, 2)

		blocks[0] = &types.ShardData{
			Number:     1,
			HeaderRoot: common.BytesToHash([]byte{'A' + uint8(i)}),
		}

		blocks[1] = &types.ShardData{
			Number:     2,
			HeaderRoot: common.BytesToHash([]byte{'B' + uint8(i)}),
		}

		shards[0] = &types.Shard{
			Id:     0,
			Blocks: blocks,
		}

		shards[1] = &types.Shard{
			Id:     1,
			Blocks: blocks,
		}

		shardInfos[i] = &types.MultiShardInfo{
			SlotInfo: slotInfo,
			Shards:   shards,
		}
	}

	return shardInfos
}
