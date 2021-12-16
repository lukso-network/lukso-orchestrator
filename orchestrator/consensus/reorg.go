package consensus

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"github.com/pkg/errors"
	"sync/atomic"
)

// checkReorg checks the incoming vanguard slot's consecutiveness with db head.
func (s *Service) checkReorg(
	vanShardInfo *types.VanguardShardInfo,
	latestVerifiedShardInfo *types.MultiShardInfo,
	latestStepId uint64,
) (*types.MultiShardInfo, uint64, error) {

	// when latest verified shard info is nil, just return nil
	if latestVerifiedShardInfo == nil || latestVerifiedShardInfo.IsNil() {
		return nil, 0, errors.New("nil latest shard info, reorg checking failed")
	}

	// if this slot is less than finalized slot then it does not check reorg
	if vanShardInfo.Slot <= vanShardInfo.FinalizedSlot {
		log.WithField("slot", vanShardInfo.Slot).
			WithField("finalizedSlot", vanShardInfo.FinalizedSlot).
			Info("Skipped reorg checking in initial-syncing")
		return nil, 0, nil
	}

	finalizedSlotInDB := s.db.FinalizedSlot()
	finalizedStepId, err := s.db.GetStepIdBySlot(finalizedSlotInDB)
	if err != nil {
		log.WithError(err).WithField("finalizedSlotInDB", finalizedSlotInDB).
			WithField("latestFinalizedStepId", finalizedStepId).
			Error("Could not found step id from DB")
		return nil, 0, errors.Wrap(err, "could not found step id from DB")
	}

	if finalizedSlotInDB > latestVerifiedShardInfo.SlotInfo.Slot {
		log.WithField("latestSlot", latestVerifiedShardInfo.SlotInfo.Slot).
			WithField("finalizedSlotInDB", finalizedSlotInDB).
			Info("Skipped reorg checking in initial-syncing")
		return nil, 0, nil
	}

	parentHash := common.BytesToHash(vanShardInfo.ParentHash)
	parentShardInfo, stepId, err := s.db.FindAncestor(latestStepId, finalizedStepId, parentHash)
	if err != nil {
		return nil, 0, errors.Wrapf(err,
			"Failed to find parent in verified db with slot %d and parentHash %v and stepId %d",
			vanShardInfo.Slot, parentHash, stepId)
	}

	// if parent shard info does not find in db
	if parentShardInfo == nil || parentShardInfo.IsNil() {
		if !s.vanShardCache.ContainsHash(parentHash.Bytes()) {
			log.WithField("parentHash", parentHash).WithField("slot", vanShardInfo.Slot).Info("Unknown parent")
			return nil, 0, errUnknownParent
		}
		return nil, 0, nil
	}

	// parent slot found in db now checking with latest verified slot
	// if they are mis-matched, then trigger reorg
	if !latestVerifiedShardInfo.DeepEqual(parentShardInfo) {
		log.WithField("latestShardInfo", latestVerifiedShardInfo.FormattedStr()).
			WithField("parentShardInfo", parentShardInfo.FormattedStr()).Info("Triggering reorg!")
		return parentShardInfo, stepId, nil
	}

	return nil, 0, nil
}

// processReorg
func (s *Service) processReorg(parentStepId uint64, parentShardInfo *types.MultiShardInfo) error {
	atomic.StoreUint32(&s.reorgInProgress, 1)
	defer func() {
		atomic.StoreUint32(&s.reorgInProgress, 0)
	}()

	if err := s.db.RemoveShardingInfos(parentStepId + 1); err != nil {
		log.WithError(err).Error("Could not revert shard info from DB during reorg")
		return err
	}

	if err := s.db.SaveLatestStepID(parentStepId); err != nil {
		log.WithError(err).Error("Could not store latest step id during reorg")
		return err
	}

	// Publish reorg info for rpc service for notifying pandora client
	s.reorgInfoFeed.Send(&types.Reorg{
		VanParentHash: parentShardInfo.GetVanSlotRootBytes(),
		PanParentHash: parentShardInfo.GetPanShardRootBytes(),
	})
	return nil
}
