package vanguardchain

import (
	"context"
	"github.com/ethereum/go-ethereum/rpc"
	testDB "github.com/lukso-network/lukso-orchestrator/orchestrator/db/testing"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/rpc/api/events"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/client"
	"github.com/lukso-network/lukso-orchestrator/shared/mock"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	eventTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	types "github.com/prysmaticlabs/eth2-types"
	eth "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
	"testing"
	"time"
)

type mocks struct {
	db *mock.MockDatabase
}

type vanClientMock struct {
	pendingBlocksClient eth.BeaconChain_StreamNewPendingBlocksClient
}

var (
	mockedStreamPendingBlocks eth.BeaconChain_StreamNewPendingBlocksClient = streamNewPendingBlocksClient{}
	mockedVanClientStruct                                                  = &vanClientMock{
		mockedStreamPendingBlocks,
	}
	mockedClient client.VanguardClient = mockedVanClientStruct
)

type streamNewPendingBlocksClient struct{}

func (s streamNewPendingBlocksClient) Recv() (*eth.BeaconBlock, error) {
	return &eth.BeaconBlock{}, nil
}

func (s streamNewPendingBlocksClient) Header() (metadata.MD, error) {
	panic("implement me")
}

func (s streamNewPendingBlocksClient) Trailer() metadata.MD {
	panic("implement me")
}

func (s streamNewPendingBlocksClient) CloseSend() error {
	panic("implement me")
}

func (s streamNewPendingBlocksClient) Context() context.Context {
	panic("implement me")
}

func (s streamNewPendingBlocksClient) SendMsg(m interface{}) error {
	panic("implement me")
}

func (s streamNewPendingBlocksClient) RecvMsg(m interface{}) error {
	panic("implement me")
}

func (v vanClientMock) CanonicalHeadSlot() (types.Slot, error) {
	panic("implement me")
}

func (v vanClientMock) StreamMinimalConsensusInfo() (stream eth.BeaconChain_StreamMinimalConsensusInfoClient, err error) {
	panic("implement me")
}

func (v vanClientMock) StreamNewPendingBlocks() (eth.BeaconChain_StreamNewPendingBlocksClient, error) {
	return v.pendingBlocksClient, nil
}

func (v vanClientMock) Close() {
	panic("implement me")
}

func GRPCFunc(endpoint string) (client.VanguardClient, error) {
	return mockedClient, nil
}

// SetupInProcServer prepares in process server with defined api. Here, this method mocks
// vanguard client's endpoint as well as backend. Use in-memory to mock the
func SetupInProcServer(t *testing.T) (*rpc.Server, *events.MockBackend) {
	consensusInfos := make([]*eventTypes.MinimalEpochConsensusInfo, 0)
	for i := 0; i < 5; i++ {
		consensusInfos = append(consensusInfos, testutil.NewMinimalConsensusInfo(uint64(i)))
	}

	backend := &events.MockBackend{
		ConsensusInfos: consensusInfos,
		CurEpoch:       4,
	}
	rpcApis := []rpc.API{
		{
			Namespace: "van",
			Version:   "1.0",
			Service:   events.NewPublicFilterAPI(backend, 5*time.Minute),
			Public:    true,
		},
	}
	iprocServer := rpc.NewServer()
	for _, api := range rpcApis {
		if err := iprocServer.RegisterName(api.Namespace, api.Service); err != nil {
			t.Fatal(err)
		}
	}
	return iprocServer, backend
}

// SetupVanguardSvc creates vanguard client service with mocked database
func SetupVanguardSvc(
	ctx context.Context,
	t *testing.T,
	dialGRPCFn DIALGRPCFn,
) (*Service, *mocks) {
	level, err := logrus.ParseLevel("debug")
	assert.NoError(t, err)
	logrus.SetLevel(level)

	db := testDB.SetupDB(t)

	vanguardClientService, err := NewService(
		ctx,
		"ws://127.0.0.1:8546",
		"127.0.0.1:4000",
		"van",
		db,
		db,
		dialGRPCFn,
	)
	if err != nil {
		t.Fatalf("failed to create protocol stack: %v", err)
	}

	return vanguardClientService, nil
}
