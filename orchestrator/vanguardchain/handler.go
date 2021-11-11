package vanguardchain

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	eth "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/proto/eth/v1alpha1/wrapper"
)

// OnNewConsensusInfo :
//	- sends the new consensus info to all subscribed pandora clients
//  - store consensus info into cache as well as into kv consensusInfoDB
func (s *Service) OnNewConsensusInfo(ctx context.Context, consensusInfo *types.MinimalEpochConsensusInfoV2) error {
	nsent := s.consensusInfoFeed.Send(consensusInfo)
	log.WithField("nsent", nsent).Trace("Send consensus info to subscribers")

	if consensusInfo.ReorgInfo != nil {
		// reorg happened. So remove info from database
		revertSlot := s.getFinalizedSlot()
		if revertSlot > 0 {
			revertSlot = revertSlot + 1
		}

		log.WithField("curSlot", consensusInfo.ReorgInfo.NewSlot).WithField("revertSlot", revertSlot).
			Warn("Stop subscription and reverting orchestrator db on live")

		// Stop subscription of vanguard new pending blocks
		s.stopSubscription()

		// Removing slot infos from verified slot info db
		err := s.orchestratorDB.RemoveRangeVerifiedInfo(revertSlot, 0)
		if err != nil {
			log.WithError(err).Error("found error while reverting orchestrator database")
			return err
		}

		// Re-subscribe vanguard new pending blocks
		go s.subscribeVanNewPendingBlockHash(revertSlot)
	}

	if err := s.orchestratorDB.SaveConsensusInfo(ctx, consensusInfo.ConvertToEpochInfo()); err != nil {
		log.WithError(err).Warn("failed to save consensus info into consensusInfoDB!")
		return err
	}

	if err := s.orchestratorDB.SaveLatestEpoch(ctx, consensusInfo.Epoch); err != nil {
		log.WithError(err).Warn("failed to save latest epoch into consensusInfoDB!")
		return err
	}
	return nil
}

// OnNewPendingVanguardBlock
func (s *Service) OnNewPendingVanguardBlock(ctx context.Context, blockInfo *eth.StreamPendingBlockInfo) error {
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

	// If current orchestrator's finalize epoch is less than incoming finalized epoch, then update into db and in-memory
	if s.finalizedEpoch < uint64(blockInfo.FinalizedEpoch) {
		newFS := uint64(blockInfo.FinalizedSlot)
		newFE := uint64(blockInfo.FinalizedEpoch)

		if err := s.updateFinalizedInfoInDB(newFS, newFE); err != nil {
			log.WithError(err).Warn("Failed to store new finalized info")
		}
		s.updateInMemoryFinalizedInfo(newFS, newFE)
	}

	shardInfo := pandoraShards[0]
	cachedShardInfo := &types.VanguardShardInfo{
		Slot:      uint64(block.Slot),
		BlockHash: blockHash[:],
		ShardInfo: shardInfo,
	}

	log.WithField("slot", block.Slot).
		WithField("blockNumber", shardInfo.BlockNumber).
		WithField("shardInfoHash", hexutil.Encode(shardInfo.Hash)).
		WithField("latestFinalizedSlot", blockInfo.FinalizedSlot).
		WithField("latestFinalizedEpoch", blockInfo.FinalizedEpoch).
		Info("New vanguard shard info has arrived")

	s.vanguardShardingInfoFeed.Send(cachedShardInfo)
	return nil
}
