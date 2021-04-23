package vanguardchain

import (
	"context"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

// SubscribeNewConsensusInfo
func (s *Service) subscribeNewConsensusInfo(
	ctx context.Context,
	epoch uint64,
	namespace string,
	client *rpc.Client,
) (*rpc.ClientSubscription, error) {
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
				log.WithField("epoch", consensusInfo.Epoch).Info("Got new consensus info from vanguard")
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
