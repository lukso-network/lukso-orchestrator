package utils

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"testing"
)

func TestUtils_PrepareMultiShardData(t *testing.T) {
	panHeader := testutil.NewEth1Header(1)
	vanShardInfo := testutil.NewVanguardShardInfo(1, panHeader, 0, 0)

	expectedShardInfo := new(types.MultiShardInfo)
	expectedShardInfo.SlotInfo = &types.NewSlotInfo{
		Slot:      1,
		BlockRoot: common.BytesToHash(vanShardInfo.BlockRoot),
	}

	shards := make([]*types.Shard, 1)
	blocks := make([]*types.ShardData, 1)
	blocks[0] = &types.ShardData{
		Number:     panHeader.Number.Uint64(),
		HeaderRoot: panHeader.Hash(),
	}

	shards[0] = &types.Shard{
		Id:     0,
		Blocks: blocks,
	}
	expectedShardInfo.Shards = shards
	require.DeepEqual(t, expectedShardInfo, PrepareMultiShardData(vanShardInfo, panHeader, 1, 1))
}
