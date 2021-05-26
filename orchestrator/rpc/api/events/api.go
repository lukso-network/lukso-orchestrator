package events

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
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

type Status int

const (
	Pending Status = iota
	Verified
	Invalid
)

const (
	MockedHashInvalid = "0xc9a190eb52c18df5ffcb1d817214ecb08f025f8583805cd12064d30e3f9bd9d5"
	MockedHashPending = "0xa99c69a301564970956edd897ff0590f4c0f1031daa464ded655af65ad0906df"
)

// PublicFilterAPI offers support to create and manage filters. This will allow external clients to retrieve various
// information related to the Ethereum protocol such als blocks, transactions and logs.
type PublicFilterAPI struct {
	backend Backend
	events  *EventSystem
	timeout time.Duration
}

type BlockHash struct {
	Slot uint64
	Hash common.Hash
}

type BlockStatus struct {
	BlockHash
	Status Status
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

// ConfirmPanBlockHashes
func (api *PublicFilterAPI) ConfirmPanBlockHashes(
	ctx context.Context,
	request []*BlockHash,
) (response []*BlockStatus, err error) {
	if len(request) < 1 {
		err = fmt.Errorf("request has empty slice")

		return
	}

	response = make([]*BlockStatus, 0)

	for _, blockRequest := range request {
		status := Verified

		if MockedHashInvalid == blockRequest.Hash.String() {
			status = Invalid
		}

		if MockedHashPending == blockRequest.Hash.String() {
			status = Pending
		}

		response = append(response, &BlockStatus{
			BlockHash: BlockHash{
				Slot: blockRequest.Slot,
				Hash: blockRequest.Hash,
			},
			Status: status,
		})
	}

	log.WithField("method", "ConfirmPanBlockHashes").
		WithField("request", request).
		WithField("response", response).
		Info("Sending back ConfirmPanBlockHashes response")

	return
}

// ConfirmVanBlockHashes
func (api *PublicFilterAPI) ConfirmVanBlockHashes(
	ctx context.Context,
	request []*BlockHash,
) (response []*BlockStatus, err error) {
	if len(request) < 1 {
		err = fmt.Errorf("request has empty slice")

		return
	}

	response = make([]*BlockStatus, 0)

	for _, blockRequest := range request {
		status := Verified

		if MockedHashInvalid == blockRequest.Hash.String() {
			status = Invalid
		}

		if MockedHashPending == blockRequest.Hash.String() {
			status = Pending
		}

		response = append(response, &BlockStatus{
			BlockHash: BlockHash{
				Slot: blockRequest.Slot,
				Hash: blockRequest.Hash,
			},
			Status: status,
		})
	}

	log.WithField("method", "ConfirmVanBlockHashes").
		WithField("request", request).
		WithField("response", response).
		Info("Sending back ConfirmVanBlockHashes response")

	return
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

	// TODO: Consider change. This is due to the mismatch on slot 0 on pandora and vanguard
	timeMismatch := time.Second * 6

	go func() {
		consensusInfo := make(chan *eventTypes.MinimalEpochConsensusInfo)
		consensusInfoSub := api.events.SubscribeConsensusInfo(consensusInfo, epoch)
		log.WithField("fromEpoch", epoch).
			WithField("alreadyKnown", alreadyKnownEpochs).
			Info("registered new subscriber for consensus info")

		if len(alreadyKnownEpochs) < 1 {
			log.WithField("fromEpoch", epoch).
				Info("there are no already known epochs, try to fetch lowest")
		}

		for index, currentEpoch := range alreadyKnownEpochs {
			// TODO: Remove it ASAP. This should not be that way
			currentEpoch.EpochStartTime = currentEpoch.EpochStartTime - uint64(timeMismatch.Seconds())

			log.WithField("epoch", index).
				WithField("epochStartTime", currentEpoch.EpochStartTime).
				Info("sending already known consensus info to subscriber")
			err := notifier.Notify(rpcSub.ID, currentEpoch)

			if nil != err {
				log.WithField("context", "already known epochs notification failure").Error(err)
			}
		}

		for {
			select {
			case currentEpoch := <-consensusInfo:
				// TODO: Remove it asap
				currentEpoch.EpochStartTime = currentEpoch.EpochStartTime - uint64(timeMismatch.Seconds())
				log.WithField("epoch", currentEpoch.Epoch).Info("sending consensus info to subscriber")
				err := notifier.Notify(rpcSub.ID, currentEpoch)

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
