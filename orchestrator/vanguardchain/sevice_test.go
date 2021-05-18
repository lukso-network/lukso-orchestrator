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

// Test_VanguardSvc_StartStop checks start and stop process. When the vanguard service starts, it also subscribes
// van_subscribe to get new consensus info
func Test_VanguardSvc_StartStop(t *testing.T) {
	hook := logTest.NewGlobal()
	reConPeriod = 1 * time.Second

	ConsensusInfoMocks = make([]*eth.MinimalConsensusInfo, 0)
	ConsensusInfoMocks = append(ConsensusInfoMocks, &eth.MinimalConsensusInfo{
		SlotTimeDuration: &types.Duration{Seconds: 6}})

	PendingBlockMocks = make([]*eth.BeaconBlock, 0)
	PendingBlockMocks = append(PendingBlockMocks, &eth.BeaconBlock{})

	// setup in process server and client
	mockServer, _ := SetupInProcServer(t)
	defer mockServer.Stop()

	time.Sleep(1 * time.Second)
	assert.LogsContain(t, hook, "Connected vanguard chain")
	assert.LogsContain(t, hook, "subscribed to vanguard chain for consensus info")
	hook.Reset()

}

// Test_VanguardSvc_NoServerConn checks that vanguard service
// should try to ping the server after certain period
func Test_VanguardSvc_NoServerConn(t *testing.T) {
	hook := logTest.NewGlobal()
	reConPeriod = 1 * time.Second

	ConsensusInfoMocks = make([]*eth.MinimalConsensusInfo, 0)
	ConsensusInfoMocks = append(ConsensusInfoMocks, &eth.MinimalConsensusInfo{
		SlotTimeDuration: &types.Duration{Seconds: 6}})
	PendingBlockMocks = make([]*eth.BeaconBlock, 0)
	PendingBlockMocks = append(PendingBlockMocks, &eth.BeaconBlock{})

	defer func() {
		CleanConsensusMocks()
		CleanPendingBlocksMocks()
	}()

	time.Sleep(2 * time.Second)
	assert.LogsContain(t, hook, "Could not connect to vanguard endpoint")
	hook.Reset()
}

// Test_VanguardSvc_RetryToConnServer
func Test_VanguardSvc_RetryToConnServer(t *testing.T) {
	hook := logTest.NewGlobal()
	ctx := context.Background()
	reConPeriod = 1 * time.Second

	ConsensusInfoMocks = make([]*eth.MinimalConsensusInfo, 0)
	ConsensusInfoMocks = append(ConsensusInfoMocks, &eth.MinimalConsensusInfo{
		SlotTimeDuration: &types.Duration{Seconds: 6}})

	PendingBlockMocks = make([]*eth.BeaconBlock, 0)
	PendingBlockMocks = append(PendingBlockMocks, &eth.BeaconBlock{Slot: 0})

	defer func() {
		CleanConsensusMocks()
		CleanPendingBlocksMocks()
	}()

	// setup vanguard service
	vanSvc, _ := SetupVanguardSvc(ctx, t, GRPCFunc)
	vanSvc.Start()
	defer vanSvc.Stop()

	time.Sleep(1 * time.Second)
	assert.LogsContain(t, hook, "Could not connect to vanguard endpoint")

	//setup in process server and client
	mockServer, _ := SetupInProcServer(t)
	defer mockServer.Stop()

	time.Sleep(2 * time.Second)
	assert.LogsContain(t, hook, "Connected vanguard chain")
}

func CleanConsensusMocks() {
	ConsensusInfoMocks = nil
}

func CleanPendingBlocksMocks() {
	PendingBlockMocks = nil
}
