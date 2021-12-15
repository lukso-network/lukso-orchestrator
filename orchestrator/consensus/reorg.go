package consensus

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"github.com/pkg/errors"
	"sync/atomic"
)

// processReorg
func (s *Service) processReorg(parentVanBlkHash common.Hash) error {
	atomic.StoreUint32(&s.reorgInProgress, 1)
	defer func() {
		atomic.StoreUint32(&s.reorgInProgress, 0)
	}()

	finalizedSlot := s.db.FinalizedSlot()
	finalizedStepId, err := s.db.GetStepIdBySlot(finalizedSlot)
	if err != nil {
		log.WithError(err).WithField("finalizedSlot", finalizedSlot).WithField("latestFinalizedStepId", finalizedStepId).
			Error("Could not found step id from DB")
		return err
	}

	latestStepId := s.db.LatestStepID()
	parentShardInfo, err := s.db.FindAncestor(latestStepId, finalizedStepId, parentVanBlkHash)
	if err != nil || parentShardInfo == nil {
		log.WithField("finalizedSlot", finalizedSlot).WithField("latestFinalizedStepId", finalizedStepId).
			Error("Could not found parent shard info from DB")
		return errors.Wrap(errUnknownParent, "Failed to process reorg")
	}

	if len(parentShardInfo.Shards) == 0 || len(parentShardInfo.Shards[0].Blocks) == 0 {
		log.WithField("finalizedSlot", finalizedSlot).WithField("latestFinalizedStepId", finalizedStepId).
			Error("Invalid length of shards in parent shard info")
		return errors.Wrap(errUnknownParent, "Failed to process reorg")
	}

	if err := s.db.RemoveShardingInfos(finalizedStepId + 1); err != nil {
		log.WithError(err).Error("Could not revert shard info from DB during reorg")
		return err
	}

	if err := s.db.SaveLatestStepID(finalizedStepId); err != nil {
		log.WithError(err).Error("Could not store latest step id during reorg")
		return err
	}

	s.pendingInfoCache.Purge()

	// Publish reorg info for rpc service for notifying pandora client
	s.reorgInfoFeed.Send(&types.Reorg{
		VanParentHash: parentShardInfo.SlotInfo.BlockRoot.Bytes(),
		PanParentHash: parentShardInfo.Shards[0].Blocks[0].HeaderRoot.Bytes(),
	})
	return nil
}
