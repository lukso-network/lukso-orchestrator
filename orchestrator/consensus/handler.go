package consensus

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

// processPandoraHeader
func (s *Service) processPandoraHeader(headerInfo *types.PandoraHeaderInfo) error {
	slot := headerInfo.Slot
	// first push the header into the cache.
	// it will update the cache if already present or enter a new info
	s.pendingInfoCache.PutPandoraHeader(slot, headerInfo.Header)

	// now mark it as we are making a decision on it
	err := s.pendingInfoCache.MarkInProgress(slot)
	if err != nil {
		return err
	}
	defer s.pendingInfoCache.MarkNotInProgress(slot)

	slotInfo, _ := s.pendingInfoCache.GetSlot(slot)
	if slotInfo != nil {
		if shardInfo := slotInfo.GetVanShard(); shardInfo != nil {
			// vanguard sharding info is present for this slot. So verify it
			return s.verifyShardingInfo(slot, shardInfo, headerInfo.Header)
		}
	}
	return nil
}

// processVanguardShardInfo
func (s *Service) processVanguardShardInfo(vanShardInfo *types.VanguardShardInfo) error {
	slot := vanShardInfo.Slot
	// first push the shardInfo into the cache.
	// it will update the cache if already present or enter a new info
	s.pendingInfoCache.PutVanguardShardingInfo(slot, vanShardInfo, vanShardInfo.IsSyncing)
	// now mark it as we are making a decision on it
	err := s.pendingInfoCache.MarkInProgress(slot)
	if err != nil {
		return err
	}
	defer s.pendingInfoCache.MarkNotInProgress(slot)

	slotInfo, _ := s.pendingInfoCache.GetSlot(slot)
	if slotInfo != nil {
		if header := slotInfo.GetPanHeader(); header != nil {
			// Header info is present for this slot. So verify it
			return s.verifyShardingInfo(slot, vanShardInfo, header)
		}
	}
	return nil
}

// verifyShardingInfo
func (s *Service) verifyShardingInfo(slot uint64, vanShardInfo *types.VanguardShardInfo, header *eth1Types.Header) error {
	slotInfo := &types.SlotInfo{
		PandoraHeaderHash: header.Hash(),
		VanguardBlockHash: common.BytesToHash(vanShardInfo.BlockHash[:]),
	}
	status := CompareShardingInfo(header, vanShardInfo.ShardInfo)
	slotInfoWithStatus := &types.SlotInfoWithStatus{
		PandoraHeaderHash: header.Hash(),
		VanguardBlockHash: common.BytesToHash(vanShardInfo.BlockHash[:]),
	}
	if !status {
		// store invalid slot info into invalid slot info bucket
		if err := s.invalidSlotInfoDB.SaveInvalidSlotInfo(slot, slotInfo); err != nil {
			log.WithField("slot", slot).WithField(
				"slotInfo", fmt.Sprintf("%+v", slotInfo)).WithError(err).Error(
				"Failed to store invalid slot info")
			return err
		}
		slotInfoWithStatus.Status = types.Invalid
		log.WithField("slot", slot).Info("Invalid sharding info")
		// sending verified slot info to rpc service
		s.verifiedSlotInfoFeed.Send(slotInfoWithStatus)
		return nil
	}

	// store verified slot info into verified slot info bucket
	if err := s.verifiedSlotInfoDB.SaveVerifiedSlotInfo(slot, slotInfo); err != nil {
		log.WithField("slot", slot).WithField(
			"slotInfo", fmt.Sprintf("%+v", slotInfo)).WithError(err).Error("Failed to store verified slot info")
		return err
	}

	// storing latest verified slot into db
	if err := s.verifiedSlotInfoDB.SaveLatestVerifiedSlot(s.ctx, slot); err != nil {
		log.WithError(err).Error("Failed to store latest verified slot")
	}

	// storing latest verified pandora header hash into db
	if err := s.verifiedSlotInfoDB.SaveLatestVerifiedHeaderHash(slotInfo.PandoraHeaderHash); err != nil {
		log.WithError(err).Error("Failed to store latest verified slot")
	}

	// Storing latest finalized slot and epoch
	if s.verifiedSlotInfoDB.LatestLatestFinalizedEpoch() < vanShardInfo.FinalizedEpoch {
		if err := s.verifiedSlotInfoDB.SaveLatestFinalizedSlot(vanShardInfo.FinalizedSlot); err != nil {
			log.WithError(err).Warn("Failed to store new finalized info")
		}

		if err := s.verifiedSlotInfoDB.SaveLatestFinalizedEpoch(vanShardInfo.FinalizedEpoch); err != nil {
			log.WithError(err).Warn("Failed to store new finalized epoch")
		}
		log.WithField("newFinalizedSlot", vanShardInfo.FinalizedSlot).
			WithField("newFinalizedEpoch", vanShardInfo.FinalizedEpoch).Debug("Saved latest finalized info")
	}

	slotInfoWithStatus.Status = types.Verified
	//removing slot that is already verified
	s.pendingInfoCache.ForceDelSlot(slot)
	log.WithField("slot", slot).Info("Successfully verified sharding info")
	// sending verified slot info to rpc service
	s.verifiedSlotInfoFeed.Send(slotInfoWithStatus)
	return nil
}

func (s *Service) reorgDB(revertSlot uint64) error {
	// Removing slot infos from verified slot info db
	if err := s.verifiedSlotInfoDB.RemoveRangeVerifiedInfo(revertSlot+1, s.verifiedSlotInfoDB.LatestSavedVerifiedSlot()); err != nil {
		log.WithError(err).Error("found error while reverting orchestrator database in reorg phase")
		return err
	}

	if err := s.verifiedSlotInfoDB.UpdateVerifiedSlotInfo(revertSlot); err != nil {
		log.WithError(err).Error("failed to update latest verified slot info in reorg phase")
		return err
	}
	return nil
}
