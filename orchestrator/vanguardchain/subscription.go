package vanguardchain

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/client"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"time"
)

func (s *Service) subscribeVanNewPendingBlockHash(
	client client.VanguardClient,
) (err error) {
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
				return
			}

			s.OnNewPendingVanguardBlock(s.ctx, vanBlock)
		}
	}()

	return
}

func (s *Service) subscribeNewConsensusInfoGRPC(
	client client.VanguardClient,
) (err error) {
	stream, err := client.StreamMinimalConsensusInfo(s.orchestratorDB.LatestSavedEpoch())

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
				log.WithError(currentErr).Error("Failed to receive minimalConsensusInfo")
				return
			}

			if nil == vanMinimalConsensusInfo {
				log.Error("Received nil consensus info")
				continue
			}

			log.WithField("fromEpoch", vanMinimalConsensusInfo.Epoch).
				Debug("subscribed to vanguard chain for consensus info")

			consensusInfo := &types.MinimalEpochConsensusInfo{
				Epoch:         uint64(vanMinimalConsensusInfo.Epoch),
				ValidatorList: vanMinimalConsensusInfo.ValidatorList,
				// TODO: consider if this should be handled on pandora side?
				EpochStartTime:   vanMinimalConsensusInfo.EpochTimeStart - 6,
				SlotTimeDuration: time.Duration(vanMinimalConsensusInfo.SlotTimeDuration.Seconds),
			}

			// Only non empty check for now
			if len(consensusInfo.ValidatorList) < 1 {
				log.WithField("consensusInfo", consensusInfo).
					WithField("err", err).
					Error("empty validator list")

				continue
			}

			for index, validator := range consensusInfo.ValidatorList {
				_, err = hexutil.Decode(validator)

				if nil != err {
					log.WithField("consensusInfo", consensusInfo).
						WithField("err", err).
						WithField("index", index).
						WithField("validator", validator).
						Error("could not sanitize the validator")

					break
				}
			}

			log.WithField("consensusInfo", vanMinimalConsensusInfo).
				Debug("consensus info passed sanitization")

			s.OnNewConsensusInfo(s.ctx, consensusInfo)
		}
	}()

	return
}
