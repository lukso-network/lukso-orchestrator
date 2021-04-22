package events

import (
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

func (backend *MockBackend) CurrentEpoch() uint64 {
	return backend.CurEpoch
}

func (backend *MockBackend) ConsensusInfoByEpochRange(fromEpoch uint64) []*eventTypes.MinimalEpochConsensusInfo {
	consensusInfos := make([]*eventTypes.MinimalEpochConsensusInfo, 0)
	for _, consensusInfo := range backend.ConsensusInfos {
		consensusInfos = append(consensusInfos, consensusInfo)
	}
	return consensusInfos
}

func (b *MockBackend) SubscribeNewEpochEvent(ch chan<- *eventTypes.MinimalEpochConsensusInfo) event.Subscription {
	return b.ConsensusInfoFeed.Subscribe(ch)
}
