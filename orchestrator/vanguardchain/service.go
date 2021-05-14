package vanguardchain

import (
	"context"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/client"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"sync"
	"time"
)

// time to wait before trying to reconnect with the vanguard node.
var reConPeriod = 15 * time.Second

// DialRPCFn dials to the given endpoint
type DialRPCFn func(endpoint string) (*rpc.Client, error)

type DIALGRPCFn func(endpoint string) (client.VanguardClient, error)

// Service:
// 	- maintains connection with vanguard chain
//	- handles vanguard subscription for consensus info.
//  - sends new consensus info to all pandora subscribers.
//  - maintains consensusInfoDB to store the coming consensus info from vanguard.
type Service struct {
	// service maintenance related attributes
	isRunning      bool
	processingLock sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	runError       error

	// vanguard chain related attributes
	connectedVanguard bool
	vanEndpoint       string
	// TODO: pass it down
	vanGRPCEndpoint string
	vanRPCClient    *rpc.Client
	vanGRPCClient   *client.VanguardClient
	dialRPCFn       DialRPCFn
	dialGRPCFn      DIALGRPCFn
	namespace       string

	// subscription
	consensusInfoFeed            event.Feed
	scope                        event.SubscriptionScope
	conInfoSubErrCh              chan error
	conInfoSub                   *rpc.ClientSubscription
	vanguardPendingBlockHashFeed event.Feed

	// db support
	consensusInfoDB      db.ConsensusInfoAccessDB
	vanguardHeaderHashDB db.VanguardHeaderHashDB
}

// NewService creates new service with vanguard endpoint, vanguard namespace and consensusInfoDB
func NewService(
	ctx context.Context,
	vanEndpoint string,
	vanGRPCEndpoint string,
	namespace string,
	consensusInfoAccessDB db.ConsensusInfoAccessDB,
	vanguardHeaderHashDB db.VanguardHeaderHashDB,
	dialRPCFn DialRPCFn,
	dialGRPCFn DIALGRPCFn,
) (*Service, error) {

	ctx, cancel := context.WithCancel(ctx)
	_ = cancel // govet fix for lost cancel. Cancel is handled in service.Stop()
	return &Service{
		ctx:                  ctx,
		cancel:               cancel,
		vanEndpoint:          vanEndpoint,
		vanGRPCEndpoint:      vanGRPCEndpoint,
		dialRPCFn:            dialRPCFn,
		dialGRPCFn:           dialGRPCFn,
		namespace:            namespace,
		conInfoSubErrCh:      make(chan error),
		consensusInfoDB:      consensusInfoAccessDB,
		vanguardHeaderHashDB: vanguardHeaderHashDB,
	}, nil
}

// Start a consensus info fetcher service's main event loop.
func (s *Service) Start() {
	// Exit early if eth1 endpoint is not set.
	if s.vanEndpoint == "" {
		return
	}
	go func() {
		s.isRunning = true
		s.waitForConnection()
		if s.ctx.Err() != nil {
			log.Info("Context closed, exiting pandora goroutine")
			return
		}
		s.run(s.ctx.Done())
	}()
}

func (s *Service) Stop() error {
	if s.cancel != nil {
		defer s.cancel()
	}
	s.scope.Close()
	s.closeClients()
	return nil
}

func (s *Service) Status() error {
	// Service don't start
	if !s.isRunning {
		return nil
	}
	// get error from run function
	if s.runError != nil {
		return s.runError
	}
	return nil
}

// closes down our active eth1 clients.
func (s *Service) closeClients() {
	if s.vanRPCClient != nil {
		s.vanRPCClient.Close()
	}
}

// waitForConnection waits for a connection with vanguard chain. Until a successful with
// vanguard chain, it retries again and again.
func (s *Service) waitForConnection() {
	var err error
	if err = s.connectToVanguardChain(); err == nil {
		log.WithField("vanguardHttp", s.vanEndpoint).Info("Connected vanguard chain")
		s.connectedVanguard = true
		return
	}
	log.WithError(err).Warn("Could not connect to vanguard endpoint")
	s.runError = err
	ticker := time.NewTicker(reConPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.WithField("endpoint", s.vanEndpoint).Debugf("Dialing vanguard node")
			var errConnect error
			if errConnect = s.connectToVanguardChain(); errConnect != nil {
				log.WithError(errConnect).Warn("Could not connect to vanguard endpoint")
				s.runError = errConnect
				continue
			}
			s.connectedVanguard = true
			s.runError = nil
			log.WithField("vanguardHttp", s.vanEndpoint).Info("Connected vanguard chain")
			return
		case <-s.ctx.Done():
			log.Debug("Received cancelled context,closing existing vanguard client service")
			return
		}
	}
}

// run subscribes to all the services for the ETH1.0 chain.
func (s *Service) run(done <-chan struct{}) {
	s.runError = nil

	// the loop waits for any error which comes from consensus info subscription
	// if any subscription error happens, it will try to reconnect and re-subscribe with vanguard chain again.
	for {
		select {
		case <-done:
			s.isRunning = false
			s.runError = nil
			log.Debug("Context closed, exiting goroutine")
			return
		case err := <-s.conInfoSubErrCh:
			if err != nil {
				log.WithError(err).Debug("Could not fetch consensus info from vanguard node")
				// Try to check the connection and retry to establish the connection
				s.retryVanguardNode(err)
				continue
			}
		}
	}
}

// connectToVanguardChain dials to vanguard chain and creates rpcClient
func (s *Service) connectToVanguardChain() error {
	if s.vanRPCClient == nil {
		vanRPCClient, err := s.dialRPCFn(s.vanEndpoint)
		if err != nil {
			return err
		}
		s.vanRPCClient = vanRPCClient
	}

	// connect to vanguard subscription
	if err := s.subscribeToVanguard(); err != nil {
		return err
	}

	err := s.subscribeToVanguardGRPC()

	if nil != err {
		return err
	}

	return nil
}

// Reconnect to vanguard node in case of any failure.
func (s *Service) retryVanguardNode(err error) {
	s.runError = err
	s.connectedVanguard = false
	// Back off for a while before resuming dialing the vanguard node.
	time.Sleep(reConPeriod)
	s.waitForConnection()
	// Reset run error in the event of a successful connection.
	s.runError = nil
}

func (s *Service) subscribeToVanguardGRPC() (err error) {
	// subscribe to vanguard client for new pending blocks
	// TODO: add this client into dependency injection
	// Vanguard endpoint will be invalid I guess as long as we support both grpc and rpc
	vanguardClient, err := s.dialGRPCFn(s.vanGRPCEndpoint)

	if nil != err {
		return
	}

	err, _ = s.subscribeVanNewPendingBlockHash(vanguardClient)

	return
}

// subscribeToVanguard subscribes to vanguard events
// DEPRECATED: we want to move to GRPC and remove rpc connection to vanguard
func (s *Service) subscribeToVanguard() error {
	latestSavedEpochInDb := s.consensusInfoDB.GetLatestEpoch()
	// subscribe to vanguard client for consensus info
	sub, err := s.subscribeNewConsensusInfo(s.ctx, latestSavedEpochInDb, s.namespace, s.vanRPCClient)
	if err != nil {
		log.WithError(err).Warn("Could not subscribe to vanguard client for consensus info")
		return err
	}
	s.conInfoSub = sub
	return nil
}

// SubscribeMinConsensusInfoEvent registers a subscription of ChainHeadEvent.
func (s *Service) SubscribeMinConsensusInfoEvent(ch chan<- *types.MinimalEpochConsensusInfo) event.Subscription {
	return s.scope.Track(s.consensusInfoFeed.Subscribe(ch))
}

func (s *Service) SubscribeVanNewPendingBlockHash(ch chan<- *types.HeaderHash) event.Subscription {
	return s.scope.Track(s.vanguardPendingBlockHashFeed.Subscribe(ch))
}
