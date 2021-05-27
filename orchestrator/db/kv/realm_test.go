package kv

import (
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	"testing"
)

func TestStore_SaveAndRetrieveLatestVerifiedRealmSlot(t *testing.T) {
	t.Parallel()
	db := setupDB(t, true)
	expectedSlot := uint64(1)
	require.NoError(t, db.SaveLatestVerifiedRealmSlot(expectedSlot))

	slot := db.LatestVerifiedRealmSlot()
	assert.Equal(t, expectedSlot, slot)
}
