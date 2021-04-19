package main

import (
	"context"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/gorilla/websocket"
	eventTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	types "github.com/prysmaticlabs/eth2-types"
	"net/http"
)

// wsRequest attempts to open a WebSocket connection to the given URL.
func wsRequest(url, browserOrigin string) error {
	log.Info("checking WebSocket on %s (origin %q)", url, browserOrigin)

	headers := make(http.Header)
	if browserOrigin != "" {
		headers.Set("Origin", browserOrigin)
	}
	conn, _, err := websocket.DefaultDialer.Dial(url, headers)
	if conn != nil {
		conn.Close()
	}
	return err
}

// subscription
func subscription(ctx context.Context, url string, subErr chan<- error, epoch types.Epoch) error {
	client, err := rpc.Dial(url)
	if nil != err {
		return err
	}
	minConsensusInfoCh := make(chan *eventTypes.MinimalEpochConsensusInfo)
	subscription, err := client.Subscribe(ctx, "orc", minConsensusInfoCh, "minimalConsensusInfo", epoch)
	if nil != err {
		return err
	}

	go func() {
		for {
			select {
			case minConsensusInfo := <-minConsensusInfoCh:
				log.WithField("epoch", minConsensusInfo.Epoch).WithField(
					"validatorList", minConsensusInfo.ValidatorList).WithField(
					"epochStartTime", minConsensusInfo.EpochStartTime).Info("minimal consensus info")

			case subscriptionErr := <-subscription.Err():
				subErr <- subscriptionErr
				subscription.Unsubscribe()
				return
			}
		}
	}()

	return nil
}

func main() {
	ctx := context.Background()
	wsBase := "ws://127.0.0.1:8546"
	subscriptionErrCh := make(chan error)
	startingEpoch := types.Epoch(0)

	log.WithField("startingEpoch", startingEpoch).Info("subscribing for consensus info")
	if err := subscription(ctx, wsBase, subscriptionErrCh, startingEpoch); err != nil {
		log.Error(err)
		return
	}

	for {
		log.Error(<-subscriptionErrCh)
		return
	}
}
