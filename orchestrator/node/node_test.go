package node

import (
	"flag"
	"github.com/lukso-network/lukso-orchestrator/shared/cmd"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
	"testing"
)

// Test that beacon chain node can register all services and close.
func Test_Node_RegisterServices(t *testing.T) {
	hook := logTest.NewGlobal()
	tmp := filepath.Join(t.TempDir(), "datadirtest")

	app := cli.App{}
	set := flag.NewFlagSet("test", 0)
	set.String("datadir", tmp, "Data directory for storing consensus metadata and block headers")

	context := cli.NewContext(&app, set, nil)
	node, err := New(context)
	require.NoError(t, err)

	node.Close()
	require.LogsContain(t, hook, "Stopping orchestrator node")
	require.NoError(t, os.RemoveAll(tmp))
}

// TestClearDB tests clearing the database
func Test_ClearDB(t *testing.T) {
	hook := logTest.NewGlobal()
	tmp := filepath.Join(t.TempDir(), "datadirtest")

	app := cli.App{}
	set := flag.NewFlagSet("test", 0)
	set.String("datadir", tmp, "node data directory")
	set.Bool(cmd.ForceClearDB.Name, true, "force clear db")

	context := cli.NewContext(&app, set, nil)
	_, err := New(context)
	require.NoError(t, err)

	require.LogsContain(t, hook, "Removing database")
	require.NoError(t, os.RemoveAll(tmp))
}
