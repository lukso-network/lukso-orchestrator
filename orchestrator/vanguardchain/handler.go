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

	if err := s.consensusInfoDB.SaveConsensusInfo(ctx, consensusInfo.ConvertToEpochInfo()); err != nil {
		log.WithError(err).Warn("failed to save consensus info into consensusInfoDB!")
		return err
	}

	if err := s.consensusInfoDB.SaveLatestEpoch(ctx, consensusInfo.Epoch); err != nil {
		log.WithError(err).Warn("failed to save latest epoch into consensusInfoDB!")
		return err
	}

	if consensusInfo.ReorgInfo != nil {
		nsent = s.subscriptionShutdownFeed.Send(consensusInfo.ReorgInfo)
		log.WithField("nsent", nsent).Trace("Send reorg info to consensus service")
		return nil
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

// ReSubscribeBlocksEvent method re-subscribe to vanguard block api.
func (s *Service) ReSubscribeBlocksEvent() error {
	finalizedSlot := s.verifiedShardInfoDB.FinalizedSlot()
	finalizedEpoch := s.verifiedShardInfoDB.FinalizedEpoch()

	log.WithField("finalizedSlot", finalizedSlot).WithField("finalizedEpoch", finalizedEpoch).Info("Resubscribing Block Event")

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

func (s *Service) StopSubscription() {
	defer log.Info("Stopped vanguard gRPC subscription")
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
}
