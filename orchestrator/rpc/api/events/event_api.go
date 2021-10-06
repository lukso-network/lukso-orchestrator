package events

import (
	"context"

	"github.com/ethereum/go-ethereum/rpc"
	generalTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	"github.com/pkg/errors"
)

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
				log.WithField("slot", i).WithField("slotInfo", slotInfos[i]).Debug("sending verifiedInfo to pandora batchsender")
				if slotInfos[i] == nil {
					// invalid slot requested. maybe slot 0.
					continue
				}
				log.WithField("hash", slotInfos[i].PandoraHeaderHash).Debug("sending verifiedInfo to pandora batchsender")
				sendingInfo := &generalTypes.BlockStatus{
					Hash:   slotInfos[i].PandoraHeaderHash,
					Status: generalTypes.Verified,
				}
				log.WithField("info", *sendingInfo).Debug("Sending pendingness status to pandora")
				if err := notifier.Notify(rpcSub.ID, sendingInfo); err != nil {
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
		log.WithField("startSlot", startSlot).WithField("endSlot", endSlot).
			Debug("received information from pandora")

		if startSlot < endSlot {
			if err := batchSender(startSlot, endSlot); err != nil {
				return
			}
		}

		slotInfoCh := make(chan *generalTypes.SlotInfoWithStatus)
		verifiedSlotInfoSub := api.events.SubscribeVerifiedSlotInfo(slotInfoCh)
		firstTime := true

		for {
			select {
			case slotInfoWithStatus := <-slotInfoCh:
				log.WithField("hash", slotInfoWithStatus.PandoraHeaderHash).Debug("Sending slot info status to pandora")
				if firstTime {
					firstTime = false
					startSlot = endSlot
					endSlot = api.backend.LatestVerifiedSlot()
					log.WithField("startSlot", startSlot).WithField("endSlot", endSlot).Debug("for the first time")
					if startSlot+1 < endSlot {
						if err := batchSender(startSlot, endSlot); err != nil {
							return
						}
					}
				}

				if err := notifier.Notify(rpcSub.ID, &generalTypes.BlockStatus{
					Hash:   slotInfoWithStatus.PandoraHeaderHash,
					Status: slotInfoWithStatus.Status,
				}); err != nil {
					log.WithField("hash", slotInfoWithStatus.PandoraHeaderHash).
						Error("Failed to notify slot info status. Could not send over stream.")
					return
				}
			case rpcErr := <-rpcSub.Err():
				log.WithField("error", rpcErr).Info("Unsubscribing registered subscriber from SteamConfirmedPanBlockHashes")
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
