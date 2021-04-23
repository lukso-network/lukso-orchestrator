package vanguardchain

import (
	"context"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/golang/mock/gomock"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	types "github.com/prysmaticlabs/eth2-types"
	"github.com/sirupsen/logrus"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"testing"
	"time"
)

// Test_VanguardChainStartStop_Initialized
func Test_VanguardChainStartStop_Initialized(t *testing.T) {
	level, err := logrus.ParseLevel("debug")
	assert.NoError(t, err)
	logrus.SetLevel(level)
	hook := logTest.NewGlobal()
	ctx := context.Background()

	mockServer, mockBackend := SetupInProcServer(t)
	mockClient := rpc.DialInProc(mockServer)
	if mockClient == nil {
		t.Fatal("failed to create inproc client")
	}
	defer mockServer.Stop()

	dialInProcRPCClient := DialInProcClient(mockServer)
	vanguardSvc, m := SetupVanguardSvc(ctx, t, dialInProcRPCClient)
	sub, err := vanguardSvc.subscribeNewConsensusInfo(ctx, 0, "van", mockClient)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)
	consensusInfo := testutil.NewMinimalConsensusInfo(types.Epoch(5))
	mockBackend.ConsensusInfoFeed.Send(consensusInfo)
	m.db.EXPECT().SaveConsensusInfo(ctx, gomock.Any()).Times(5).Return(nil)

	time.Sleep(1 * time.Second)
	assert.LogsContainNTimes(t, hook, "Got new consensus info from vanguard", 5)
	sub.Err()
	hook.Reset()
}
