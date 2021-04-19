package events

import (
	"github.com/ethereum/go-ethereum/event"
	eventTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	types "github.com/prysmaticlabs/eth2-types"
	"time"
)

var (
	deadline = 5 * time.Minute
)

type MockBackend struct {
	ConsensusInfoFeed    event.Feed
	ConsensusInfoMapping map[types.Epoch]*eventTypes.MinConsensusInfoEvent
	CurEpoch             types.Epoch
}

func (backend *MockBackend) CurrentEpoch() types.Epoch {
	return backend.CurEpoch
}

func (backend *MockBackend) ConsensusInfoByEpochRange(fromEpoch, toEpoch types.Epoch,
) map[types.Epoch]*eventTypes.MinConsensusInfoEvent {

	consensusInfoMapping := make(map[types.Epoch]*eventTypes.MinConsensusInfoEvent)
	for epoch := fromEpoch; epoch <= toEpoch; epoch++ {
		item, exists := backend.ConsensusInfoMapping[epoch]
		if exists && item != nil {
			consensusInfoMapping[epoch] = item
		}
	}
	return consensusInfoMapping
}

func (b *MockBackend) SubscribeNewEpochEvent(ch chan<- *eventTypes.MinConsensusInfoEvent) event.Subscription {
	return b.ConsensusInfoFeed.Subscribe(ch)
}
