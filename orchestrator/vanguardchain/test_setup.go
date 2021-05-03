package vanguardchain

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/rpc"
	testDB "github.com/lukso-network/lukso-orchestrator/orchestrator/db/testing"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/rpc/api/events"
	"github.com/lukso-network/lukso-orchestrator/shared/mock"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	eventTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	"github.com/sirupsen/logrus"
	"testing"
	"time"
)

type mocks struct {
	db *mock.MockDatabase
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
func SetupVanguardSvc(ctx context.Context, t *testing.T, dialRPCFn DialRPCFn) (*Service, *mocks) {
	level, err := logrus.ParseLevel("debug")
	assert.NoError(t, err)
	logrus.SetLevel(level)

	//ctrl := gomock.NewController(t)
	//m := &mocks{
	//	db: mock.NewMockDatabase(ctrl),
	//}

	vanguardClientService, err := NewService(
		ctx,
		"ws://127.0.0.1:8546",
		"van",
		testDB.SetupDB(t),
		dialRPCFn)
	if err != nil {
		t.Fatalf("failed to create protocol stack: %v", err)
	}

	return vanguardClientService, nil
}

// DialInProcClient creates in process client for vanguard mocked server
func DialInProcClient(server *rpc.Server) DialRPCFn {
	return func(endpoint string) (*rpc.Client, error) {
		client := rpc.DialInProc(server)
		if client == nil {
			return nil, errors.New("failed to create in-process client")
		}
		return client, nil
	}
}

// DialRPCClient creates in process client for vanguard rpc server
func DialRPCClient() DialRPCFn {
	return func(endpoint string) (*rpc.Client, error) {
		client, err := rpc.Dial(endpoint)
		if err != nil {
			return nil, err
		}
		return client, nil
	}
}
