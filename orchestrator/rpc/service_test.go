package rpc

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/cache"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/consensus"
	testDB "github.com/lukso-network/lukso-orchestrator/orchestrator/db/testing"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/utils"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain"
	"github.com/lukso-network/lukso-orchestrator/shared/cmd"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"testing"
)

func setup(t *testing.T) (*Config, error) {
	orchestratorDB := testDB.SetupDB(t)
	consensusInfoFeed, err := vanguardchain.NewService(
		context.Background(),
		cmd.DefaultVanguardGRPCEndpoint,
		orchestratorDB,
	)
	if err != nil {
		return nil, err
	}

	consensusSvr := consensus.New(
		context.Background(),
		&consensus.Config{
			VerifiedShardInfoDB: orchestratorDB,
			PanHeaderCache:      cache.NewPandoraCache(1<<10, 0, 6, utils.NewStack()),
			VanShardCache:       cache.NewVanguardCache(1<<10, 0, 6, utils.NewStack()),
		})

	return &Config{
		ConsensusInfoFeed:    consensusInfoFeed,
		VerifiedSlotInfoFeed: consensusSvr,
		Db:                   orchestratorDB,
		IPCPath:              cmd.DefaultIpcPath,
		HTTPEnable:           true,
		HTTPHost:             cmd.DefaultHTTPHost,
		HTTPPort:             9874,
		WSEnable:             true,
		WSHost:               cmd.DefaultWSHost,
		WSPort:               9875,
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
