package events

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db/kv"
	generalTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	"time"
)

type Backend interface {
	CurrentEpoch() uint64
	ConsensusInfoByEpochRange(fromEpoch uint64) []*generalTypes.MinimalEpochConsensusInfo
	SubscribeNewEpochEvent(chan<- *generalTypes.MinimalEpochConsensusInfo) event.Subscription
	FetchPanBlockStatus(slot uint64, hash common.Hash) (status Status, err error)
	FetchVanBlockStatus(slot uint64, hash common.Hash) (status Status, err error)
	GetPendingHashes() (response *PendingHashesResponse, err error)
}

type Status string

const (
	Pending  Status = "Pending"
	Verified Status = "Verified"
	Invalid  Status = "Invalid"
	Skipped  Status = "Skipped"
)

// PublicFilterAPI offers support to create and manage filters. This will allow external clients to retrieve various
// information related to the Ethereum protocol such als blocks, transactions and logs.
type PublicFilterAPI struct {
	backend Backend
	events  *EventSystem
	timeout time.Duration
}

type BlockHash struct {
	Slot uint64      `json:"slot"`
	Hash common.Hash `json:"hash"`
}

type BlockStatus struct {
	BlockHash
	Status Status
}

type RealmPair struct {
	Slot          uint64
	VanguardHash  *generalTypes.HeaderHash
	PandoraHashes []*generalTypes.HeaderHash
}

// TODO: consider it to merge into only string-based statuses
func FromDBStatus(status generalTypes.Status) (eventStatus Status) {
	if generalTypes.Pending == status {
		eventStatus = Pending
	}

	if generalTypes.Verified == status {
		eventStatus = Verified
	}

	if generalTypes.Invalid == status {
		eventStatus = Invalid
	}

	if generalTypes.Skipped == status {
		eventStatus = Skipped
	}

	return
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

// This is for debug purpose only
type PendingHashesResponse struct {
	VanguardHashes    []*generalTypes.HeaderHash
	PandoraHashes     []*generalTypes.HeaderHash
	VanguardHashesLen int64
	PandoraHashesLen  int64
	UnixTime          int64
}

// GetPendingHashes This is only for debug purpose
func (api *PublicFilterAPI) GetPendingHashes() (response *PendingHashesResponse, err error) {
	return api.backend.GetPendingHashes()
}

// ConfirmPanBlockHashes should be used to get the confirmation about known state of Pandora block hashes
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
		status, currentErr := api.backend.FetchPanBlockStatus(blockRequest.Slot, blockRequest.Hash)
		hash := blockRequest.Hash

		if nil != currentErr {
			log.Errorf("Invalid block in ConfirmPanBlockHashes: %v", err)
			response = nil
			err = currentErr

			return
		}

		if Skipped == status {
			hash = kv.EmptyHash
		}

		response = append(response, &BlockStatus{
			BlockHash: BlockHash{
				Slot: blockRequest.Slot,
				Hash: hash,
			},
			Status: status,
		})
	}

	respBytes, err := json.Marshal(response)

	if nil != err {
		log.WithField("err", err).Error("error unmarshaling response inConfirmVanBlockHashes")

		return
	}

	log.WithField("method", "ConfirmPanBlockHashes").
		WithField("request", request).
		WithField("response", response).
		WithField("jsonResponse", string(respBytes)).
		Info("Sending back ConfirmPanBlockHashes response")

	return
}

// ConfirmVanBlockHashes should be used to get the confirmation about known state of Vanguard block hashes
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
		status, currentErr := api.backend.FetchVanBlockStatus(blockRequest.Slot, blockRequest.Hash)
		hash := blockRequest.Hash

		if nil != currentErr {
			log.Errorf("Invalid block in ConfirmVanBlockHashes: %v", err)
			response = nil
			err = currentErr

			return
		}

		if Skipped == status {
			hash = kv.EmptyHash
		}

		response = append(response, &BlockStatus{
			BlockHash: BlockHash{
				Slot: blockRequest.Slot,
				Hash: hash,
			},
			Status: status,
		})
	}

	respBytes, err := json.Marshal(response)

	if nil != err {
		log.WithField("err", err).Error("error unmarshaling response inConfirmVanBlockHashes")

		return
	}

	log.WithField("method", "ConfirmVanBlockHashes").
		WithField("request", request).
		WithField("response", response).
		WithField("jsonResponse", string(respBytes)).
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
	//timeMismatch := time.Second * 6 <- this was proper
	timeMismatch := time.Second * 7

	go func() {
		consensusInfo := make(chan *generalTypes.MinimalEpochConsensusInfo)
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
			differ := currentEpoch.EpochStartTime - uint64(timeMismatch.Seconds())

			log.WithField("epoch", index).
				WithField("epochStartTime", currentEpoch.EpochStartTime).
				Info("sending already known consensus info to subscriber")
			err := notifier.Notify(rpcSub.ID, &generalTypes.MinimalEpochConsensusInfo{
				Epoch:            currentEpoch.Epoch,
				ValidatorList:    currentEpoch.ValidatorList,
				EpochStartTime:   differ,
				SlotTimeDuration: currentEpoch.SlotTimeDuration,
			})

			if nil != err {
				log.WithField("context", "already known epochs notification failure").Error(err)
			}
		}

		for {
			select {
			case currentEpoch := <-consensusInfo:
				// TODO: Remove it asap
				differ := currentEpoch.EpochStartTime - uint64(timeMismatch.Seconds())
				log.WithField("epoch", currentEpoch.Epoch).Info("sending consensus info to subscriber")
				err := notifier.Notify(rpcSub.ID, &generalTypes.MinimalEpochConsensusInfo{
					Epoch:            currentEpoch.Epoch,
					ValidatorList:    currentEpoch.ValidatorList,
					EpochStartTime:   differ,
					SlotTimeDuration: currentEpoch.SlotTimeDuration,
				})

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
