package pandorachain

import (
	"context"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/cache"
	testDB "github.com/lukso-network/lukso-orchestrator/orchestrator/db/testing"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"testing"
)

// pandoraChainService
type pandoraChainService struct {
	unsubscribed    chan string
	pendingHeaderCh chan *eth1Types.Header
}

// Unsubscribe
func (s *pandoraChainService) Unsubscribe(subid string) {
	if s.unsubscribed != nil {
		s.unsubscribed <- subid
	}
}

// NewPendingBlockHeaders
func (s *pandoraChainService) NewPendingBlockHeaders(
	ctx context.Context, filter types.PandoraPendingHeaderFilter,
) (*rpc.Subscription, error) {

	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return nil, rpc.ErrNotificationsUnsupported
	}
	subscription := notifier.CreateSubscription()

	go func() {
		for {
			select {
			case c := <-s.pendingHeaderCh:
				log.WithField("header", c).Info("sending pending header to subscriber")
				err := notifier.Notify(subscription.ID, c)

				if nil != err {
					log.WithField("context", "error during epoch send").Error(err)
				}
			case <-subscription.Err():
				log.Info("unsubscribing registered subscriber")
				s.unsubscribed <- string(subscription.ID)
				return
			case <-notifier.Closed():
				log.Info("unsubscribing registered subscriber")
				s.unsubscribed <- string(subscription.ID)
				return
			}
		}
	}()
	return subscription, nil
}

// SetupInProcServer
func SetupInProcServer(t *testing.T) (*rpc.Server, *pandoraChainService) {
	server := rpc.NewServer()
	panService := &pandoraChainService{
		unsubscribed:    make(chan string),
		pendingHeaderCh: make(chan *eth1Types.Header),
	}
	if err := server.RegisterName("eth", panService); err != nil {
		panic(err)
	}
	return server, panService
}

// SetupPandoraSvc creates pandora chain service with mocked pandora chain interop
func SetupPandoraSvc(ctx context.Context, t *testing.T, dialRPCFn DialRPCFn) *Service {
	level, err := logrus.ParseLevel("debug")
	assert.NoError(t, err)
	logrus.SetLevel(level)

	svc, err := NewService(
		ctx,
		"ws://127.0.0.1:8546",
		"eth",
		testDB.SetupDB(t),
		cache.NewPanHeaderCache(),
		dialRPCFn)
	if err != nil {
		t.Fatalf("failed to create protocol stack: %v", err)
	}

	return svc
}

// DialInProcClient creates in process client for pandora mocked server
func DialInProcClient(server *rpc.Server) DialRPCFn {
	return func(endpoint string) (*rpc.Client, error) {
		client := rpc.DialInProc(server)
		if client == nil {
			return nil, errors.New("failed to create in-process client")
		}
		return client, nil
	}
}

// DialRPCClient creates in process client for pandora rpc server
func DialRPCClient() DialRPCFn {
	return func(endpoint string) (*rpc.Client, error) {
		client, err := rpc.Dial(endpoint)
		if err != nil {
			return nil, err
		}
		return client, nil
	}
}
