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
	"strings"
	"sync"
	"testing"
	"time"
)

func TestVanguradChainStartStop_Initialized(t *testing.T) {
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
	vanguardSvc, m := SetupVanguardSvc(ctx, t)

	sub, err := vanguardSvc.subscribeNewConsensusInfo(ctx, 0, "van", mockClient)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)
	consensusInfo := testutil.NewMinimalConsensusInfo(types.Epoch(5))
	mockBackend.ConsensusInfoFeed.Send(consensusInfo)
	m.db.EXPECT().SaveConsensusInfo(ctx, gomock.Any()).Times(5).Return(nil)

	wg := sync.WaitGroup{}
	wg.Add(1)

	tries := 0
	go func() {
		want := "Got new consensus info from vanguard"
		for {
			entries := hook.AllEntries()
			totalExpectedLogs := 5
			logCount := 0
			for _, e := range entries {
				msg, err := e.String()
				if err != nil {
					t.Fatal(err)
					wg.Done()
					return
				}
				if strings.Contains(msg, want) {
					logCount++
				}
			}
			if logCount == totalExpectedLogs {
				wg.Done()
				return
			}
			tries++
			if tries > 50 {
				wg.Done()
				return
			}
		}
		wg.Done()
	}()

	wg.Wait()
	if tries > 50 {
		t.Fatal("Expected logs not found")
	}
	sub.Err()
	hook.Reset()
}