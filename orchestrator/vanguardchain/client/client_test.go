package client_test

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/client"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	"math"
	"testing"
	"time"
)

func TestVanguardClient(t *testing.T) {
	tests := []struct{
		vanguardRpcEndpoint string
	}{
		{
			vanguardRpcEndpoint: "127.0.0.1:4000",
		},
		{
			vanguardRpcEndpoint: "http://127.0.0.1:4000",
		},
		{
			vanguardRpcEndpoint: "./vanguard.ipc",
		},
		{
			vanguardRpcEndpoint: "/tmp/vanguard.ipc",
		},
	}

	for _, tt := range tests{
		ctx := context.Background()
		//vanguardService, _ := vanguardchain.SetupVanguardSvc(ctx, t, vanguardchain.GRPCFunc)

		vanguardClient, err := client.Dial(ctx, tt.vanguardRpcEndpoint, time.Minute*6, 32, math.MaxInt32)
		require.NoError(t, err)
		require.NotNil(t, vanguardClient)
	}
}

