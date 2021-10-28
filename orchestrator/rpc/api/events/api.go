package events

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/rpc"
	generalTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	"github.com/pkg/errors"
)

var lastSendEpoch uint64

type Backend interface {
	ConsensusInfoByEpochRange(fromEpoch uint64) []*generalTypes.MinimalEpochConsensusInfoV2
	SubscribeNewEpochEvent(chan<- *generalTypes.MinimalEpochConsensusInfoV2) event.Subscription
	GetSlotStatus(ctx context.Context, slot uint64, hash common.Hash, requestFrom bool) generalTypes.Status
	LatestEpoch() uint64
	SubscribeNewVerifiedSlotInfoEvent(chan<- *generalTypes.SlotInfoWithStatus) event.Subscription
	VerifiedSlotInfos(fromSlot uint64) map[uint64]*generalTypes.SlotInfo
	LatestVerifiedSlot() uint64
	PendingPandoraHeaders() []*eth1Types.Header
}

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
	Status generalTypes.Status
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
		status := api.backend.GetSlotStatus(ctx, req.Slot, req.Hash, true)
		log.WithField("slot", req.Slot).WithField("status", status).WithField(
			"api", "ConfirmPanBlockHashes").Debug("status of the requested slot")
		hash := req.Hash
		res = append(res, &BlockStatus{
			BlockHash: BlockHash{
				Slot: req.Slot,
				Hash: hash,
			},
			Status: status,
		})
	}
	return res, nil
}

// ConfirmVanBlockHashes should be used to get the confirmation about known state of Vanguard block hashes
func (api *PublicFilterAPI) ConfirmVanBlockHashes(
	ctx context.Context,
	requests []*BlockHash,
) (response []*BlockStatus, err error) {
	if len(requests) < 1 {
		err := fmt.Errorf("invalid request")
		return nil, err
	}
	res := make([]*BlockStatus, 0)
	for _, req := range requests {
		status := api.backend.GetSlotStatus(ctx, req.Slot, req.Hash, false)
		log.WithField("slot", req.Slot).WithField("status", status).WithField(
			"api", "ConfirmVanBlockHashes").Debug("Status of the requested slot")
		hash := req.Hash
		res = append(res, &BlockStatus{
			BlockHash: BlockHash{
				Slot: req.Slot,
				Hash: hash,
			},
			Status: status,
		})
	}
	return res, nil
}

// MinimalConsensusInfo
func (api *PublicFilterAPI) MinimalConsensusInfo(ctx context.Context, requestedEpoch uint64) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}
	rpcSub := notifier.CreateSubscription()

	go func() {

		batchSender := func(start, end uint64) error {
			epochInfos := api.backend.ConsensusInfoByEpochRange(start)
			for _, ei := range epochInfos {
				if err := notifier.Notify(rpcSub.ID, &generalTypes.MinimalEpochConsensusInfoV2{
					Epoch:            ei.Epoch,
					ValidatorList:    ei.ValidatorList,
					EpochStartTime:   ei.EpochStartTime,
					SlotTimeDuration: ei.SlotTimeDuration,
				}); err != nil {
					log.WithField("start", start).
						WithField("end", end).
						WithError(err).
						Error("Failed to send epoch info. Could not send over stream.")
					return errors.Wrap(err, "Failed to send epoch info. Could not send over stream.")
				}
			}
			return nil
		}

		startEpoch := requestedEpoch
		endEpoch := api.backend.LatestEpoch()
		if startEpoch <= endEpoch {
			if err := batchSender(startEpoch, endEpoch); err != nil {
				return
			}
		}

		consensusInfo := make(chan *generalTypes.MinimalEpochConsensusInfoV2)
		consensusInfoSub := api.events.SubscribeConsensusInfo(consensusInfo, requestedEpoch)
		firstTime := true

		for {
			select {
			case currentEpochInfo := <-consensusInfo:
				log.WithField("epoch", currentEpochInfo.Epoch).
					WithField("epochStartTime", currentEpochInfo.EpochStartTime).
					Info("Sending consensus info to subscriber")

				if firstTime {
					firstTime = false
					startEpoch = endEpoch
					endEpoch = api.backend.LatestEpoch()

					if startEpoch+1 < endEpoch {
						if err := batchSender(startEpoch, endEpoch); err != nil {
							return
						}
					}
				}

				err := notifier.Notify(rpcSub.ID, &generalTypes.MinimalEpochConsensusInfoV2{
					Epoch:            currentEpochInfo.Epoch,
					ValidatorList:    currentEpochInfo.ValidatorList,
					EpochStartTime:   currentEpochInfo.EpochStartTime,
					SlotTimeDuration: currentEpochInfo.SlotTimeDuration,
					ReorgInfo:        currentEpochInfo.ReorgInfo,
				})
				if nil != err {
					log.WithField("epoch", currentEpochInfo.Epoch).WithError(err).Error(
						"Failed to notify consensus info")
					return
				}
			case <-rpcSub.Err():
				log.Info("Unsubscribing registered pandora client")
				consensusInfoSub.Unsubscribe()
				return
			case <-notifier.Closed():
				log.Info("Closing notifier. Unsubscribing registered pandora subscriber")
				consensusInfoSub.Unsubscribe()
				return
			}
		}
	}()

	return rpcSub, nil
}
