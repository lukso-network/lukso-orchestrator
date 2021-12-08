package vanguardchain

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"

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

	if err := s.consensusInfoDB.SaveConsensusInfo(ctx, consensusInfo.ConvertToEpochInfo()); err != nil {
		log.WithError(err).Warn("failed to save consensus info into consensusInfoDB!")
		return err
	}

	if err := s.consensusInfoDB.SaveLatestEpoch(ctx, consensusInfo.Epoch); err != nil {
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
		ParentHash:     blockInfo.GetBlock().ParentRoot[:],
	}

	log.WithField("slot", block.Slot).WithField("panBlockNum", shardInfo.BlockNumber).
		WithField("shardingHash", common.BytesToHash(shardInfo.Hash)).WithField("finalizedSlot", blockInfo.FinalizedSlot).
		WithField("finalizedEpoch", blockInfo.FinalizedEpoch).Info("New vanguard shard info has arrived")

	s.vanguardShardingInfoFeed.Send(cachedShardInfo)
	return nil
}
