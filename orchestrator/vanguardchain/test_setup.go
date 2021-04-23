package vanguardchain

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/golang/mock/gomock"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/rpc/api/events"
	"github.com/lukso-network/lukso-orchestrator/shared/mock"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	eventTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	types "github.com/prysmaticlabs/eth2-types"
	"testing"
	"time"
)

type mocks struct {
	db *mock.MockDatabase
}

// SetupInProcServer
func SetupInProcServer(t *testing.T) (*rpc.Server, *events.MockBackend) {
	consensusInfos := make([]*eventTypes.MinimalEpochConsensusInfo, 0)
	for i := 0; i < 5; i++ {
		consensusInfos = append(consensusInfos, testutil.NewMinimalConsensusInfo(types.Epoch(i)))
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

// SetupVanguardSvc
func SetupVanguardSvc(ctx context.Context, t *testing.T) (*Service, *mocks) {

	mockServer, _ := SetupInProcServer(t)
	dialInProcRPCClient := func (endpoint string) (*rpc.Client, error) {
		client := rpc.DialInProc(mockServer)
		if client == nil {
			return nil, errors.New("failed to create in-process client")
		}
		return client, nil
	}

	ctrl := gomock.NewController(t)
	m := &mocks{
		db: mock.NewMockDatabase(ctrl),
	}

	vanguardClientService, err := NewService(
		ctx,
		"ws://127.0.0.1:8546",
		"van",
		m.db,
		dialInProcRPCClient)
	if err != nil {
		t.Fatalf("failed to create protocol stack: %v", err)
	}

	return vanguardClientService, m
}



