package consensus

import (
	"fmt"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/cache"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/utils"
	"github.com/pkg/errors"

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
	if shardInfo := s.getShardingInfo(slot); shardInfo != nil && shardInfo.NotNil() {
		if shardInfo.GetPanShardRoot() == headerInfo.Header.Hash() {
			log.WithField("shardInfo", shardInfo.FormattedStr()).Debug("Pandora shard header is already in verified shard info db")
			s.verifiedSlotInfoFeed.Send(&types.SlotInfoWithStatus{
				VanguardBlockHash: shardInfo.GetVanSlotRoot(),
				PandoraHeaderHash: shardInfo.GetPanShardRoot(),
				Status:            types.Verified,
			})
			return nil
		}
	}

	latestStepId := s.db.LatestStepID()
	latestShardInfo, err := s.db.VerifiedShardInfo(latestStepId)
	if err != nil {
		return err
	}

	if latestStepId > 1 && (latestShardInfo == nil || latestShardInfo.IsNil()) {
		return errors.New("nil latest shard info")
	}

	// first push the header into the cache.
	// it will update the cache if already present or enter a new info
	if err := s.panHeaderCache.Put(slot, &cache.PanCacheInsertParams{
		CurrentVerifiedHeader:  headerInfo.Header,
		LastVerifiedHeaderHash: latestShardInfo.GetPanShardRootBytes(),
	}); err != nil {
		log.WithError(err).WithField("blockNumber", headerInfo.Header.Number).
			WithField("slot", headerInfo.Slot).WithField("headerRoot", headerInfo.Header.Hash()).
			Info("Unknown parent in db and cache so discarding this pandora header")

		return nil
	}

	// now mark it as we are making a decision on it
	err = s.panHeaderCache.MarkInProgress(slot)
	if err != nil {
		return err
	}
	defer s.panHeaderCache.MarkNotInProgress(slot)

	vanShardInfo := s.vanShardCache.Get(slot)
	if vanShardInfo != nil && vanShardInfo.GetVanShard() != nil {
		return s.insertIntoChain(vanShardInfo.GetVanShard(), headerInfo.Header, latestShardInfo)
	}

	return nil
}

// processVanguardShardInfo
func (s *Service) processVanguardShardInfo(vanShardInfo *types.VanguardShardInfo) error {
	slot := vanShardInfo.Slot

	// short circuit check, if this header is already in verified sharding info db then send confirmation instantly
	if shardInfo := s.getShardingInfo(slot); shardInfo != nil && shardInfo.NotNil() {
		if shardInfo.GetVanSlotRoot() != common.BytesToHash(vanShardInfo.BlockHash) {
			log.WithField("shardInfo", shardInfo.FormattedStr()).Debug("Van header is already in verified shard info db")
			return nil
		}
	}

	latestStepId := s.db.LatestStepID()
	latestShardInfo, err := s.db.VerifiedShardInfo(latestStepId)
	if err != nil {
		return errors.Wrap(err, "DB is corrupted! Failed to retrieve latest shard info")
	}

	if latestStepId > 1 && (latestShardInfo == nil || latestShardInfo.IsNil()) {
		return errors.New("nil latest shard info")
	}

	// if reorg triggers here, orc will start processing reorg
	parentShardInfo, parentStepId, err := s.checkReorg(vanShardInfo, latestShardInfo, latestStepId)
	if err != nil {
		return errors.Wrap(err, "failed to check reorg!")
	}

	if parentShardInfo != nil {
		log.Info("Start processing reorg!")
		if err := s.processReorg(parentStepId, parentShardInfo); err != nil {
			log.WithError(err).Error("failed to process reorg!")
			return errors.Wrap(err, "failed to process reorg!")
		}
	}

	disableDelete := false
	if slot <= vanShardInfo.FinalizedSlot {
		// if slot number is less than finalized slot then initial sync is happening
		disableDelete = true
	}

	// first push the shardInfo into the cache.
	// it will update the cache if already present or enter a new info
	if err := s.vanShardCache.Put(slot, &cache.VanCacheInsertParams{
		DisableDelete:        disableDelete,
		CurrentShardInfo:     vanShardInfo,
		LastVerfiedShardRoot: latestShardInfo.GetVanSlotRootBytes(),
	}); err != nil {
		log.WithError(err).WithField("slot", vanShardInfo.Slot).WithField("blockRoot", common.BytesToHash(vanShardInfo.BlockHash)).
			Info("Unknown parent in db and cache so discarding this vanguard block")

		return nil
	}

	// now mark it as we are making a decision on it
	err = s.vanShardCache.MarkInProgress(slot)
	if err != nil {
		return err
	}
	defer s.vanShardCache.MarkNotInProgress(slot)

	pandoraHeaderInfo := s.panHeaderCache.Get(slot)
	if pandoraHeaderInfo != nil && pandoraHeaderInfo.GetPanHeader() != nil {
		return s.insertIntoChain(vanShardInfo, pandoraHeaderInfo.GetPanHeader(), latestShardInfo)
	}

	return nil
}

// insertIntoChain method
//	- verifies shard info and pandora header
//  - write into db
//  - send status to pandora chain
func (s *Service) insertIntoChain(
	vanShardInfo *types.VanguardShardInfo,
	header *eth1Types.Header,
	latestShardInfo *types.MultiShardInfo,
) error {

	confirmationStatus := &types.SlotInfoWithStatus{
		PandoraHeaderHash: header.Hash(),
		VanguardBlockHash: common.BytesToHash(vanShardInfo.BlockHash[:]),
		Status:            types.Invalid,
	}

	if compareShardingInfo(header, vanShardInfo.ShardInfo) && s.verifyShardInfo(latestShardInfo, header, vanShardInfo) {
		newShardInfo := utils.PrepareMultiShardData(vanShardInfo, header, TotalExecutionShardCount, ShardsPerVanBlock)
		// Write shard info into db
		if err := s.writeShardInfoInDB(newShardInfo); err != nil {
			return errors.Wrap(err, "failed to write shard info in db")
		}
		// write finalize info into db
		s.writeFinalizeInfo(vanShardInfo.FinalizedSlot, vanShardInfo.FinalizedEpoch)
		confirmationStatus.Status = types.Verified
		//removing slot that is already verified
		s.panHeaderCache.ForceDelSlot(vanShardInfo.Slot)
		s.vanShardCache.ForceDelSlot(vanShardInfo.Slot)

	}

	// sending confirmation status to pandora
	s.verifiedSlotInfoFeed.Send(confirmationStatus)
	return nil
}

func (s *Service) getShardingInfo(slot uint64) *types.MultiShardInfo {
	// Removing slot infos from verified slot info db
	stepId, err := s.db.GetStepIdBySlot(slot)
	if err != nil {
		return nil
	}

	shardInfo, err := s.db.VerifiedShardInfo(stepId)
	if err != nil {
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

	log.WithField("stepId", nextStepId).WithField("shardInfo", fmt.Sprintf("%+v", shardInfo)).Info("Inserted sharding info into verified DB")
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
