package consensus

import (
	"github.com/lukso-network/lukso-orchestrator/orchestrator/utils"

	"github.com/ethereum/go-ethereum/common"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

// processPandoraHeader method process incoming pandora shard header from pandora chain
// - First it checks the pandora header hash in verified shard info db. If it's already in db then it's already verified, so return nil
// - If it is not in verified db, then this method finds vanguard shard into pending cache.
// - If vanguard shard is already into pending cache, then calls insertIntoChain method to verify the sharding info and
// checks consecutiveness and trigger reorg if vanguard block's parent hash does not match with latest verified slot's hash
func (s *Service) processPandoraHeader(headerInfo *types.PandoraHeaderInfo) error {
	slot := headerInfo.Slot
	// short circuit check, if this header is already in verified sharding info db then send confirmation instantly
	if shardInfo := s.getShardingInfo(slot); shardInfo != nil {
		if len(shardInfo.Shards) > 0 && len(shardInfo.Shards[0].Blocks) > 0 && shardInfo.Shards[0].Blocks[0].HeaderRoot == headerInfo.Header.Hash() {
			log.WithField("shardInfo", shardInfo).Debug("Pandora shard header is already in verified shard info db")
			s.verifiedSlotInfoFeed.Send(&types.SlotInfoWithStatus{
				VanguardBlockHash: shardInfo.SlotInfo.BlockRoot,
				PandoraHeaderHash: shardInfo.Shards[0].Blocks[0].HeaderRoot,
				Status:            types.Verified,
			})
			return nil
		}
	}

	// first push the header into the cache.
	// it will update the cache if already present or enter a new info
	s.pendingInfoCache.PutPandoraHeader(slot, headerInfo.Header)

	// now mark it as we are making a decision on it
	err := s.pendingInfoCache.MarkInProgress(slot)
	if err != nil {
		return err
	}
	defer s.pendingInfoCache.MarkNotInProgress(slot)

	vanShardInfo, _ := s.pendingInfoCache.GetSlot(slot)
	if vanShardInfo == nil {
		return nil
	}

	return s.insertIntoChain(vanShardInfo.GetVanShard(), headerInfo.Header)
}

// processVanguardShardInfo
func (s *Service) processVanguardShardInfo(vanShardInfo *types.VanguardShardInfo) error {
	slot := vanShardInfo.Slot

	// short circuit check, if this header is already in verified sharding info db then send confirmation instantly
	if shardInfo := s.getShardingInfo(slot); shardInfo != nil {
		if shardInfo.SlotInfo != nil && shardInfo.SlotInfo.BlockRoot != common.BytesToHash(vanShardInfo.BlockHash) {
			log.WithField("shardInfo", shardInfo).Debug("Van header is already in verified shard info db")
			return nil
		}
	}

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
	if slotInfo.GetPanHeader() == nil {
		return nil
	}

	return s.insertIntoChain(vanShardInfo, slotInfo.GetPanHeader())
}

// insertIntoChain method
//	- verifies shard info and pandora header
//  - write into db
//  - send status to pandora chain
func (s *Service) insertIntoChain(vanShardInfo *types.VanguardShardInfo, header *eth1Types.Header) error {
	status, err := s.verifyShardingInfo(vanShardInfo, header)
	if err != nil {
		return err
	}

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
		//removing slot that is already verified
		s.pendingInfoCache.ForceDelSlot(vanShardInfo.Slot)

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
func (s *Service) verifyShardingInfo(vanShardInfo *types.VanguardShardInfo, header *eth1Types.Header) (bool, error) {
	// Here comparing sharding info with vanguard and pandora block's header
	if !compareShardingInfo(header, vanShardInfo.ShardInfo) {
		return false, nil
	}

	if verificationStatus, triggerReorg := s.verifyConsecutiveHashes(header, vanShardInfo); !verificationStatus {
		if triggerReorg {
			log.Info("Reorg triggered")
			if err := s.processReorg(common.BytesToHash(vanShardInfo.ParentHash)); err != nil {

				return false, err
			}
		}
	}

	return true, nil
}

func (s *Service) getShardingInfo(slot uint64) *types.MultiShardInfo {
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
