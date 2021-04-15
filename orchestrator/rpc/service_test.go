package rpc

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/shared/cmd"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"testing"
)

func testServerConfig() *Config {
	return &Config{
		HTTPHost: cmd.DefaultHTTPHost,
		HTTPPort: cmd.DefaultHTTPPort,
		WSHost:   cmd.DefaultWSHost,
		WSPort:   cmd.DefaultWSPort,
	}
}

// Tests that an empty protocol stack can be closed more than once.
func TestServerStartMultipleTimes(t *testing.T) {
	hook := logTest.NewGlobal()
	ctx := context.Background()
	rpcService, err := NewService(ctx, testServerConfig())
	if err != nil {
		t.Fatalf("failed to create protocol stack: %v", err)
	}
	// Ensure that a node can be successfully started, but only once
	rpcService.Start()

	require.LogsContain(t, hook, "listening on port")
	assert.NoError(t, rpcService.Stop())
}

func TestRPCServiceStart(t *testing.T) {
	ctx := context.Background()
	rpcService, err := NewService(ctx, testServerConfig())
	if err != nil {
		t.Fatalf("failed to create protocol stack: %v", err)
	}
	if err := rpcService.startRPC(); err != nil {
		t.Fatalf("failed to start rpc service: %v", err)
	}
}
