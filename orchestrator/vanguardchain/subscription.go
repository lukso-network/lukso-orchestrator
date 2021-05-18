package vanguardchain

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
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
			log.WithField("block", vanBlock).Trace("Got new block")

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

// SubscribeNewConsensusInfo subscribes to vanguard client from latest saved epoch using given rpc client
func (s *Service) subscribeNewConsensusInfo(ctx context.Context, epoch uint64, namespace string,
	client *rpc.Client) (*rpc.ClientSubscription, error) {

	ch := make(chan *types.MinimalEpochConsensusInfo)
	sub, err := client.Subscribe(ctx, namespace, ch, "minimalConsensusInfo", epoch)
	if nil != err {
		return nil, err
	}
	log.WithField("fromEpoch", epoch).Debug("subscribed to vanguard chain for consensus info")
	// Start up a dispatcher to feed into the callback
	go func() {
		for {
			select {
			case consensusInfo := <-ch:
				log.WithField("consensusInfo", consensusInfo).Debug("Got new consensus info from vanguard")

				if nil == consensusInfo {
					log.WithField("consensusInfo", consensusInfo).Info("nil consensus info, discarding")
					continue
				}

				//Sanitize the incoming request
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

				if nil != err {
					log.WithField("consensusInfo", consensusInfo).
						WithField("err", err).
						Error("I am not inserting invalid consensus info")
					continue
				}

				log.WithField("consensusInfo", consensusInfo).Debug("consensus info passed sanitization")

				// dispatch to handle consensus info
				s.OnNewConsensusInfo(ctx, consensusInfo)
			case err := <-sub.Err():
				if err != nil {
					log.WithError(err).Debug("Got subscription error")
					s.conInfoSubErrCh <- err
				}
				return
			}
		}
	}()

	return sub, nil
}
