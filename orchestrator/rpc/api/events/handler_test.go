package events

import (
	"context"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/lukso-network/lukso-orchestrator/shared/fork"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	eventTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	generalTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	"testing"
	"time"
)

// setup
func setup(t *testing.T) (*MockBackend, *PublicFilterAPI) {
	consensusInfos := make([]*eventTypes.MinimalEpochConsensusInfo, 0)
	for i := 0; i < 5; i++ {
		consensusInfos = append(consensusInfos, testutil.NewMinimalConsensusInfo(uint64(i)))
	}

	backend := &MockBackend{
		ConsensusInfos: consensusInfos,
		CurEpoch:       4,
	}

	eventApi := NewPublicFilterAPI(backend, deadline)
	return backend, eventApi
}

func subscribeMinimalConsensusInfo(
	t *testing.T,
	eventApi *PublicFilterAPI,
	fromEpoch uint64,
) *Subscription {

	receiverChan := make(chan *eventTypes.MinimalEpochConsensusInfo)
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
	subscriber := subscribeMinimalConsensusInfo(t, eventApi, fromEpoch)

	time.Sleep(1 * time.Second)
	consensusInfo := testutil.NewMinimalConsensusInfo(5)
	backend.ConsensusInfoFeed.Send(consensusInfo)

	<-subscriber.Err()
}

// Test_MinimalConsensusInfo_Multiple_Subscriber_Success test the consensusInfo subscription event.
// Test config: In this test, multiple subscribers subscribeMinimalConsensusInfo for consensus info from different epochs and backend service
// has already 5 epoch consensus information in memory.
// Expected behaviour is that - subscribers will get expected consensus info
func Test_MinimalConsensusInfo_Multiple_Subscriber_Success(t *testing.T) {
	backend, eventApi := setup(t)

	fromEpoch0 := uint64(0)
	subscriber0 := subscribeMinimalConsensusInfo(t, eventApi, fromEpoch0)

	fromEpoch1 := uint64(0)
	subscriber1 := subscribeMinimalConsensusInfo(t, eventApi, fromEpoch1)

	time.Sleep(1 * time.Second)
	consensusInfo := testutil.NewMinimalConsensusInfo(5)
	backend.ConsensusInfoFeed.Send(consensusInfo)

	<-subscriber0.Err()
	<-subscriber1.Err()
}

func Test_PublicFilterApi_FirstAid(t *testing.T) {
	_, eventApi := setup(t)

	t.Run("should return aid for invalid hash", func(t *testing.T) {
		for slot, hash := range fork.UnsupportedForkL15PandoraProd {
			healHeader, err := eventApi.FirstAid(context.Background(), hash, slot)
			assert.NoError(t, err)
			assert.NotNil(t, healHeader)
		}
	})

	t.Run("should return no information for unkown hashes", func(t *testing.T) {
		healHeader, err := eventApi.FirstAid(context.Background(), eth1Types.EmptyUncleHash, 25)
		assert.NoError(t, err)
		assert.Equal(t, (*eth1Types.Header)(nil), healHeader)
	})

}

// Test_MinimalConsensusInfo_With_Future_Epoch checks when subscriber subscribes from future epoch
func Test_MinimalConsensusInfo_With_Future_Epoch(t *testing.T) {
	backend, eventApi := setup(t)
	fromEpoch := uint64(20) // 20 is the future epoch
	subscriber := subscribeMinimalConsensusInfo(t, eventApi, fromEpoch)

	time.Sleep(1 * time.Second)

	curEpoch := uint64(5)
	consensusInfo := testutil.NewMinimalConsensusInfo(curEpoch)
	backend.ConsensusInfoFeed.Send(consensusInfo)

	<-subscriber.Err()
}

func Test_ConfirmPanBlockHashes_Fork(t *testing.T) {
	_, eventApi := setup(t)
	ctx := context.Background()
	require.Equal(t, true, len(fork.UnsupportedForkL15PandoraProd) > 0)

	t.Run("should restrict fork", func(t *testing.T) {
		request := make([]*BlockHash, 0)

		var (
			chosenSlot uint64
			chosenHash BlockHash
		)

		for slot, hash := range fork.UnsupportedForkL15PandoraProd {
			chosenSlot = slot
			chosenHash = BlockHash{Slot: chosenSlot, Hash: hash}
			request = append(request, &chosenHash)
			break
		}

		blockStatuses, err := eventApi.ConfirmPanBlockHashes(ctx, request)
		require.NoError(t, err)

		expectedBlockStatuses := make([]*BlockStatus, 1)
		expectedBlockStatuses[0] = &BlockStatus{
			BlockHash: chosenHash,
			Status:    generalTypes.Invalid,
		}

		require.DeepEqual(t, expectedBlockStatuses, blockStatuses)
	})

	t.Run("should return desired fork", func(t *testing.T) {
		request := make([]*BlockHash, 0)

		var (
			chosenSlot uint64
			chosenHash BlockHash
		)

		for slot, hash := range fork.SupportedForkL15PandoraProd {
			chosenSlot = slot
			chosenHash = BlockHash{Slot: chosenSlot, Hash: hash}
			request = append(request, &chosenHash)
			break
		}

		blockStatuses, err := eventApi.ConfirmPanBlockHashes(ctx, request)
		require.NoError(t, err)

		expectedBlockStatuses := make([]*BlockStatus, 1)
		expectedBlockStatuses[0] = &BlockStatus{
			BlockHash: chosenHash,
			Status:    generalTypes.Verified,
		}

		require.DeepEqual(t, expectedBlockStatuses, blockStatuses)
	})
}
