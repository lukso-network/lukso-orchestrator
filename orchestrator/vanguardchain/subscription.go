package vanguardchain

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/client"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

func (s *Service) subscribeVanNewPendingBlockHash(
	client client.VanguardClient,
) (err error, errChan chan error) {
	errChan = make(chan error)
	stream, err := client.StreamNewPendingBlocks()

	if nil != err {
		return
	}

	go func() {
		for {

			vanBlock, currentErr := stream.Recv()

			if currentErr != nil {
				log.WithError(currentErr).Error("Failed to receive chain header")
				continue
			}

			hashTreeRoot, currentErr := vanBlock.HashTreeRoot()

			if nil != currentErr {
				errChan <- currentErr

				continue
			}

			currentBlockHash := common.BytesToHash(hashTreeRoot[:])
			currentHeaderHash := &types.HeaderHash{
				HeaderHash: currentBlockHash,
				Status:     types.Pending,
			}

			nSent := s.vanguardPendingBlockHashFeed.Send(currentHeaderHash)
			log.WithField("nsent", nSent).Trace("Pending Block Hash feed info to subscribers")
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
