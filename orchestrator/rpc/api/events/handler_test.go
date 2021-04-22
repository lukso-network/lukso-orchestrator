package events

import (
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	eventTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	"github.com/prysmaticlabs/eth2-types"
	"testing"
	"time"
)

// setup
func setup(t *testing.T) (*MockBackend, *PublicFilterAPI) {
	consensusInfos := make([]*eventTypes.MinimalEpochConsensusInfo, 0)
	for i := 0; i < 5; i++ {
		consensusInfos = append(consensusInfos, testutil.NewMinimalConsensusInfo(types.Epoch(i)))
	}

	backend := &MockBackend{
		ConsensusInfos: consensusInfos,
		CurEpoch:       4,
	}

	eventApi := NewPublicFilterAPI(backend, deadline)
	return backend, eventApi
}

func subscribe(t *testing.T, eventApi *PublicFilterAPI, backend *MockBackend, fromEpoch uint64, curEpoch uint64) *Subscription {
	receiverChan := make(chan *eventTypes.MinimalEpochConsensusInfo)
	subscriber := eventApi.events.SubscribeConsensusInfo(receiverChan, fromEpoch)
	totalEvents := len(backend.ConsensusInfos) - int(fromEpoch)
	actualConsensusInfos := backend.ConsensusInfos

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
				var flag bool
				for _, c := range actualConsensusInfos {
					if c.Epoch == consensusInfo.Epoch {
						flag = true
						assert.DeepEqual(t, c, consensusInfo)
					}
				}
				assert.Equal(t, true, flag, "Not found")
				eventCount++
				epoch++
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
	backend, eventApi := setup(t)
	fromEpoch := uint64(0)
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
	backend, eventApi := setup(t)

	fromEpoch0 := uint64(0)
	subscriber0 := subscribe(t, eventApi, backend, fromEpoch0, backend.CurEpoch)

	fromEpoch1 := uint64(0)
	subscriber1 := subscribe(t, eventApi, backend, fromEpoch1, backend.CurEpoch)

	time.Sleep(1 * time.Second)
	consensusInfo := testutil.NewMinimalConsensusInfo(types.Epoch(5))
	backend.ConsensusInfoFeed.Send(consensusInfo)

	<-subscriber0.Err()
	<-subscriber1.Err()
}

// Test_MinimalConsensusInfo_With_Future_Epoch checks when subscriber subscribes from future epoch
//func Test_MinimalConsensusInfo_With_Future_Epoch(t *testing.T) {
//	backend, eventApi := setup(t)
//	fromEpoch := uint64(0) // 20 is the future epoch
//	subscriber := subscribe(t, eventApi, backend, fromEpoch, 5)
//
//	time.Sleep(1 * time.Second)
//
//	curEpoch := uint64(0)
//	consensusInfo := testutil.NewMinimalConsensusInfo(types.Epoch(curEpoch))
//	backend.ConsensusInfos = append(backend.ConsensusInfos, consensusInfo)
//	backend.CurEpoch = curEpoch
//	backend.ConsensusInfoFeed.Send(consensusInfo)
//
//	<-subscriber.Err()
//}
