package vanguardchain

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
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

	if err := s.orchestratorDB.SaveLatestEpoch(ctx); err != nil {
		log.WithError(err).Warn("failed to save latest epoch into consensusInfoDB!")
		return
	}
}

// OnNewPendingVanguardBlock
func (s *Service) OnNewPendingVanguardBlock(ctx context.Context, block *eth.BeaconBlock) error {
	blockHash, err := block.HashTreeRoot()
	if nil != err {
		log.WithError(err).Warn("failed to retrieve vanguard block hash from HashTreeRoot")
		return err
	}
	pandoraShards := block.GetBody().GetPandoraShard()
	if len(pandoraShards) < 1 {
		// The first value is the sharding info. If not present throw error
		log.WithField("pandoraShard length", len(pandoraShards)).Error("pandora sharding info not present")
		return errors.New("Invalid shard info length in vanguard block body")
	}

	shardInfo := pandoraShards[0]
	cachedShardInfo := &types.VanguardShardInfo{
		Slot:      uint64(block.Slot),
		BlockHash: blockHash[:],
		ShardInfo: shardInfo,
	}

	if slotInfo, _ := s.orchestratorDB.VerifiedSlotInfo(uint64(block.Slot)); slotInfo != nil {
		blockHashHex := common.BytesToHash(cachedShardInfo.BlockHash[:])
		if slotInfo.VanguardBlockHash == blockHashHex {
			log.WithField("slot", block.Slot).
				WithField("headerHash", blockHash).
				Info("Vanguard shard info is already in verified slot info db")
			return nil
		}
		// TODO- When vanguard pushes new shard info for old slot, then we should take take a rational decision for the header
		// TODO: We also need to have a fork choice mechanism in orchestrator client as well as pandora client
	}

	log.WithField("slot", block.Slot).
		WithField("headerHash", common.BytesToHash(cachedShardInfo.BlockHash[:])).
		Debug("New vanguard shard info has arrived")

	// caching the shard info into sharding cache
	err = s.shardingInfoCache.Put(ctx, uint64(block.Slot), cachedShardInfo)
	if err != nil {
		log.WithField("slot number", block.Slot).
			WithField("err", err).Error("error while inserting sharding info into vanguard cache")
	}

	nSent := s.vanguardShardingInfoFeed.Send(cachedShardInfo)
	log.WithField("nSent", nSent).Trace("Sharding info pushed to consensus service")
	return nil
}
