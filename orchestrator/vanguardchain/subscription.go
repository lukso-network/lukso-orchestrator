package vanguardchain

import (
	"fmt"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/client"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"time"
)

// TODO(Atif): Need to subscribe from latest block hash
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
			select {
			case <-s.ctx.Done():
				log.Debug("closing subscribeVanNewPendingBlockHash")
				return
			default:
				vanBlock, currentErr := stream.Recv()
				if nil != currentErr {
					log.WithError(currentErr).Error("Failed to receive new pending vanguard block")
					continue
				}

				if err := s.OnNewPendingVanguardBlock(s.ctx, vanBlock); err != nil {
					log.WithError(err).Error("Failed to process the pending vanguard shardInfo")
				}
			}
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
			select {
			case <-s.ctx.Done():
				log.Debug("closing subscribeNewConsensusInfoGRPC")
				return
			default:
				vanMinimalConsensusInfo, currentErr := stream.Recv()
				if nil != currentErr {
					log.WithError(currentErr).Error("Failed to receive minimalConsensusInfo")
					continue
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
		}
	}()
	return
}
