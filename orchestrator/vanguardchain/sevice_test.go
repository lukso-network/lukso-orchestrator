package vanguardchain

import (
	"context"
	"fmt"
	duration "github.com/golang/protobuf/ptypes/duration"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/client"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	eth "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"testing"
	"time"
)

// Test_VanguardSvc_StartStop checks start and stop process. When the vanguard service starts, it also subscribes
// van_subscribe to get new consensus info
func Test_VanguardSvc_StartStop(t *testing.T) {
	hook := logTest.NewGlobal()
	reConPeriod = 1 * time.Second

	ConsensusInfoMocks = make([]*eth.MinimalConsensusInfo, 0)
	ConsensusInfoMocks = append(ConsensusInfoMocks, &eth.MinimalConsensusInfo{
		SlotTimeDuration: &duration.Duration{Seconds: 6}})

	PendingBlockMocks = nil

	defer func() {
		CleanConsensusMocks()
		CleanPendingBlocksMocks()
	}()

	ctx := context.Background()
	vanSvc, _ := SetupVanguardSvc(ctx, t, GRPCFunc)

	vanSvc.Start()
	defer func() {
		_ = vanSvc.Stop()
	}()

	time.Sleep(1 * time.Second)
	assert.LogsContain(t, hook, "Connected vanguard chain")
	hook.Reset()
}

// Test_VanguardSvc_StartStop checks start and stop process. When the vanguard service starts, it also subscribes
// van_subscribe to get new consensus info
func Test_VanguardSvc_NoServerConn(t *testing.T) {
	hook := logTest.NewGlobal()
	reConPeriod = 1 * time.Second

	ConsensusInfoMocks = make([]*eth.MinimalConsensusInfo, 0)
	ConsensusInfoMocks = append(ConsensusInfoMocks, &eth.MinimalConsensusInfo{
		SlotTimeDuration: &duration.Duration{Seconds: 6}})

	PendingBlockMocks = nil

	defer func() {
		CleanConsensusMocks()
		CleanPendingBlocksMocks()
	}()

	ctx := context.Background()
	vanSvc, _ := SetupVanguardSvc(ctx, t, GRPCFunc)
	vanSvc.vanguardRpc = "wsad://invalid.not.reachable!@:BrOKkeeeeeennnnnnnnnn"
	vanSvc.dialGRPCFn = func(endpoint string) (client.VanguardClient, error) {
		return nil, fmt.Errorf("dummy error")
	}

	vanSvc.Start()
	defer func() {
		_ = vanSvc.Stop()
	}()

	time.Sleep(1 * time.Second)
	assert.LogsContain(t, hook, "Could not connect to vanguard endpoint")
	hook.Reset()
}

func Test_VanguardSvc_RetryToConnServer(t *testing.T) {
	hook := logTest.NewGlobal()
	reConPeriod = 1 * time.Second

	ConsensusInfoMocks = make([]*eth.MinimalConsensusInfo, 0)
	ConsensusInfoMocks = append(ConsensusInfoMocks, &eth.MinimalConsensusInfo{
		SlotTimeDuration: &duration.Duration{Seconds: 6}})

	PendingBlockMocks = nil

	defer func() {
		CleanConsensusMocks()
		CleanPendingBlocksMocks()
	}()

	ctx := context.Background()
	vanSvc, _ := SetupVanguardSvc(ctx, t, GRPCFunc)
	shouldPass := false

	vanSvc.dialGRPCFn = func(endpoint string) (client.VanguardClient, error) {
		if shouldPass {
			return GRPCFunc(endpoint)
		}

		return nil, fmt.Errorf("dummy error")
	}

	oldReconPeriod := reConPeriod
	reConPeriod = time.Millisecond * 50

	defer func() {
		reConPeriod = oldReconPeriod
	}()

	vanSvc.Start()
	defer func() {
		_ = vanSvc.Stop()
	}()

	time.Sleep(10 * reConPeriod)
	assert.LogsContain(t, hook, "Could not connect to vanguard endpoint")
	shouldPass = true

	time.Sleep(time.Second)
	assert.LogsContain(t, hook, "Connected vanguard chain")
	hook.Reset()
}

func CleanConsensusMocks() {
	ConsensusInfoMocks = nil
}

func CleanPendingBlocksMocks() {
	PendingBlockMocks = nil
}
