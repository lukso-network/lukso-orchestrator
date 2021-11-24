package vanguardchain

import (
	"context"
	"errors"

	"github.com/lukso-network/lukso-orchestrator/shared/types"
	eth "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/proto/eth/v1alpha1/wrapper"
)

// onNewConsensusInfo :
//	- sends the new consensus info to all subscribed pandora clients
//  - store consensus info into cache as well as into kv consensusInfoDB
func (s *Service) onNewConsensusInfo(ctx context.Context, consensusInfo *types.MinimalEpochConsensusInfoV2) error {
	nsent := s.consensusInfoFeed.Send(consensusInfo)
	log.WithField("nsent", nsent).Trace("Send consensus info to subscribers")

	if consensusInfo.ReorgInfo != nil {
		// reorg happened. So remove info from database
		finalizedSlot := s.db.LatestLatestFinalizedSlot()
		finalizedEpoch := s.db.LatestLatestFinalizedEpoch()
		log.WithField("curSlot", consensusInfo.ReorgInfo.NewSlot).WithField("revertSlot", finalizedSlot).Warn("Triggered reorg event")

		// Stop subscription of vanguard new pending blocks
		s.stopSubscription()
		s.subscriptionShutdownFeed.Send(&types.PandoraShutDownSignal{Shutdown: true})
		log.WithField("finalizedEpoch", finalizedEpoch).Warn("Stop subscription and reverting orchestrator db to latest finalized slot")

		if err := s.reorgDB(finalizedSlot); err != nil {
			log.WithError(err).Warn("Failed to revert verified info db")
			return err

		}
		// Removing slot infos from vanguard cache
		s.shardingInfoCache.Purge()

		// Re-subscribe vanguard new pending blocks
		if err := s.reSubscribeBlocksEvent(finalizedSlot, finalizedEpoch); err != nil {
			log.WithError(err).Warn("Failed to resubscribe to vanguard blocks event api")
			return err
		}

		s.subscriptionShutdownFeed.Send(&types.PandoraShutDownSignal{Shutdown: false})
	}

	if err := s.db.SaveConsensusInfo(ctx, consensusInfo.ConvertToEpochInfo()); err != nil {
		log.WithError(err).Warn("failed to save consensus info into consensusInfoDB!")
		return err
	}

	if err := s.db.SaveLatestEpoch(ctx, consensusInfo.Epoch); err != nil {
		log.WithError(err).Warn("failed to save latest epoch into consensusInfoDB!")
		return err
	}
	return nil
}

// onNewPendingVanguardBlock
func (s *Service) onNewPendingVanguardBlock(ctx context.Context, blockInfo *eth.StreamPendingBlockInfo) error {
	block := blockInfo.Block
	blockHash, err := block.HashTreeRoot()
	if nil != err {
		log.WithError(err).Warn("failed to retrieve vanguard block hash from HashTreeRoot")
		return err
	}
	wrappedPhase0Blk := wrapper.WrappedPhase0BeaconBlock(block)
	pandoraShards := wrappedPhase0Blk.Body().PandoraShards()
	if len(pandoraShards) < 1 {
		// The first value is the sharding info. If not present throw error
		log.WithField("pandoraShard length", len(pandoraShards)).Error("pandora sharding info not present")
		return errors.New("invalid shard info length in vanguard block body")
	}

	shardInfo := pandoraShards[0]
	cachedShardInfo := &types.VanguardShardInfo{
		Slot:           uint64(block.Slot),
		BlockHash:      blockHash[:],
		ShardInfo:      shardInfo,
		FinalizedSlot:  uint64(blockInfo.FinalizedSlot),
		FinalizedEpoch: uint64(blockInfo.FinalizedEpoch),
	}

	log.WithField("slot", block.Slot).WithField("panBlockNum", shardInfo.BlockNumber).
		WithField("finalizedSlot", blockInfo.FinalizedSlot).WithField("finalizedEpoch", blockInfo.FinalizedEpoch).
		Info("New vanguard shard info has arrived")

	s.vanguardShardingInfoFeed.Send(cachedShardInfo)
	return nil
}

func (s *Service) reorgDB(revertSlot uint64) error {
	// Removing slot infos from verified slot info db
	if err := s.db.RemoveRangeVerifiedInfo(revertSlot+1, s.db.LatestSavedVerifiedSlot()); err != nil {
		log.WithError(err).Error("found error while reverting orchestrator database in reorg phase")
		return err
	}

	if err := s.db.UpdateVerifiedSlotInfo(revertSlot); err != nil {
		log.WithError(err).Error("failed to update latest verified slot info in reorg phase")
		return err
	}
	return nil
}

// reSubscribeBlocksEvent method re-subscribe to vanguard block api.
func (s *Service) reSubscribeBlocksEvent(finalizedSlot, finalizedEpoch uint64) error {
	if s.conn != nil {
		log.Warn("Connection is not nil, could not re-subscribe to vanguard blocks event")
		return nil
	}

	if err := s.dialConn(); err != nil {
		log.WithError(err).Error("Could not create connection with vanguard node during re-subscription")
		return err
	}

	// Re-subscribe vanguard new pending blocks
	go s.subscribeVanNewPendingBlockHash(s.ctx, finalizedSlot)
	go s.subscribeNewConsensusInfoGRPC(s.ctx, finalizedEpoch)
	return nil
}

func (s *Service) stopSubscription() {
	s.processingLock.Lock()
	defer s.processingLock.Unlock()

	s.stopPendingBlkSubCh <- struct{}{}
	s.stopEpochInfoSubCh <- struct{}{}

	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
}
