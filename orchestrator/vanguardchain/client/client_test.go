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
		name			    string
		vanguardRpcEndpoint string
	}{
		{
			name:				 "Test IPv4:Port dial",
			vanguardRpcEndpoint: "127.0.0.1:4000",
		},
		{
			name:				 "Test HTTP dial",
			vanguardRpcEndpoint: "http://127.0.0.1:4000",
		},
		{
			name:				 "Test IPC socket dial",
			vanguardRpcEndpoint: "./vanguard.ipc",
		},
		{
			name:				 "Test IPC socket dial (absolute unix path)",
			vanguardRpcEndpoint: "/tmp/vanguard.ipc",
		},
	}

	for _, tt := range tests{
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			vanguardClient, err := client.Dial(ctx, tt.vanguardRpcEndpoint, time.Minute*6, 32, math.MaxInt32)
			require.NoError(t, err)
			require.NotNil(t, vanguardClient)
			vanguardClient.Close()
		})
	}
}
