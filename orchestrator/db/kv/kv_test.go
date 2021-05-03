package kv

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	"testing"
)

// setupDB instantiates and returns a Store instance.
func setupDB(t testing.TB, useTempDir bool) *Store {
	var dbPath string
	if !useTempDir {
		dbPath = "./test-data/" + OrchestratorNodeDbDirName
	} else {
		dbPath = t.TempDir()
	}
	db, err := NewKVStore(context.Background(), dbPath, &Config{})
	require.NoError(t, err, "Failed to instantiate DB")
	t.Cleanup(func() {
		require.NoError(t, db.Close(), "Failed to close database")
	})
	return db
}
