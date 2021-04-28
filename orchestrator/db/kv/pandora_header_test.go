package kv

import (
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"testing"
)

func Test_PanHeader_Save_Retrieve(t *testing.T) {
	t.Parallel()
	db := setupDB(t, true)
	pandoraHeaders := make([]*types.PanBlockHeader, 50)
	for i := 0; i < 50; i++ {
		pandoraHeader := testutil.NewPandoraHeader(uint64(i), types.Status(0))
		pandoraHeaders[i] = pandoraHeader
		require.NoError(t, db.SavePanHeader(pandoraHeader))
	}

	// checking retrieval from cache
	retrievedPanHeader, err := db.PanHeader(49)
	require.NoError(t, err)
	assert.DeepEqual(t, pandoraHeaders[49], retrievedPanHeader)

	// checking retrieval from db
	db.panHeaderCache.Clear()
	retrievedPanHeader, err = db.PanHeader(49)
	require.NoError(t, err)
	assert.DeepEqual(t, pandoraHeaders[49], retrievedPanHeader)
}
