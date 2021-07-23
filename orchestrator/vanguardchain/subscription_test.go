package vanguardchain

import (
	"context"
	"github.com/gogo/protobuf/types"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	eth "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"testing"
	"time"
)

// Test_VanguardChainStartStop_Initialized
func Test_VanguardChainStartStop_Initialized(t *testing.T) {
	hook := logTest.NewGlobal()
	ctx := context.Background()
	vanSvc, _ := SetupVanguardSvc(ctx, t, GRPCFunc)
	vanSvc.Start()
	defer func() {
		_ = vanSvc.Stop()
	}()

	time.Sleep(1 * time.Second)
	ConsensusInfoMocks = make([]*eth.MinimalConsensusInfo, 0)
	minimalConsensusInfo := testutil.NewMinimalConsensusInfo(0)

	ConsensusInfoMocks = append(ConsensusInfoMocks, &eth.MinimalConsensusInfo{
		SlotTimeDuration: &types.Duration{Seconds: 6},
		ValidatorList:    minimalConsensusInfo.ValidatorList,
	})
	PendingBlockMocks = nil

	defer func() {
		ConsensusInfoMocks = nil
		PendingBlockMocks = nil
	}()

	time.Sleep(1 * time.Second)
	assert.LogsContain(t, hook, "Received new consensus info for next epoch")
	hook.Reset()
}
