package consensus

import (
	"github.com/lukso-network/lukso-orchestrator/orchestrator/utils"

	"github.com/ethereum/go-ethereum/common"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

// processPandoraHeader
func (s *Service) processPandoraHeader(headerInfo *types.PandoraHeaderInfo) error {
	slot := headerInfo.Slot

	// short circuit check, if this header is already in verified sharding info db then send confirmation instantly
	if shardInfo := s.getShardingInfoInDB(slot); shardInfo != nil {
		if len(shardInfo.Shards) > 0 {
			blocks := shardInfo.Shards[0].Blocks
			if len(blocks) > 0 && blocks[0].HeaderRoot == headerInfo.Header.Hash() {
				log.WithField("shardInfo", shardInfo).Debug("Pan header is already in verified shard info db")

				s.verifiedSlotInfoFeed.Send(&types.SlotInfoWithStatus{
					VanguardBlockHash: shardInfo.SlotInfo.BlockRoot,
					PandoraHeaderHash: blocks[0].HeaderRoot,
					Status:            types.Verified,
				})

				return nil
			}
		}
	}

	// retrieving sharding
	s.pandoraPendingHeaderCache.Put(s.ctx, slot, headerInfo.Header)
	vanShardInfo, _ := s.vanguardPendingShardingCache.Get(s.ctx, slot)
	if vanShardInfo == nil {
		return nil
	}

	return s.insertIntoChain(vanShardInfo, headerInfo.Header)
}

// processVanguardShardInfo
func (s *Service) processVanguardShardInfo(vanShardInfo *types.VanguardShardInfo) error {
	slot := vanShardInfo.Slot

	// short circuit check, if this header is already in verified sharding info db then send confirmation instantly
	if shardInfo := s.getShardingInfoInDB(slot); shardInfo != nil {
		if shardInfo.SlotInfo != nil && shardInfo.SlotInfo.BlockRoot != common.BytesToHash(vanShardInfo.BlockHash) {
			log.WithField("shardInfo", shardInfo).Debug("Van header is already in verified shard info db")
			return nil
		}
	}

	s.vanguardPendingShardingCache.Put(s.ctx, slot, vanShardInfo)
	header, _ := s.pandoraPendingHeaderCache.Get(s.ctx, slot)
	if header == nil {
		return nil
	}

	return s.insertIntoChain(vanShardInfo, header)
}

// insertIntoChain method
//	- verifies shard info and pandora header
//  - write into db
//  - send status to pandora chain
func (s *Service) insertIntoChain(vanShardInfo *types.VanguardShardInfo, header *eth1Types.Header) error {
	status := s.verifyShardingInfo(vanShardInfo, header)
	confirmationStatus := &types.SlotInfoWithStatus{
		PandoraHeaderHash: header.Hash(),
		VanguardBlockHash: common.BytesToHash(vanShardInfo.BlockHash[:]),
	}

	if status {
		shardInfo := utils.PrepareMultiShardData(vanShardInfo, header, TotalExecutionShardCount, ShardsPerVanBlock)
		// Write shard info into db
		if err := s.writeShardInfoInDB(shardInfo); err != nil {
			return err
		}
		// write finalize info into db
		s.writeFinalizeInfo(vanShardInfo.FinalizedSlot, vanShardInfo.FinalizedEpoch)
		confirmationStatus.Status = types.Verified

	} else {
		confirmationStatus.Status = types.Invalid
		log.WithField("slot", vanShardInfo.Slot).Info("Invalid sharding info")
	}

	// sending confirmation status to pandora
	s.verifiedSlotInfoFeed.Send(confirmationStatus)
	return nil
}

// verifyShardingInfo checks
//	- sharding info between incoming vanguard and pandora sharding infos
// 	- checks consecutive parent hash of vanguard and pandora sharding hash
// 	- if parent hash of vanguard block does not match with latest verified slot then trigger reorg
func (s *Service) verifyShardingInfo(vanShardInfo *types.VanguardShardInfo, header *eth1Types.Header) bool {
	// Here comparing sharding info with vanguard and pandora block's header
	if !CompareShardingInfo(header, vanShardInfo.ShardInfo) {
		return false
	}

	// TODO- Checking block consecutive of pandora sharding info and vanguard slot info
	// TODO- Trigger reorg if slot consecutive fails

	return true
}

func (s *Service) getShardingInfoInDB(slot uint64) *types.MultiShardInfo {
	// Removing slot infos from verified slot info db
	stepId, err := s.db.GetStepIdBySlot(slot)
	if err != nil {
		log.WithError(err).WithField("slot", slot).Error("Could not found step id from DB")
		return nil
	}

	shardInfo, err := s.db.VerifiedShardInfo(stepId)
	if err != nil {
		log.WithError(err).WithField("stepId", stepId).Error("Could not found shard info from DB during reorg")
		return nil
	}

	if shardInfo == nil {
		log.WithField("stepId", stepId).Debug("Could not found shard info from DB during reorg")
		return nil
	}

	return shardInfo
}

// WriteShardInfoInDB method converts vanShardInfo and panHeader to multiShardingInfo
// Store multiShardingInfo into db
// Update stepId into db
func (s *Service) writeShardInfoInDB(shardInfo *types.MultiShardInfo) error {
	latestStepId := s.db.LatestStepID()
	nextStepId := latestStepId + 1
	if err := s.db.SaveVerifiedShardInfo(nextStepId, shardInfo); err != nil {
		return err
	}

	if err := s.db.SaveLatestStepID(nextStepId); err != nil {
		return err
	}

	if err := s.db.SaveSlotStepIndex(shardInfo.SlotInfo.Slot, nextStepId); err != nil {
		return err
	}

	log.WithField("stepId", nextStepId).WithField("shardInfo", shardInfo).Info("Inserted sharding info into verified DB")
	return nil
}

// writeFinalizeInfo method store latest finalize slot and epoch if needed
func (s *Service) writeFinalizeInfo(finalizeSlot, finalizeEpoch uint64) {
	curFinalizeSlot := s.db.FinalizedSlot()
	if finalizeSlot > curFinalizeSlot {
		if err := s.db.SaveFinalizedSlot(finalizeSlot); err != nil {
			log.WithError(err).Warn("Failed to store new finalized info")
		}
	}

	curFinalizeEpoch := s.db.FinalizedEpoch()
	if finalizeEpoch > curFinalizeEpoch {
		if err := s.db.SaveFinalizedEpoch(finalizeEpoch); err != nil {
			log.WithError(err).Warn("Failed to store new finalized epoch")
		}
	}
}

// reorgDB
func (s *Service) reorgDB(revertSlot uint64) error {
	// Removing slot infos from verified slot info db
	stepId, err := s.db.GetStepIdBySlot(revertSlot)
	if err != nil {
		log.WithError(err).WithField("revertSlot", revertSlot).Error("Could not found step id from DB during reorg")
		return err
	}

	shardInfo, err := s.db.VerifiedShardInfo(stepId)
	if err != nil {
		log.WithError(err).WithField("stepId", stepId).Error("Could not found shard info from DB during reorg")
		return err
	}

	if stepId > 0 && shardInfo != nil {
		if err := s.db.RemoveShardingInfos(stepId); err != nil {
			log.WithError(err).Error("Could not revert shard info from DB during reorg")
			return err
		}
	}

	if err := s.db.SaveLatestStepID(stepId); err != nil {
		log.WithError(err).Error("Could not store latest step id during reorg")
		return err
	}

	return nil
}
