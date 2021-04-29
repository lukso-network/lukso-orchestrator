package cache

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"testing"
)

var expectedPanHeaders map[uint64]*types.PanBlockHeader

func setup(num int) {
	expectedPanHeaders = make(map[uint64]*types.PanBlockHeader)
	for i := 0; i < num; i++ {
		panHeader := testutil.NewPandoraHeader(uint64(i), types.Status(0))
		expectedPanHeaders[uint64(i)] = panHeader
	}
}

// Test_PandoraHeaderCache_Apis
func Test_PandoraHeaderCache_Apis(t *testing.T) {
	pc := NewPanHeaderCache()
	ctx := context.Background()
	setup(100)

	for slot, header := range expectedPanHeaders {
		pc.Put(ctx, slot, header)

		actualHeader, err := pc.Get(ctx, slot)
		require.NoError(t, err)
		assert.DeepEqual(t, header, actualHeader)
	}
}

// Test_PandoraHeaderCache_Size
func Test_PandoraHeaderCache_Size(t *testing.T) {
	maxPanHeaderCacheSize = 10
	pc := NewPanHeaderCache()
	ctx := context.Background()
	setup(100)

	for slot, header := range expectedPanHeaders {
		pc.Put(ctx, slot, header)
	}

	// Should not found slot-0 because cache size is 10
	actualHeader, err := pc.Get(ctx, 88)
	require.ErrorContains(t, "Invalid slot", err, "Should not found because cache size is 10")

	actualHeader, err = pc.Get(ctx, 90)
	require.NoError(t, err, "Should be found slot-90")
	assert.DeepEqual(t, expectedPanHeaders[90], actualHeader)
}
