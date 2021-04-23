package vanguardchain

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/sirupsen/logrus"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"testing"
	"time"
)

// Test_VanguardSvc_StartStop checks start and stop process. When the vanguard service starts, it also subscribes
// van_subscribe to get new consensus info
func Test_VanguardSvc_StartStop(t *testing.T) {
	level, err := logrus.ParseLevel("debug")
	assert.NoError(t, err)
	logrus.SetLevel(level)
	hook := logTest.NewGlobal()
	ctx := context.Background()

	// setup in process server and client
	mockServer, _ := SetupInProcServer(t)
	defer mockServer.Stop()

	// setup vanguard service
	dialInProcRPCClient := DialInProcClient(mockServer)
	vanSvc, m := SetupVanguardSvc(ctx, t, dialInProcRPCClient)
	vanSvc.Start()
	m.db.EXPECT().LatestSavedEpoch().Return(uint64(0), nil) // nil - error

	time.Sleep(1 * time.Second)
	assert.LogsContain(t, hook, "Connected vanguard chain")
	assert.LogsContain(t, hook, "subscribed to vanguard chain for consensus info")
	hook.Reset()
	assert.NoError(t, vanSvc.Stop())
}

// Test_VanguardSvc_NoServerConn checks that vanguard service
// should try to ping the server after certain period
func Test_VanguardSvc_NoServerConn(t *testing.T) {
	level, err := logrus.ParseLevel("debug")
	assert.NoError(t, err)
	logrus.SetLevel(level)
	hook := logTest.NewGlobal()
	ctx := context.Background()

	// setup vanguard service
	dialRPCClient := DialRPCClient()
	vanSvc, _ := SetupVanguardSvc(ctx, t, dialRPCClient)
	vanSvc.Start()

	time.Sleep(10 * time.Second)
	assert.LogsContain(t, hook, "Could not connect to vanguard endpoint")
	hook.Reset()
}

// Test_VanguardSvc_RetryToConnServer
func Test_VanguardSvc_RetryToConnServer(t *testing.T) {
	hook := logTest.NewGlobal()
	ctx := context.Background()

	// setup vanguard service
	dialRPCClient := DialRPCClient()
	vanSvc, m := SetupVanguardSvc(ctx, t, dialRPCClient)
	vanSvc.Start()
	defer vanSvc.Stop()

	time.Sleep(1 * time.Second)
	assert.LogsContain(t, hook, "Could not connect to vanguard endpoint")

	//setup in process server and client
	mockServer, _ := SetupInProcServer(t)
	defer mockServer.Stop()
	dialInProcRPCClient := DialInProcClient(mockServer)
	m.db.EXPECT().LatestSavedEpoch().Return(uint64(0), nil) // nil - error
	vanSvc.dialRPCFn = dialInProcRPCClient

	time.Sleep(2 * time.Second)
	assert.LogsContain(t, hook, "Connected vanguard chain")
}
