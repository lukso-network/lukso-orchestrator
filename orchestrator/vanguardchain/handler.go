package vanguardchain

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/common/hexutil"
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
		// Stop subscription of vanguard new pending blocks
		s.stopSubscription()
		// TODO- Stop pandora pending block subscription
		s.subscriptionShutdownFeed.Send(true)

		// reorg happened. So remove info from database
		revertSlot := s.getFinalizedSlot()

		log.WithField("curSlot", consensusInfo.ReorgInfo.NewSlot).WithField("revertSlot", revertSlot).
			Warn("Stop subscription and reverting orchestrator db on live")

		// Removing slot infos from verified slot info db
		if err := s.reorgDB(revertSlot); err != nil {
			log.WithError(err).Warn("Failed to revert verified info db")
			return err

		}

		// Re-subscribe vanguard new pending blocks
		go s.subscribeVanNewPendingBlockHash(revertSlot)
		//TODO- start pandora pending block subscription
		s.subscriptionShutdownFeed.Send(false)
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

	// If current orchestrator's finalize epoch is less than incoming finalized epoch, then update into db and in-memory
	if s.getFinalizedEpoch() < uint64(blockInfo.FinalizedEpoch) {
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
