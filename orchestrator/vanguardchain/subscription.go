package vanguardchain

import (
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/client"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"time"
)

// TODO: remove errChan
func (s *Service) subscribeVanNewPendingBlockHash(
	client client.VanguardClient,
) (err error, errChan chan error) {
	errChan = make(chan error)
	stream, err := client.StreamNewPendingBlocks()

	if nil != err {
		log.WithError(err).Error("Failed to subscribe to stream of new pending blocks")
		return
	}

	log.WithField("context", "awaiting to start vanguard block subscription").
		Trace("subscribeVanNewPendingBlockHash")

	go func() {
		for {
			log.WithField("context", "awaiting to fetch vanBlock from stream").Trace("Got new block")
			vanBlock, currentErr := stream.Recv()
			log.WithField("block", vanBlock).Trace("Got new van block")

			if nil != currentErr {
				log.WithError(currentErr).Error("Failed to receive chain header")
				errChan <- currentErr
				continue
			}

			s.OnNewPendingVanguardBlock(s.ctx, vanBlock)
		}
	}()

	return
}

func (s *Service) subscribeNewConsensusInfoGRPC(
	client client.VanguardClient,
) (err error) {
	stream, err := client.StreamMinimalConsensusInfo()

	if nil != err {
		log.WithError(err).Error("Failed to subscribe to stream of new pending blocks")
		return
	}

	log.WithField("context", "awaiting to start vanguard block subscription").
		Trace("subscribeVanNewPendingBlockHash")

	go func() {
		for {
			log.WithField("context", "awaiting to fetch minimalConsensusInfo from stream").
				Trace("Got new minimal consensus info")
			vanMinimalConsensusInfo, currentErr := stream.Recv()
			log.WithField("minimalConsensusInfo", vanMinimalConsensusInfo).
				Trace("Got new minimal consensus info")

			if nil != currentErr {
				log.WithError(currentErr).Error("Failed to receive chain header")
				return
			}

			if nil == vanMinimalConsensusInfo {
				log.Error("Received nil consensus info")
				continue
			}

			log.WithField("fromEpoch", vanMinimalConsensusInfo.Epoch).
				Debug("subscribed to vanguard chain for consensus info")

			// TODO: PROVIDE SANITIZATION
			log.WithField("consensusInfo", vanMinimalConsensusInfo).
				Debug("consensus info passed sanitization")

			consensusInfo := &types.MinimalEpochConsensusInfo{
				Epoch: uint64(vanMinimalConsensusInfo.Epoch),
				// TODO: this part is missing!
				//ValidatorList:    vanMinimalConsensusInfo,
				EpochStartTime:   vanMinimalConsensusInfo.EpochTimeStart,
				SlotTimeDuration: time.Duration(vanMinimalConsensusInfo.SlotTimeDuration.Seconds),
			}
			s.OnNewConsensusInfo(s.ctx, consensusInfo)
		}
	}()

	return
}
