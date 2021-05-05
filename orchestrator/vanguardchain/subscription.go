package vanguardchain

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

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
