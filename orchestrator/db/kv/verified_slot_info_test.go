package kv

import (
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	types "github.com/lukso-network/lukso-orchestrator/shared/types"
	"testing"
)

func TestStore_VerifiedSlotInfo(t *testing.T) {
	t.Parallel()
	db := setupDB(t, true)
	slotInfos := make([]*types.SlotInfo, 2001)
	for i := 0; i <= 2000; i++ {
		slotInfo := new(types.SlotInfo)
		slotInfo.VanguardBlockHash = eth1Types.EmptyRootHash
		slotInfo.PandoraHeaderHash = eth1Types.EmptyRootHash
		slotInfos[i] = slotInfo

		require.NoError(t, db.SaveVerifiedSlotInfo(uint64(i), slotInfo))
	}

	retrievedSlotInfo, err := db.VerifiedSlotInfo(0)
	require.NoError(t, err)
	assert.DeepEqual(t, slotInfos[0], retrievedSlotInfo)
}
