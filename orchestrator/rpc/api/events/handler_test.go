package events

import (
	"github.com/lukso-network/lukso-orchestrator/orchestrator/rpc/api"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	eventTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	"testing"
	"time"
)

// setup
func setup(t *testing.T) (*api.MockBackend, *PublicFilterAPI) {
	consensusInfos := make([]*eventTypes.MinimalEpochConsensusInfoV2, 0)
	for i := 0; i < 5; i++ {
		consensusInfos = append(consensusInfos, testutil.NewMinimalConsensusInfo(uint64(i)))
	}

	backend := &api.MockBackend{
		ConsensusInfos: consensusInfos,
		CurEpoch:       4,
	}

	eventApi := NewPublicFilterAPI(backend, api.Deadline)
	return backend, eventApi
}

func subscribe(
	t *testing.T,
	eventApi *PublicFilterAPI,
	fromEpoch uint64,
) *Subscription {

	receiverChan := make(chan *eventTypes.MinimalEpochConsensusInfoV2)
	subscriber := eventApi.events.SubscribeConsensusInfo(receiverChan, fromEpoch)
	expectedConInfo := testutil.NewMinimalConsensusInfo(5)

	go func() { // simulate client
		select {
		case consensusInfo := <-receiverChan:
			assert.DeepEqual(t, expectedConInfo, consensusInfo)
			subscriber.Unsubscribe()
		}
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
	subscriber := subscribe(t, eventApi, fromEpoch)

	time.Sleep(1 * time.Second)
	consensusInfo := testutil.NewMinimalConsensusInfo(5)
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
	subscriber0 := subscribe(t, eventApi, fromEpoch0)

	fromEpoch1 := uint64(0)
	subscriber1 := subscribe(t, eventApi, fromEpoch1)

	time.Sleep(1 * time.Second)
	consensusInfo := testutil.NewMinimalConsensusInfo(5)
	backend.ConsensusInfoFeed.Send(consensusInfo)

	<-subscriber0.Err()
	<-subscriber1.Err()
}

// Test_MinimalConsensusInfo_With_Future_Epoch checks when subscriber subscribes from future epoch
func Test_MinimalConsensusInfo_With_Future_Epoch(t *testing.T) {
	backend, eventApi := setup(t)
	fromEpoch := uint64(20) // 20 is the future epoch
	subscriber := subscribe(t, eventApi, fromEpoch)

	time.Sleep(1 * time.Second)

	curEpoch := uint64(5)
	consensusInfo := testutil.NewMinimalConsensusInfo(curEpoch)
	backend.ConsensusInfoFeed.Send(consensusInfo)

	<-subscriber.Err()
}
