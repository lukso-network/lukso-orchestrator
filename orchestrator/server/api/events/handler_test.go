package events

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
    eth2Types "github.com/prysmaticlabs/eth2-types"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

var (
	deadline = 5 * time.Minute
)

type testBackend struct {
	mux             	*event.TypeMux
	sections        	uint64
	consensusInfoFeed   event.Feed
}

func(b *testBackend) SubscribeNewEpochEvent(ch chan<- *types.MinConsensusInfoEvent) event.Subscription {
	return b.consensusInfoFeed.Subscribe(ch)
}

// TestMinimalConsensusInfoSubscription tests if a block subscription returns block hashes for posted chain events.
// It creates multiple subscriptions:
// - one at the start and should receive all posted chain events and a second (blockHashes)
// - one that is created after a cutoff moment and uninstalled after a second cutoff moment (blockHashes[cutoff1:cutoff2])
// - one that is created after the second cutoff moment (blockHashes[cutoff2:])
func TestMinimalConsensusInfoSubscription(t *testing.T) {
	t.Parallel()

	var (
		backend     		= &testBackend{}
		api         		= NewPublicFilterAPI(backend, false, deadline)
		consensusInfoEvents = []*types.MinConsensusInfoEvent{}
	)

	for i := 0; i < 10; i++ {
		epoch := eth2Types.Epoch(rand.Uint64())
		validatorList := make([]string, 32)
		epochStartTime := rand.Uint64()
		slotTimeDuration := uint64(6)

		for idx := 0; idx < 32; idx++ {
			bs := []byte(strconv.Itoa(31415926))
			pubKey := common.Bytes2Hex(bs)
			validatorList[idx] = pubKey
		}

		consensusInfoEvents = append(consensusInfoEvents,
			&types.MinConsensusInfoEvent{
				Epoch: epoch,
				ValidatorList: validatorList,
				EpochStartTime: epochStartTime,
				SlotTimeDuration: slotTimeDuration,
			})
	}

	chan0 := make(chan *types.MinConsensusInfoEvent)
	sub0 := api.events.SubscribeConsensusInfo(chan0)
	chan1 := make(chan *types.MinConsensusInfoEvent)
	sub1 := api.events.SubscribeConsensusInfo(chan1)

	go func() { // simulate client
		i1, i2 := 0, 0
		for i1 != len(consensusInfoEvents) || i2 != len(consensusInfoEvents) {
			select {
			case consensusInfo := <-chan0:
				if consensusInfo.Epoch != consensusInfoEvents[i1].Epoch {
					t.Errorf("sub0 received invalid epcho on index %d, want %x, got %x", i1, consensusInfoEvents[i1].Epoch, consensusInfo.Epoch)
				}
				i1++
			case consensusInfo := <-chan1:
				if consensusInfo.Epoch != consensusInfoEvents[i2].Epoch {
					t.Errorf("sub1 received invalid epcho on index %d, want %x, got %x", i2, consensusInfoEvents[i2].Epoch, consensusInfo.Epoch)
				}
				i2++
			}
		}

		sub0.Unsubscribe()
		sub1.Unsubscribe()
	}()

	time.Sleep(1 * time.Second)
	for _, e := range consensusInfoEvents {
		backend.consensusInfoFeed.Send(e)
	}

	<-sub0.Err()
	<-sub1.Err()
}

