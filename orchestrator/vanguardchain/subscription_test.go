package vanguardchain

import (
	"context"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"testing"
	"time"
)

// Test_VanguardChainStartStop_Initialized
func Test_VanguardChainStartStop_Initialized(t *testing.T) {
	hook := logTest.NewGlobal()
	ctx := context.Background()

	mockServer, mockBackend := SetupInProcServer(t)
	mockClient := rpc.DialInProc(mockServer)
	if mockClient == nil {
		t.Fatal("failed to create inproc client")
	}
	defer mockServer.Stop()

	dialInProcRPCClient := DialInProcClient(mockServer)
	vanguardSvc, _ := SetupVanguardSvc(ctx, t, dialInProcRPCClient, GRPCFunc)
	sub, err := vanguardSvc.subscribeNewConsensusInfo(ctx, 0, "van", mockClient)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)
	consensusInfo := testutil.NewMinimalConsensusInfo(5)
	mockBackend.ConsensusInfoFeed.Send(consensusInfo)

	time.Sleep(1 * time.Second)
	// TODO: Don't leave it as it is. Tests should not rely on logs. They should test side effect
	// I have changed behaviour of function entirely and test was still passing.
	assert.LogsContainNTimes(t, hook, "consensus info passed sanitization", 6)
	sub.Err()
	hook.Reset()
}
