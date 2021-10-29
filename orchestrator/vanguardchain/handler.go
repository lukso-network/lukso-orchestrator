package vanguardchain

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/shared/fork"
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

	shardInfoHash := common.BytesToHash(shardInfo.Hash)

	log.WithField("slot", block.Slot).
		WithField("blockNumber", shardInfo.BlockNumber).
		WithField("shardInfoHash", shardInfoHash).
		Info("New vanguard shard info has arrived")

	s.vanguardShardingInfoFeed.Send(cachedShardInfo)

	// This is done for manual trigger of reorg
	for slot, hash := range fork.SupportedForkL15PandoraProd {
		triggerred := false

		for invalidSlot, invalidHash := range fork.UnsupportedForkL15PandoraProd {
			if invalidSlot == uint64(block.Slot) && shardInfoHash.String() == invalidHash.String() {
				triggerred = true
			}
		}

		if slot == uint64(block.Slot) && shardInfoHash.String() == hash.String() {
			triggerred = true
		}

		if !triggerred {
			continue
		}

		epoch := block.Slot.Div(32)
		consensusInfo, currentErr := s.orchestratorDB.ConsensusInfo(s.ctx, uint64(epoch))

		if nil != currentErr {
			log.WithField("epoch", epoch).Error(err)

			continue
		}

		if nil == consensusInfo {
			log.Warn("no consensus info for supported fork")

			continue
		}

		log.Warn("forced trigger of reorg")

		err = s.OnNewConsensusInfo(s.ctx, &types.MinimalEpochConsensusInfoV2{
			Epoch:            consensusInfo.Epoch,
			ValidatorList:    consensusInfo.ValidatorList,
			EpochStartTime:   consensusInfo.EpochStartTime,
			SlotTimeDuration: consensusInfo.SlotTimeDuration,
			ReorgInfo: &types.Reorg{
				VanParentHash: blockHash[:],
				PanParentHash: hash.Bytes(),
			},
		})

		if nil != err {
			return err
		}
	}

	return nil
}
