package pandorachain

import (
	"context"

	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"github.com/pkg/errors"
)

var (
	errPandoraHeaderProcessing = errors.New("Failed to process the pending pandora header")
)

// subscribePendingHeaders subscribes to pandora client from latest saved slot using given rpc client
func (s *Service) SubscribePendingHeaders(
	ctx context.Context,
	crit *types.PandoraPendingHeaderFilter,
	namespace string,
	client *rpc.Client,
) (*rpc.ClientSubscription, error) {
	ch := make(chan *eth1Types.Header)
	sub, err := client.Subscribe(ctx, namespace, ch, "newPendingBlockHeaders", crit)
	if nil != err {
		return nil, err
	}

	// Start up a dispatcher to feed into the callback
	go func() {
		for {
			select {
			case newPendingHeader := <-ch:
				// dispatch newPendingHeader to handler
				err = s.OnNewPendingHeader(ctx, newPendingHeader)
				if nil != err {
					log.WithError(err).Error("Failed to process the pending pandora header")
					s.conInfoSubErrCh <- errPandoraHeaderProcessing

					IsPandoraConnected.Set(float64(0))
					return
				}
			case err := <-sub.Err():
				log.WithError(err).Debug("Got subscription error, closing existing pending pandora headers subscription")
				return
			case <-ctx.Done():
				log.Info("Received cancelled context, closing existing pending pandora headers subscription")
				return
			}
		}
	}()

	return sub, nil
}
