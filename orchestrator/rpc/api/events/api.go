package events

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	generalTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
)

var lastSendEpoch uint64

type Backend interface {
	ConsensusInfoByEpochRange(fromEpoch uint64) ([]*generalTypes.MinimalEpochConsensusInfoV2, error)
	SubscribeNewEpochEvent(chan<- *generalTypes.MinimalEpochConsensusInfoV2) event.Subscription
	GetSlotStatus(ctx context.Context, slot uint64, hash common.Hash, requestFrom bool) generalTypes.Status
	LatestEpoch() uint64
	SubscribeNewVerifiedSlotInfoEvent(chan<- *generalTypes.SlotInfoWithStatus) event.Subscription
	VerifiedSlotInfos(fromSlot uint64) (map[uint64]*generalTypes.BlockStatus, error)
	LatestFinalizedSlot() uint64
	StepId(slot uint64) uint64
	LatestStepId() uint64
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
