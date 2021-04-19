package events

import (
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	eventTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	"github.com/prysmaticlabs/eth2-types"
	"testing"
	"time"
)

// setup
func setup(t *testing.T) (*MockBackend, *PublicFilterAPI) {
	consensusInfoMapping := make(map[types.Epoch]*eventTypes.MinimalEpochConsensusInfo, 10)
	for i := 0; i < 5; i++ {
		consensusInfoMapping[types.Epoch(i)] = testutil.NewMinimalConsensusInfo(types.Epoch(i))
	}

	backend := &MockBackend{
		ConsensusInfoMapping: consensusInfoMapping,
		CurEpoch:             types.Epoch(4),
	}

	eventApi := NewPublicFilterAPI(backend, deadline)
	return backend, eventApi
}

func subscribe(t *testing.T, eventApi *PublicFilterAPI, backend *MockBackend, fromEpoch types.Epoch, curEpoch types.Epoch) *Subscription {
	receiverChan := make(chan *eventTypes.MinimalEpochConsensusInfo)
	subscriber := eventApi.events.SubscribeConsensusInfo(receiverChan, fromEpoch)
	totalEvents := len(backend.ConsensusInfoMapping) - int(fromEpoch)

	// when subscribe from future epoch
	if totalEvents <= 0 {
		totalEvents = 1
		fromEpoch = curEpoch
	}

	go func() { // simulate client
		eventCount := 0
		epoch := fromEpoch
		for eventCount != totalEvents {
			select {
			case consensusInfo := <-receiverChan:
				if consensusInfo.Epoch != backend.ConsensusInfoMapping[epoch].Epoch {
					t.Errorf("subscriber received invalid epcho on index %d, want %x, got %x",
						epoch, backend.ConsensusInfoMapping[epoch].Epoch, consensusInfo.Epoch)
				}
				eventCount++
				epoch = epoch.Add(1)
			}
		}

		subscriber.Unsubscribe()
	}()

	return subscriber
}

// Test_MinimalConsensusInfo_One_Subscriber_Success test the consensusInfo subscription event.
// Test config: In this test, one subscriber subscribes for consensus info from epoch-0 and backend service
// has already 5 epoch consensus information in memory.
// Expected behaviour is that - subscriber will get consensus info from epoch-0 to epoch-4.
func Test_MinimalConsensusInfo_One_Subscriber_Success(t *testing.T) {
	t.Parallel()
	backend, eventApi := setup(t)
	fromEpoch := types.Epoch(0)
	subscriber := subscribe(t, eventApi, backend, fromEpoch, backend.CurEpoch)

	time.Sleep(1 * time.Second)
	consensusInfo := testutil.NewMinimalConsensusInfo(types.Epoch(5))
	backend.ConsensusInfoFeed.Send(consensusInfo)

	<-subscriber.Err()
}

// Test_MinimalConsensusInfo_Multiple_Subscriber_Success test the consensusInfo subscription event.
// Test config: In this test, multiple subscribers subscribe for consensus info from different epochs and backend service
// has already 5 epoch consensus information in memory.
// Expected behaviour is that - subscribers will get expected consensus info
func Test_MinimalConsensusInfo_Multiple_Subscriber_Success(t *testing.T) {
	t.Parallel()
	backend, eventApi := setup(t)

	fromEpoch0 := types.Epoch(0)
	subscriber0 := subscribe(t, eventApi, backend, fromEpoch0, backend.CurEpoch)

	fromEpoch1 := types.Epoch(3)
	subscriber1 := subscribe(t, eventApi, backend, fromEpoch1, backend.CurEpoch)

	time.Sleep(1 * time.Second)
	consensusInfo := testutil.NewMinimalConsensusInfo(types.Epoch(5))
	backend.ConsensusInfoFeed.Send(consensusInfo)

	<-subscriber0.Err()
	<-subscriber1.Err()
}

// Test_MinimalConsensusInfo_With_Future_Epoch checks when subscriber subscribes from future epoch
func Test_MinimalConsensusInfo_With_Future_Epoch(t *testing.T) {
	t.Parallel()
	backend, eventApi := setup(t)
	fromEpoch := types.Epoch(20) // 20 is the future epoch
	subscriber := subscribe(t, eventApi, backend, fromEpoch, types.Epoch(5))

	time.Sleep(1 * time.Second)

	curEpoch := types.Epoch(5)
	consensusInfo := testutil.NewMinimalConsensusInfo(curEpoch)
	backend.ConsensusInfoMapping[curEpoch] = consensusInfo
	backend.CurEpoch = curEpoch
	backend.ConsensusInfoFeed.Send(consensusInfo)

	<-subscriber.Err()
}
