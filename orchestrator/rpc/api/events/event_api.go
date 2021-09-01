package events

import (
	"context"
	"github.com/ethereum/go-ethereum/rpc"
	generalTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	"github.com/pkg/errors"
)

var lastSendSlot uint64

// SteamConfirmedPanBlockHashes
func (api *PublicFilterAPI) SteamConfirmedPanBlockHashes(
	ctx context.Context,
	request *BlockHash,
) (*rpc.Subscription, error) {

	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}
	rpcSub := notifier.CreateSubscription()

	go func() {

		batchSender := func(start, end uint64) error {
			slotInfos := api.backend.VerifiedSlotInfos(start)

			for i := start; i <= end; i++ {
				if err := notifier.Notify(rpcSub.ID, &generalTypes.BlockStatus{
					Hash:   slotInfos[i].PandoraHeaderHash,
					Status: generalTypes.Verified,
				}); err != nil {
					log.WithField("start", start).
						WithField("end", end).
						WithError(err).
						Error("Failed to notify verified slot info. Could not send over stream.")
					return errors.Wrap(err, "Failed to notify verified slot info. Could not send over stream")
				}
			}
			return nil
		}

		startSlot := request.Slot
		endSlot := api.backend.LatestVerifiedSlot()

		if startSlot < endSlot {
			if err := batchSender(startSlot, endSlot); err != nil {
				return
			}
		}

		slotInfoCh := make(chan *generalTypes.SlotInfo)
		verifiedSlotInfoSub := api.events.SubscribeVerifiedSlotInfo(slotInfoCh)
		firstTime := true

		for {
			select {
			case verifiedSlotInfo := <-slotInfoCh:
				if firstTime {
					firstTime = false
					startSlot = endSlot
					endSlot = api.backend.LatestVerifiedSlot()
					if startSlot+1 < endSlot {
						if err := batchSender(startSlot, endSlot); err != nil {
							return
						}
					}
				}

				if err := notifier.Notify(rpcSub.ID, &generalTypes.BlockStatus{
					Hash:   verifiedSlotInfo.PandoraHeaderHash,
					Status: generalTypes.Verified,
				}); err != nil {
					log.WithField("hash", verifiedSlotInfo.PandoraHeaderHash).
						Error("Failed to notify verified slot info. Could not send over stream.")
					return
				}

			case <-rpcSub.Err():
				log.Info("Unsubscribing registered subscriber from SteamConfirmedPanBlockHashes")
				verifiedSlotInfoSub.Unsubscribe()
				return
			case <-notifier.Closed():
				log.Info("Closing notifier. Unsubscribing registered subscriber from SteamConfirmedPanBlockHashes")
				verifiedSlotInfoSub.Unsubscribe()
				return
			}
		}
	}()

	return rpcSub, nil
}
