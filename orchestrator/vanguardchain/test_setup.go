package vanguardchain

import (
	"context"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/cache"
	testDB "github.com/lukso-network/lukso-orchestrator/orchestrator/db/testing"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/rpc/api/events"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/client"
	"github.com/lukso-network/lukso-orchestrator/shared/mock"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	eventTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	types "github.com/prysmaticlabs/eth2-types"
	eth "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
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
	consensusInfoClient eth.BeaconChain_StreamMinimalConsensusInfoClient
}

func (v vanClientMock) IsValidBlock(slot types.Slot, blockHash []byte) (bool, error) {
	panic("implement me")
}

func (v vanClientMock) GetFinalizedEpoch() (types.Epoch, error) {
	panic("implement me")
}

var (
	ConsensusInfoMocks        []*eth.MinimalConsensusInfo
	PendingBlockMocks         []*eth.BeaconBlock
	mockedStreamPendingBlocks eth.BeaconChain_StreamNewPendingBlocksClient = streamNewPendingBlocksClient{
		pendingBlocks: PendingBlockMocks,
	}
	mockedStreamConsensusInfoClient eth.BeaconChain_StreamMinimalConsensusInfoClient = streamConsensusInfoClient{
		consensusInfos: ConsensusInfoMocks,
	}
	mockedVanClientStruct = &vanClientMock{
		mockedStreamPendingBlocks,
		mockedStreamConsensusInfoClient,
	}
	mockedClient client.VanguardClient = mockedVanClientStruct
)

type streamNewPendingBlocksClient struct {
	pendingBlocks []*eth.BeaconBlock
}

func (s streamNewPendingBlocksClient) Recv() (*eth.BeaconBlock, error) {
	if len(PendingBlockMocks) > 0 {
		toReturn := PendingBlockMocks[0]
		PendingBlockMocks = append([]*eth.BeaconBlock(nil), PendingBlockMocks[1:]...)

		return toReturn, nil
	}

	//     Should not receive anything until mocks are present
	time.Sleep(time.Millisecond * 200)
	return s.Recv()
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

func (v vanClientMock) StreamMinimalConsensusInfo(epoch uint64) (stream eth.BeaconChain_StreamMinimalConsensusInfoClient, err error) {
	return v.consensusInfoClient, nil
}

func (v vanClientMock) StreamNewPendingBlocks(blockRoot []byte, fromSlot types.Slot) (eth.BeaconChain_StreamNewPendingBlocksClient, error) {
	return v.pendingBlocksClient, nil
}

func (v vanClientMock) Close() {
	panic("implement me")
}

type streamConsensusInfoClient struct {
	consensusInfos []*eth.MinimalConsensusInfo
}

func (s streamConsensusInfoClient) Recv() (*eth.MinimalConsensusInfo, error) {
	if len(ConsensusInfoMocks) > 0 {
		toReturn := ConsensusInfoMocks[0]
		ConsensusInfoMocks = append([]*eth.MinimalConsensusInfo(nil), ConsensusInfoMocks[1:]...)

		return toReturn, nil
	}

	//     Should not receive anything until mocks are present
	time.Sleep(time.Millisecond * 200)
	return s.Recv()
}

func (s streamConsensusInfoClient) Header() (metadata.MD, error) {
	panic("implement me")
}

func (s streamConsensusInfoClient) Trailer() metadata.MD {
	panic("implement me")
}

func (s streamConsensusInfoClient) CloseSend() error {
	panic("implement me")
}

func (s streamConsensusInfoClient) Context() context.Context {
	panic("implement me")
}

func (s streamConsensusInfoClient) SendMsg(m interface{}) error {
	panic("implement me")
}

func (s streamConsensusInfoClient) RecvMsg(m interface{}) error {
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
	level, err := logrus.ParseLevel("trace")
	assert.NoError(t, err)
	logrus.SetLevel(level)

	db := testDB.SetupDB(t)

	vanguardClientService, err := NewService(
		ctx,
		"127.0.0.1:4000",
		db,
		cache.NewVanShardInfoCache(1<<10),
		dialGRPCFn,
	)
	if err != nil {
		t.Fatalf("failed to create protocol stack: %v", err)
	}

	return vanguardClientService, nil
}
