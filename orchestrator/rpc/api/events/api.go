package events

import (
	"context"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/rpc"
	eventTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	"time"
)

type Backend interface {
	CurrentEpoch() uint64
	ConsensusInfoByEpochRange(fromEpoch uint64) []*eventTypes.MinimalEpochConsensusInfo
	SubscribeNewEpochEvent(chan<- *eventTypes.MinimalEpochConsensusInfo) event.Subscription
}

// PublicFilterAPI offers support to create and manage filters. This will allow external clients to retrieve various
// information related to the Ethereum protocol such als blocks, transactions and logs.
type PublicFilterAPI struct {
	backend Backend
	events  *EventSystem
	timeout time.Duration
}

// NewPublicFilterAPI returns a new PublicFilterAPI instance.
func NewPublicFilterAPI(backend Backend, timeout time.Duration) *PublicFilterAPI {
	api := &PublicFilterAPI{
		backend: backend,
		events:  NewEventSystem(backend),
		timeout: timeout,
	}

	return api
}

// MinimalConsensusInfo
func (api *PublicFilterAPI) MinimalConsensusInfo(ctx context.Context, epoch uint64) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}
	rpcSub := notifier.CreateSubscription()

	// Fill already known epochs
	alreadyKnownEpochs := api.backend.ConsensusInfoByEpochRange(epoch)

	go func() {
		consensusInfo := make(chan *eventTypes.MinimalEpochConsensusInfo)
		consensusInfoSub := api.events.SubscribeConsensusInfo(consensusInfo, epoch)
		log.WithField("fromEpoch", epoch).Info("registered new subscriber for consensus info")

		for index, currentEpoch := range alreadyKnownEpochs {
			log.WithField("epoch", index).WithField("epochStartTime", currentEpoch.EpochStartTime).Info("sending already known consensus info to subscriber")
			err := notifier.Notify(rpcSub.ID, currentEpoch)

			if nil != err {
				log.WithField("context", "already known epochs notification failure").Error(err)
			}
		}

		for {
			select {
			case c := <-consensusInfo:
				log.WithField("epoch", c.Epoch).Info("sending consensus info to subscriber")
				err := notifier.Notify(rpcSub.ID, c)

				if nil != err {
					log.WithField("context", "error during epoch send").Error(err)
				}
			case <-rpcSub.Err():
				log.Info("unsubscribing registered subscriber")
				consensusInfoSub.Unsubscribe()
				return
			case <-notifier.Closed():
				log.Info("unsubscribing registered subscriber")
				consensusInfoSub.Unsubscribe()
				return
			}
		}
	}()

	return rpcSub, nil
}
