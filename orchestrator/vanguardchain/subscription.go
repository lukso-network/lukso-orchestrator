package vanguardchain

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/client"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	eth2Types "github.com/prysmaticlabs/eth2-types"
	"time"
)

// subscribeVanNewPendingBlockHash
func (s *Service) subscribeVanNewPendingBlockHash(
	client client.VanguardClient,
) (err error) {

	latestVerifiedSlot := s.orchestratorDB.LatestSavedVerifiedSlot()
	latestVerifiedSlotInfo, err := s.orchestratorDB.VerifiedSlotInfo(latestVerifiedSlot)
	var blockRoot []byte
	if err != nil {
		log.WithField("latestVerifiedSlot", latestVerifiedSlot).
			WithError(err).
			Warn("Failed to retrieve latest verified slot info for pending block subscription")
	}
	if latestVerifiedSlotInfo != nil {
		blockRoot = latestVerifiedSlotInfo.VanguardBlockHash.Bytes()
	}
	if latestVerifiedSlot == 0 {
		latestVerifiedSlot = latestVerifiedSlot + 1
	}
	stream, err := client.StreamNewPendingBlocks(blockRoot, eth2Types.Slot(latestVerifiedSlot))
	if nil != err {
		log.WithError(err).Error("Failed to subscribe to stream of new pending blocks")
		return
	}
	log.WithField("fromSlot", latestVerifiedSlot).
		WithField("blockRoot", hexutil.Encode(blockRoot)).
		Debug("Subscribed to vanguard blocks")

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
