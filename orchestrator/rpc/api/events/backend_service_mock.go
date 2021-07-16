package events

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	eventTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	"time"
)

var (
	deadline = 5 * time.Minute
)

type MockBackend struct {
	ConsensusInfoFeed event.Feed
	ConsensusInfos    []*eventTypes.MinimalEpochConsensusInfo
	CurEpoch          uint64
}

func (b *MockBackend) GetPendingHashes() (response *PendingHashesResponse, err error) {
	panic("implement me")
}

var _ Backend = &MockBackend{}

func (b *MockBackend) FetchPanBlockStatus(slot uint64, hash common.Hash) (status Status, err error) {
	panic("implement me")
}

func (b *MockBackend) FetchVanBlockStatus(slot uint64, hash common.Hash) (status Status, err error) {
	panic("implement me")
}

func (b *MockBackend) CurrentEpoch() uint64 {
	return b.CurEpoch
}

func (b *MockBackend) ConsensusInfoByEpochRange(fromEpoch uint64) []*eventTypes.MinimalEpochConsensusInfo {
	consensusInfos := make([]*eventTypes.MinimalEpochConsensusInfo, 0)
	for _, consensusInfo := range b.ConsensusInfos {
		consensusInfos = append(consensusInfos, consensusInfo)
	}
	return consensusInfos
}

func (b *MockBackend) SubscribeNewEpochEvent(ch chan<- *eventTypes.MinimalEpochConsensusInfo) event.Subscription {
	return b.ConsensusInfoFeed.Subscribe(ch)
}

func (mb *MockBackend) GetSlotStatus(ctx context.Context, slot uint64, requestType bool) Status {
	return Pending
}
