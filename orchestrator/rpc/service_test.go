package rpc

import (
	"context"
	ethRpc "github.com/ethereum/go-ethereum/rpc"
	testDB "github.com/lukso-network/lukso-orchestrator/orchestrator/db/testing"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain"
	"github.com/lukso-network/lukso-orchestrator/shared/cmd"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"testing"
)

func setup(t *testing.T) (*Config, error) {
	orchestratorDB := testDB.SetupDB(t)
	dialRPCClient := func(endpoint string) (*ethRpc.Client, error) {
		client, err := ethRpc.Dial(endpoint)
		if err != nil {
			return nil, err
		}
		return client, nil
	}
	namespace := "van"
	consensusInfoFeed, err := vanguardchain.NewService(
		context.Background(),
		cmd.DefaultVanguardRPCEndpoint,
		cmd.DefaultVanguardGRPCEndpoint,
		namespace,
		orchestratorDB,
		dialRPCClient,
	)
	if err != nil {
		return nil, err
	}

	return &Config{
		ConsensusInfoFeed: consensusInfoFeed,
		ConsensusInfoDB:   orchestratorDB,
		IPCPath:           cmd.DefaultIpcPath,
		HTTPEnable:        true,
		HTTPHost:          cmd.DefaultHTTPHost,
		HTTPPort:          9874,
		WSEnable:          true,
		WSHost:            cmd.DefaultWSHost,
		WSPort:            9875,
	}, nil
}

// TODO- Need to implement more integration test cases
// TestServerStart_Success
func TestServerStart_Success(t *testing.T) {
	hook := logTest.NewGlobal()
	ctx := context.Background()
	config, err := setup(t)
	require.NoError(t, err)

	rpcService, err := NewService(ctx, config)
	if err != nil {
		t.Fatalf("failed to create protocol stack: %v", err)
	}

	// Ensure that a node can be successfully started, but only once
	assert.NoError(t, rpcService.startRPC())
	require.LogsContain(t, hook, "IPC endpoint opened", "IPC server not started")
	require.LogsContain(t, hook, "HTTP server started", "Http server not started")
	require.LogsContain(t, hook, "WebSocket enabled", "Web socket server not started")

	hook.Reset()
	assert.NoError(t, rpcService.Stop())
}
