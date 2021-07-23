package vanguardchain

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	eth "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
)

// OnNewConsensusInfo :
//	- sends the new consensus info to all subscribed pandora clients
//  - store consensus info into cache as well as into kv consensusInfoDB
func (s *Service) OnNewConsensusInfo(ctx context.Context, consensusInfo *types.MinimalEpochConsensusInfo) {
	nsent := s.consensusInfoFeed.Send(consensusInfo)
	log.WithField("nsent", nsent).Trace("Send consensus info to subscribers")

	if err := s.orchestratorDB.SaveConsensusInfo(ctx, consensusInfo); err != nil {
		log.WithError(err).Warn("failed to save consensus info into consensusInfoDB!")
		return
	}
}

// OnNewPendingVanguardBlock
func (s *Service) OnNewPendingVanguardBlock(ctx context.Context, block *eth.BeaconBlock) {
	blockHash, err := block.HashTreeRoot()
	if nil != err {
		log.WithError(err).Warn("failed to retrieve vanguard block hash from HashTreeRoot")
		return
	}
	pandoraShards := block.GetBody().GetPandoraShard()
	if len(pandoraShards) < 1 {
		// The first value is the sharding info. If not present throw error
		log.WithField("pandoraShard length", len(pandoraShards)).Error("pandora sharding info not present")
		return
	}

	shardInfo := pandoraShards[0]
	cachedShardInfo := &types.VanguardShardInfo{
		Slot:      uint64(block.Slot),
		BlockHash: blockHash[:],
		ShardInfo: shardInfo,
	}
	// caching the shard info into sharding cache
	err = s.shardingInfoCache.Put(ctx, uint64(block.Slot), cachedShardInfo)
	if err != nil {
		log.WithField("slot number", block.Slot).
			WithField("error", err).Error("error while inserting sharding info into vanguard cache")
	}

	nSent := s.vanguardShardingInfoFeed.Send(cachedShardInfo)
	log.WithField("nSent", nSent).Trace("Sharding info pushed to consensus service")
}
