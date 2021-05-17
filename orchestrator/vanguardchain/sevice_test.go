package vanguardchain

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"testing"
	"time"
)

// Test_VanguardSvc_StartStop checks start and stop process. When the vanguard service starts, it also subscribes
// van_subscribe to get new consensus info
func Test_VanguardSvc_StartStop(t *testing.T) {
	hook := logTest.NewGlobal()
	ctx := context.Background()
	reConPeriod = 1 * time.Second

	// setup in process server and client
	mockServer, _ := SetupInProcServer(t)
	defer mockServer.Stop()

	// setup vanguard service
	dialInProcRPCClient := DialInProcClient(mockServer)
	vanSvc, _ := SetupVanguardSvc(ctx, t, dialInProcRPCClient, GRPCFunc)
	vanSvc.Start()

	time.Sleep(1 * time.Second)
	assert.LogsContain(t, hook, "Connected vanguard chain")
	assert.LogsContain(t, hook, "subscribed to vanguard chain for consensus info")
	hook.Reset()
	assert.NoError(t, vanSvc.Stop())

}

// Test_VanguardSvc_NoServerConn checks that vanguard service
// should try to ping the server after certain period
func Test_VanguardSvc_NoServerConn(t *testing.T) {
	hook := logTest.NewGlobal()
	ctx := context.Background()
	reConPeriod = 1 * time.Second

	// setup vanguard service
	dialRPCClient := DialRPCClient()
	vanSvc, _ := SetupVanguardSvc(ctx, t, dialRPCClient, GRPCFunc)
	vanSvc.Start()

	time.Sleep(2 * time.Second)
	assert.LogsContain(t, hook, "Could not connect to vanguard endpoint")
	hook.Reset()
}

// Test_VanguardSvc_RetryToConnServer
func Test_VanguardSvc_RetryToConnServer(t *testing.T) {
	hook := logTest.NewGlobal()
	ctx := context.Background()
	reConPeriod = 1 * time.Second

	// setup vanguard service
	dialRPCClient := DialRPCClient()
	vanSvc, _ := SetupVanguardSvc(ctx, t, dialRPCClient, GRPCFunc)
	vanSvc.Start()
	defer vanSvc.Stop()

	time.Sleep(1 * time.Second)
	assert.LogsContain(t, hook, "Could not connect to vanguard endpoint")

	//setup in process server and client
	mockServer, _ := SetupInProcServer(t)
	defer mockServer.Stop()
	dialInProcRPCClient := DialInProcClient(mockServer)
	vanSvc.dialRPCFn = dialInProcRPCClient

	time.Sleep(2 * time.Second)
	assert.LogsContain(t, hook, "Connected vanguard chain")
}
