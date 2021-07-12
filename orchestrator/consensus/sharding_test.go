package consensus

import (
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	"github.com/lukso-network/lukso-orchestrator/shared/types"

	"testing"
)

func generateHeaderHash() (*types.HeaderHash, error) {
	headerHash := new(types.HeaderHash)
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return headerHash, err
	}
	headerHash.HeaderHash = common.BytesToHash(randomBytes)

	if _, err := rand.Read(randomBytes); err != nil {
		return headerHash, err
	}
	headerHash.PandoraShardHash = common.BytesToHash(randomBytes)

	if _, err := rand.Read(randomBytes); err != nil {
		return headerHash, err
	}
	headerHash.Signature = randomBytes

	return headerHash, nil
}

func TestCompareShardingInfo(t *testing.T) {
	var pandoraHeaderHash, vanguardHeaderHash *types.HeaderHash
	var err error
	t.Run("empty headerHash", func(t *testing.T) {
		require.Equal(t, true, CompareShardingInfo(pandoraHeaderHash, vanguardHeaderHash))
	})

	t.Run("headerHash with status and hash", func(t *testing.T) {
		pandoraHeaderHash = &types.HeaderHash{
			HeaderHash: common.HexToHash("0x123456"),
			Status:     types.Pending,
		}
		vanguardHeaderHash = &types.HeaderHash{
			HeaderHash: common.HexToHash("0x123456"),
			Status:     types.Pending,
		}
		require.Equal(t, true, CompareShardingInfo(pandoraHeaderHash, vanguardHeaderHash))
	})

	if pandoraHeaderHash, err = generateHeaderHash(); err != nil {
		t.Error("error while generating pandoraHeaderHash", err)
	}
	if vanguardHeaderHash, err = generateHeaderHash(); err != nil {
		t.Error("error while generating vanguardHeaderHash", err)
	}
	t.Run("headerHash with different value", func(t *testing.T) {
		require.Equal(t, false, CompareShardingInfo(pandoraHeaderHash, vanguardHeaderHash))
	})

	pandoraHeaderHash = vanguardHeaderHash
	t.Run("headerHash with same value", func(t *testing.T) {
		require.Equal(t, true, CompareShardingInfo(pandoraHeaderHash, vanguardHeaderHash))
	})
}
