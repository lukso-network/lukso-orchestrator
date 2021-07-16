package vanguardchain

import (
	"context"
	"fmt"

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
}

func (s *Service) OnNewPendingVanguardBlock(ctx context.Context, block *eth.BeaconBlock) {
	if nil == block {
		err := fmt.Errorf("block cannot be nil")
		log.WithError(err).Warn("failed to save vanguard block hash")
		return
	}

	latestRealmVerifiedSlot := s.orchestratorDB.LatestVerifiedRealmSlot()

	if uint64(block.Slot) < latestRealmVerifiedSlot {
		log.WithField("extraData", block.Slot).
			WithField("latestRealmVerifiedSlot", latestRealmVerifiedSlot).Error("reorgs not supported")

		return
	}

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
	cachedShardInfo := &types.VanguardShardInfo{Slot: uint64(block.Slot), BlockHash: blockHash[:], ShardInfo: shardInfo}
	err = s.shardingInfoCache.Put(ctx, uint64(block.Slot), cachedShardInfo)
	if err != nil {
		log.WithField("slot number", block.Slot).
			WithField("error", err).Error("error while inserting sharding info into vanguard cache")
	}

	hash := common.BytesToHash(blockHash[:])
	headerHash := &types.HeaderHash{
		HeaderHash:       hash,
		Status:           types.Pending,
		PandoraShardHash: common.BytesToHash(shardInfo.Hash[:]),
		Signature:        shardInfo.Signature,
	}

	nSent := s.vanguardPendingBlockHashFeed.Send(headerHash)
	log.WithField("nsent", nSent).Trace("Pending Block PandoraShardHash feed info to subscribers")

	nSent = s.vanguardShardingInfoFeed.Send(cachedShardInfo)
	log.WithField("nSent", nSent).Trace("Sharding info pushed to vanguardShardingInfoFeed")

	//err = s.orchestratorDB.SaveVanguardHeaderHash(uint64(block.Slot), headerHash)

	//if nil != err {
	//	log.WithError(err).Warn("failed to save vanguard block hash")
	//	return
	//}

	log.WithField("blockHash", headerHash).
		WithField("slot", block.Slot).
		Trace("Successfully inserted vanguard block to db")
}
