package vanguardchain

import (
	"fmt"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/client"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"time"
)

// subscribeVanNewPendingBlockHash
func (s *Service) subscribeVanNewPendingBlockHash(
	client client.VanguardClient,
) (err error) {
	stream, err := client.StreamNewPendingBlocks()
	if nil != err {
		log.WithError(err).Error("Failed to subscribe to stream of new pending blocks")
		return
	}
	log.Debug("Subscribed to new pending vanguard block")
	go func() {
		for {
			vanBlock, currentErr := stream.Recv()
			if nil != currentErr {
				log.WithError(currentErr).Error("Failed to receive new pending vanguard block")
				return
			}
			log.WithField("slot", vanBlock.Slot).Debug("Got new van block")
			s.OnNewPendingVanguardBlock(s.ctx, vanBlock)
		}
	}()
	return
}

// subscribeNewConsensusInfoGRPC
func (s *Service) subscribeNewConsensusInfoGRPC(client client.VanguardClient) (err error) {
	stream, err := client.StreamMinimalConsensusInfo(s.orchestratorDB.LatestSavedEpoch())
	if nil != err {
		log.WithError(err).Error("Failed to subscribe to stream of new pending blocks")
		return
	}
	log.Debug("Subscribed to minimal consensus info")
	go func() {
		for {
			vanMinimalConsensusInfo, currentErr := stream.Recv()
			if nil != currentErr {
				log.WithError(currentErr).Error("Failed to receive minimalConsensusInfo")
				return
			}
			if nil == vanMinimalConsensusInfo {
				log.Error("Received nil consensus info")
				continue
			}
			consensusInfo := &types.MinimalEpochConsensusInfo{
				Epoch:            uint64(vanMinimalConsensusInfo.Epoch),
				ValidatorList:    vanMinimalConsensusInfo.ValidatorList,
				EpochStartTime:   vanMinimalConsensusInfo.EpochTimeStart,
				SlotTimeDuration: time.Duration(vanMinimalConsensusInfo.SlotTimeDuration.Seconds),
			}
			// Only non empty check for now
			if len(consensusInfo.ValidatorList) < 1 {
				log.WithField("consensusInfo", consensusInfo).WithField("err", err).Error(
					"empty validator list")
				continue
			}
			log.WithField("epoch", vanMinimalConsensusInfo.Epoch).Debug(
				"Received new consensus info for next epoch")
			log.WithField("consensusInfo", fmt.Sprintf("%+v", consensusInfo)).Trace(
				"Received consensus info")
			s.OnNewConsensusInfo(s.ctx, consensusInfo)
		}
	}()
	return
}
