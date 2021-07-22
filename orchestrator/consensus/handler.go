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

	vanShardInfo, _ := s.vanguardPendingShardingCache.Get(s.ctx, slot)
	if vanShardInfo != nil {
		return s.verifyShardingInfo(slot, vanShardInfo, headerInfo.Header)
	}

	log.WithField("slot", slot).Info("Waiting for pandora shard info")
	return nil
}

// processVanguardShardInfo
func (s *Service) processVanguardShardInfo(vanShardInfo *types.VanguardShardInfo) error {
	slot := vanShardInfo.Slot

	headerInfo, _ := s.pandoraPendingHeaderCache.Get(s.ctx, slot)
	if headerInfo != nil {
		// TODO- compare shard info and header
		return s.verifyShardingInfo(slot, vanShardInfo, headerInfo)
	}

	log.WithField("slot", slot).Info("Waiting for pandora header")
	return nil
}

// verifyShardingInfo
func (s *Service) verifyShardingInfo(slot uint64, vanShardInfo *types.VanguardShardInfo, header *eth1Types.Header) error {
	slotInfo := &types.SlotInfo{
		PandoraHeaderHash: header.Hash(),
		VanguardBlockHash: common.BytesToHash(vanShardInfo.BlockHash[:]),
	}
	status := CompareShardingInfo(header, vanShardInfo.ShardInfo)

	log.WithField("slot", slot).Debug("Consensus is established between pandora header and shard info")
	if status {
		// store verified slot info into verified slot info bucket
		if err := s.verifiedSlotInfoDB.SaveVerifiedSlotInfo(slot, slotInfo); err != nil {
			log.WithField("slot", slot).WithField(
				"slotInfo", fmt.Sprintf("%+v", slotInfo)).WithError(err).Error(
				"Failed to store verified slot info")
			return err
		}
		log.WithField("slot", slot).WithField(
			"slotInfo", fmt.Sprintf("%+v", slotInfo)).Info("Successfully verified sharding info")
	} else {
		// store invalid slot info into invalid slot info bucket
		if err := s.invalidSlotInfoDB.SaveInvalidSlotInfo(slot, slotInfo); err != nil {
			log.WithField("slot", slot).WithField(
				"slotInfo", fmt.Sprintf("%+v", slotInfo)).WithError(err).Error(
				"Failed to store invalid slot info")
			return err
		}
		log.WithField("slot", slot).WithField(
			"slotInfo", fmt.Sprintf("%+v", slotInfo)).Info("Invalid sharding info")
	}

	//removing previous cached slots which dont verified yet. By convention, they are skipped
	s.pandoraPendingHeaderCache.Remove(s.ctx, slot)
	s.vanguardPendingShardingCache.Remove(s.ctx, slot)
	return nil
}
