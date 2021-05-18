package vanguardchain

import (
	"context"
	"github.com/gogo/protobuf/types"
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
	ConsensusInfoMocks = append(ConsensusInfoMocks, &eth.MinimalConsensusInfo{
		SlotTimeDuration: &types.Duration{Seconds: 6}})
	PendingBlockMocks = make([]*eth.BeaconBlock, 0)
	PendingBlockMocks = append(PendingBlockMocks, &eth.BeaconBlock{})

	defer func() {
		ConsensusInfoMocks = nil
		PendingBlockMocks = nil
	}()

	time.Sleep(1 * time.Second)
	// TODO: Don't leave it as it is. Tests should not rely on logs. They should test side effect
	// I have changed behaviour of function entirely and test was still passing.
	assert.LogsContain(t, hook, "consensus info passed sanitization")
	hook.Reset()
}
