package vanguardchain

import (
	"context"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/sirupsen/logrus"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"testing"
	"time"
)

func Test_VanguardSvc_Start_Stop(t *testing.T) {
	level, err := logrus.ParseLevel("debug")
	assert.NoError(t, err)
	logrus.SetLevel(level)
	hook := logTest.NewGlobal()
	ctx := context.Background()

	mockServer, _ := SetupInProcServer(t)
	mockClient := rpc.DialInProc(mockServer)
	if mockClient == nil {
		t.Fatal("failed to create inproc client")
	}
	defer mockServer.Stop()
	vanSvc, m := SetupVanguardSvc(ctx, t)
	vanSvc.Start()
	m.db.EXPECT().LatestSavedEpoch().Return(uint64(0), nil) // nil - error

	time.Sleep(1 * time.Second)
	assert.LogsContain(t, hook, "Connected vanguard chain")
	assert.LogsContain(t, hook, "subscribed to vanguard chain for consensus info")
	hook.Reset()
	assert.NoError(t, vanSvc.Stop())
}