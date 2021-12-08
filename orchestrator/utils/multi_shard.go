package utils

import (
	"github.com/ethereum/go-ethereum/common"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

func PrepareMultiShardData(
	vanShardInfo *types.VanguardShardInfo,
	panHeader *eth1Types.Header,
	totalExecShardCnt, shardsPerVanBlock uint64) *types.MultiShardInfo {
	// In future there will be multiple execution sharding chain so orchestrator needs to have support
	shards := make([]*types.Shard, totalExecShardCnt)
	shardData := make([]*types.ShardData, shardsPerVanBlock)

	// assign sharding data
	// currently vanguard block contains one sharding info per block,
	// so shardData contains only one pandora sharding header
	shardData[0] = &types.ShardData{
		Number:     panHeader.Number.Uint64(),
		HeaderRoot: panHeader.Hash(),
	}

	// assign shard with shard id.
	// currently we are running only one execution shard so shard id is zero
	shards[0] = &types.Shard{
		Id:     0,
		Blocks: shardData,
	}

	shardInfo := new(types.MultiShardInfo)
	shardInfo.SlotInfo = &types.NewSlotInfo{
		Slot:      vanShardInfo.Slot,
		BlockRoot: common.BytesToHash(vanShardInfo.BlockHash),
	}
	shardInfo.Shards = shards

	return shardInfo
}

// ConvertShardInfoToBlockStatus converts sharding info to blockStatus
func ConvertShardInfoToBlockStatus(
	shardInfo *types.MultiShardInfo,
	verifiedStatus types.Status, finalizedSlot uint64) (blockStatus *types.BlockStatus) {

	if shardInfo == nil {
		return nil
	}

	if len(shardInfo.Shards) == 0 {
		return nil
	}

	if len(shardInfo.Shards[0].Blocks) == 0 {
		return nil
	}

	shardData := shardInfo.Shards[0].Blocks[0]
	return &types.BlockStatus{
		Hash:          shardData.HeaderRoot,
		Status:        verifiedStatus,
		FinalizedSlot: finalizedSlot,
	}
}
