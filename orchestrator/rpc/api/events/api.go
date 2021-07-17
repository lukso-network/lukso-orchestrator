package events

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/rpc"
	generalTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	"time"
)

type Backend interface {
	ConsensusInfoByEpochRange(fromEpoch uint64) []*generalTypes.MinimalEpochConsensusInfo
	SubscribeNewEpochEvent(chan<- *generalTypes.MinimalEpochConsensusInfo) event.Subscription
	GetSlotStatus(ctx context.Context, slot uint64, requestType bool) Status
}

type Status string

const (
	Pending  Status = "Pending"
	Verified Status = "Verified"
	Invalid  Status = "Invalid"
	Skipped  Status = "Skipped"
	Unknown  Status = "Unknown"
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

// ConfirmPanBlockHashes should be used to get the confirmation about known state of Pandora block hashes
func (api *PublicFilterAPI) ConfirmPanBlockHashes(
	ctx context.Context,
	requests []*BlockHash,
) ([]*BlockStatus, error) {
	if len(requests) < 1 {
		err := fmt.Errorf("invalid request")
		return nil, err
	}
	res := make([]*BlockStatus, 0)
	for _, req := range requests {
		status := api.backend.GetSlotStatus(ctx, req.Slot, true)
		hash := req.Hash
		res = append(res, &BlockStatus{
			BlockHash: BlockHash{
				Slot: req.Slot,
				Hash: hash,
			},
			Status: status,
		})
	}

	log.WithField("method", "ConfirmPanBlockHashes").WithField(
		"request", requests).WithField("response", res).Debug(
		"Sending back ConfirmPanBlockHashes response")
	return res, nil
}

// ConfirmVanBlockHashes should be used to get the confirmation about known state of Vanguard block hashes
func (api *PublicFilterAPI) ConfirmVanBlockHashes(
	ctx context.Context,
	requests []*BlockHash,
) (response []*BlockStatus, err error) {
	if len(requests) < 1 {
		err := fmt.Errorf("request has empty slice")
		return nil, err
	}
	res := make([]*BlockStatus, 0)
	for _, req := range requests {
		status := api.backend.GetSlotStatus(ctx, req.Slot, false)
		hash := req.Hash
		res = append(res, &BlockStatus{
			BlockHash: BlockHash{
				Slot: req.Slot,
				Hash: hash,
			},
			Status: status,
		})
	}

	log.WithField("method", "ConfirmVanBlockHashes").WithField(
		"request", requests).WithField("response", res).Debug(
		"Sending back ConfirmVanBlockHashes response")
	return res, nil
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
	//timeMismatch := time.Second * 3
	go func() {
		consensusInfo := make(chan *generalTypes.MinimalEpochConsensusInfo)
		consensusInfoSub := api.events.SubscribeConsensusInfo(consensusInfo, epoch)

		log.WithField("fromEpoch", epoch).Info("registered new subscriber for consensus info")
		if len(alreadyKnownEpochs) < 1 {
			log.WithField("fromEpoch", epoch).Info("there are no already known epochs, try to fetch lowest")
		}

		for index, currentEpoch := range alreadyKnownEpochs {
			// TODO: Remove it ASAP. This should not be that way
			//differ := currentEpoch.EpochStartTime - uint64(timeMismatch.Seconds())
			log.WithField("epoch", index).WithField("epochStartTime", currentEpoch.EpochStartTime).Info(
				"sending already known consensus info to subscriber")
			err := notifier.Notify(rpcSub.ID, &generalTypes.MinimalEpochConsensusInfo{
				Epoch:            currentEpoch.Epoch,
				ValidatorList:    currentEpoch.ValidatorList,
				EpochStartTime:   currentEpoch.EpochStartTime,
				SlotTimeDuration: currentEpoch.SlotTimeDuration,
			})
			if nil != err {
				log.WithError(err).Error("Failed to notify already known consensus infos")
			}
		}

		for {
			select {
			case currentEpoch := <-consensusInfo:
				// TODO: Remove it asap
				//differ := currentEpoch.EpochStartTime - uint64(timeMismatch.Seconds())
				log.WithField("epoch", currentEpoch.Epoch).WithField(
					"epochStartTime", currentEpoch.EpochStartTime).Info(
					"sending consensus info to subscriber")
				err := notifier.Notify(rpcSub.ID, &generalTypes.MinimalEpochConsensusInfo{
					Epoch:            currentEpoch.Epoch,
					ValidatorList:    currentEpoch.ValidatorList,
					EpochStartTime:   currentEpoch.EpochStartTime,
					SlotTimeDuration: currentEpoch.SlotTimeDuration,
				})
				if nil != err {
					log.WithField("epoch", currentEpoch.Epoch).WithError(err).Error(
						"Failed to notify consensus info")
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
