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
	s.pandoraPendingHeaderCache.Put(s.ctx, slot, headerInfo.Header)
	vanShardInfo, _ := s.vanguardPendingShardingCache.Get(s.ctx, slot)
	if vanShardInfo != nil {
		return s.verifyShardingInfo(slot, vanShardInfo, headerInfo.Header)
	}
	return nil
}

// processVanguardShardInfo
func (s *Service) processVanguardShardInfo(vanShardInfo *types.VanguardShardInfo) error {
	slot := vanShardInfo.Slot
	s.vanguardPendingShardingCache.Put(s.ctx, slot, vanShardInfo)
	headerInfo, _ := s.pandoraPendingHeaderCache.Get(s.ctx, slot)
	if headerInfo != nil {
		return s.verifyShardingInfo(slot, vanShardInfo, headerInfo)
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
	if status {
		// store verified slot info into verified slot info bucket
		if err := s.verifiedSlotInfoDB.SaveVerifiedSlotInfo(slot, slotInfo); err != nil {
			log.WithField("slot", slot).WithField(
				"slotInfo", fmt.Sprintf("%+v", slotInfo)).WithError(err).Error(
				"Failed to store verified slot info")
			return err
		}
		if err := s.verifiedSlotInfoDB.SaveLatestVerifiedSlot(s.ctx); err != nil {
			log.WithError(err).Error("Failed to store latest verified slot")
		}
		if err := s.verifiedSlotInfoDB.SaveLatestVerifiedHeaderHash(); err != nil {
			log.WithError(err).Error("Failed to store latest verified slot")
		}
		slotInfoWithStatus.Status = types.Verified
		//removing previous cached slots which dont verified yet. By convention, they are skipped
		s.pandoraPendingHeaderCache.Remove(s.ctx, slot)
		s.vanguardPendingShardingCache.Remove(s.ctx, slot)
		log.WithField("slot", slot).Info("Successfully verified sharding info")
	} else {
		// store invalid slot info into invalid slot info bucket
		if err := s.invalidSlotInfoDB.SaveInvalidSlotInfo(slot, slotInfo); err != nil {
			log.WithField("slot", slot).WithField(
				"slotInfo", fmt.Sprintf("%+v", slotInfo)).WithError(err).Error(
				"Failed to store invalid slot info")
			return err
		}
		slotInfoWithStatus.Status = types.Invalid
		log.WithField("slot", slot).Info("Invalid sharding info")
	}
	// sending verified slot info to rpc service
	s.verifiedSlotInfoFeed.Send(slotInfoWithStatus)
	return nil
}
