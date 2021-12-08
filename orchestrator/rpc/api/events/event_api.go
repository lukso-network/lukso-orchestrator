package events

import (
	"context"

	"github.com/ethereum/go-ethereum/rpc"
	generalTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	"github.com/pkg/errors"
)

// MinimalConsensusInfo
func (api *PublicFilterAPI) MinimalConsensusInfo(ctx context.Context, requestedEpoch uint64) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}
	rpcSub := notifier.CreateSubscription()

	go func() {

		batchSender := func(start, end uint64) error {
			epochInfos, err := api.backend.ConsensusInfoByEpochRange(start)
			if err != nil {
				log.WithError(err).Error("Some epoch infos are missing in db.")
				return errors.Wrap(err, "Missing epoch infos in db. Could not send over stream.")
			}
			latestFinalizedSlot := api.backend.LatestFinalizedSlot()
			for _, ei := range epochInfos {
				if err := notifier.Notify(rpcSub.ID, &generalTypes.MinimalEpochConsensusInfoV2{
					Epoch:            ei.Epoch,
					ValidatorList:    ei.ValidatorList,
					EpochStartTime:   ei.EpochStartTime,
					SlotTimeDuration: ei.SlotTimeDuration,
					FinalizedSlot:    latestFinalizedSlot,
				}); err != nil {
					log.WithField("start", start).
						WithField("end", end).
						WithError(err).
						Error("Failed to send epoch info. Could not send over stream.")
					return errors.Wrap(err, "Failed to send epoch info. Could not send over stream.")
				}
				log.WithField("epoch", ei.Epoch).WithField("latestFinalizedSlot", latestFinalizedSlot).
					Info("published epoch info to pandora")
			}
			return nil
		}

		startEpoch := requestedEpoch
		endEpoch := api.backend.LatestEpoch()
		if startEpoch <= endEpoch {
			log.WithField("startEpoch", startEpoch).WithField("endEpoch", endEpoch).Debug("Sending previous epoch infos to pandora")
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
						log.WithField("startEpoch", startEpoch).WithField("endEpoch", endEpoch).
							Debug("successfully published left over epoch infos")
						if err := batchSender(startEpoch, endEpoch); err != nil {
							return
						}
					}
					log.WithField("liveSyncEpoch", endEpoch+1).Debug("start publishing live epoch info to pandora")
				}

				err := notifier.Notify(rpcSub.ID, &generalTypes.MinimalEpochConsensusInfoV2{
					Epoch:            currentEpochInfo.Epoch,
					ValidatorList:    currentEpochInfo.ValidatorList,
					EpochStartTime:   currentEpochInfo.EpochStartTime,
					SlotTimeDuration: currentEpochInfo.SlotTimeDuration,
					ReorgInfo:        currentEpochInfo.ReorgInfo,
					FinalizedSlot:    currentEpochInfo.FinalizedSlot,
				})
				if nil != err {
					log.WithField("epoch", currentEpochInfo.Epoch).WithError(err).Error(
						"Failed to notify consensus info")
					return
				}

				log.WithField("epoch", currentEpochInfo.Epoch).WithField("latestFinalizedSlot", currentEpochInfo.FinalizedSlot).
					Info("published epoch info to pandora")

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
			blockStatus, err := api.backend.VerifiedSlotInfos(start)
			if err != nil {
				return err
			}

			for i := start; i <= end; i++ {
				if blockStatus[i] == nil {
					// invalid slot requested. maybe slot 0.
					continue
				}
				if err := notifier.Notify(rpcSub.ID, blockStatus[i]); err != nil {
					log.WithField("stepId", i).WithField("hash", blockStatus[i].Hash).WithError(err).
						Error("Failed to notify verified slot info. Could not send over stream.")
					return errors.Wrap(err, "Failed to notify verified slot info. Could not send over stream")
				}
			}
			return nil
		}

		startStepId := api.backend.StepId(request.Slot)
		endStepId := api.backend.LatestStepId()
		log.WithField("requestedSlot", request.Slot).WithField("startStepId", startStepId).WithField("endStepId", endStepId).
			Debug("Sending previous verified slot info status to pandora subscriber")

		if startStepId < endStepId {
			if err := batchSender(startStepId, endStepId); err != nil {
				log.WithError(err).Error("Failed to send slot status to pandora, closing event api for pandora subscriber")
				return
			}
		}

		slotInfoCh := make(chan *generalTypes.SlotInfoWithStatus)
		verifiedSlotInfoSub := api.events.SubscribeVerifiedSlotInfo(slotInfoCh)
		firstTime := true

		for {
			select {
			case slotInfoWithStatus := <-slotInfoCh:
				if firstTime {
					firstTime = false
					startStepId = endStepId
					endStepId = api.backend.LatestStepId()

					if startStepId+1 < endStepId {
						if err := batchSender(startStepId, endStepId); err != nil {
							log.WithError(err).Error("Failed to send slot status to pandora, closing go routine")
							return
						}
						log.WithField("startStepId", startStepId).WithField("endStepId", endStepId).
							Debug("Sent left over slot infos status to pandora")
					}
				}

				if err := notifier.Notify(rpcSub.ID, &generalTypes.BlockStatus{
					Hash:          slotInfoWithStatus.PandoraHeaderHash,
					Status:        slotInfoWithStatus.Status,
					FinalizedSlot: api.backend.LatestFinalizedSlot(),
				}); err != nil {
					log.WithField("hash", slotInfoWithStatus.PandoraHeaderHash).Error("Failed to notify slot info status. Could not send over stream.")
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
