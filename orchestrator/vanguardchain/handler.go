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
		log.Info("reorg has been triggered")
		err := s.orchestratorDB.RevertConsensusInfo(consensusInfo)
		if err != nil {
			log.WithError(err).Error("found error while reverting orchestrator database")
			return err
		}
	}

	if err := s.orchestratorDB.SaveConsensusInfo(ctx, consensusInfo.ConvertToEpochInfo()); err != nil {
		log.WithError(err).Warn("failed to save consensus info into consensusInfoDB!")
		return err
	}

	if err := s.orchestratorDB.SaveLatestEpoch(ctx); err != nil {
		log.WithError(err).Warn("failed to save latest epoch into consensusInfoDB!")
		return err
	}
	return nil
}

// OnNewPendingVanguardBlock
func (s *Service) OnNewPendingVanguardBlock(ctx context.Context, block *eth.BeaconBlock) error {
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
		Slot:      uint64(block.Slot),
		BlockHash: blockHash[:],
		ShardInfo: shardInfo,
	}

	log.WithField("slot", block.Slot).
		WithField("blockNumber", shardInfo.BlockNumber).
		WithField("shardInfoHash", hexutil.Encode(shardInfo.Hash)).
		Info("New vanguard shard info has arrived")

	s.vanguardShardingInfoFeed.Send(cachedShardInfo)
	return nil
}
