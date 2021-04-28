package kv

import (
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"testing"
)

// Test_SlotFromExtraData
func Test_SlotFromExtraData(t *testing.T) {
	expectedSlot := uint64(45)
	pandHeader := testutil.NewPandoraHeader(expectedSlot, types.Status(0))
	actualSlot, err := SlotFromExtraData(pandHeader)
	require.NoError(t, err)
	assert.Equal(t, expectedSlot, actualSlot)
}
