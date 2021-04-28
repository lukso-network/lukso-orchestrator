package pandorachain

import (
	"context"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

// time to wait before trying to reconnect with the vanguard node.
var reConPeriod = 15 * time.Second

// DialRPCFn dials to the given endpoint
type DialRPCFn func(endpoint string) (*rpc.Client, error)

// Service:
// 	- maintains connection with vanguard chain
//	- handles vanguard subscription for consensus info.
//  - sends new consensus info to all pandora subscribers.
//  - maintains db to store the coming consensus info from vanguard.
type Service struct {
	// service maintenance related attributes
	isRunning      bool
	processingLock sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	runError       error

	// vanguard chain related attributes
	connected bool
	endpoint  string
	rpcClient *rpc.Client
	dialRPCFn DialRPCFn
	namespace string

	// subscription
	headerInfoFeed  event.Feed
	scope           event.SubscriptionScope
	conInfoSubErrCh chan error
	conInfoSub      *rpc.ClientSubscription

	// db support
	db db.Database
}

// NewService creates new service with vanguard endpoint, vanguard namespace and db
func NewService(ctx context.Context, vanEndpoint string, namespace string,
	db db.Database, dialRPCFn DialRPCFn) (*Service, error) {

	ctx, cancel := context.WithCancel(ctx)
	_ = cancel // govet fix for lost cancel. Cancel is handled in service.Stop()
	return &Service{
		ctx:       ctx,
		cancel:    cancel,
		endpoint:  vanEndpoint,
		dialRPCFn: dialRPCFn,
		namespace: namespace,
		db:        db,
	}, nil
}

// Start a consensus info fetcher service's main event loop.
func (s *Service) Start() {
	// Exit early if eth1 endpoint is not set.
	if s.endpoint == "" {
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
	if s.rpcClient != nil {
		s.rpcClient.Close()
	}
}

// waitForConnection waits for a connection with vanguard chain. Until a successful with
// vanguard chain, it retries again and again.
func (s *Service) waitForConnection() {
	var err error
	if err = s.connectToVanguardChain(); err == nil {
		log.WithField("vanguardHttp", s.endpoint).Info("Connected vanguard chain")
		s.connected = true
		return
	}
	log.WithError(err).Warn("Could not connect to vanguard endpoint")
	s.runError = err
	ticker := time.NewTicker(reConPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.WithField("endpoint", s.endpoint).Debugf("Dialing vanguard node")
			var errConnect error
			if errConnect = s.connectToVanguardChain(); errConnect != nil {
				log.WithError(errConnect).Warn("Could not connect to vanguard endpoint")
				s.runError = errConnect
				continue
			}
			s.connected = true
			s.runError = nil
			log.WithField("vanguardHttp", s.endpoint).Info("Connected vanguard chain")
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
	// getting latest saved epoch number from db
	latestSavedEpochInDb, err := s.db.LatestSavedEpoch()
	if err != nil {
		log.WithError(err).Warn("Failed to retrieve latest saved epoch information")
		return
	}
	// subscribe to vanguard client for consensus info
	sub, err := s.subscribeNewConsensusInfo(s.ctx, latestSavedEpochInDb, s.namespace, s.rpcClient)
	if err != nil {
		log.WithError(err).Debug("Could not subscribe to vanguard client for consensus info")
	}
	s.conInfoSub = sub

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
				// Before re-subscribe to vanguard, it must have knowledge about last saved epoch consensus info
				latestSavedEpochInDb, err := s.db.LatestSavedEpoch()
				if err != nil {
					log.WithError(err).Warn("Failed to retrieve latest saved epoch information")
					continue
				}
				// re-subscribe to the vanguard chain
				sub, err := s.subscribeNewConsensusInfo(s.ctx, latestSavedEpochInDb, s.namespace, s.rpcClient)
				if err != nil {
					log.WithError(err).Debug("Could not subscribe to vanguard client for consensus info")
					continue
				}
				s.conInfoSub = sub
				continue
			}
		}
	}
}

// connectToVanguardChain dials to vanguard chain and creates rpcClient
func (s *Service) connectToVanguardChain() error {
	if s.rpcClient == nil {
		vanRPCClient, err := s.dialRPCFn(s.endpoint)
		if err != nil {
			return err
		}
		s.rpcClient = vanRPCClient
	}
	return nil
}

// Reconnect to vanguard node in case of any failure.
func (s *Service) retryVanguardNode(err error) {
	s.runError = err
	s.connected = false
	// Back off for a while before resuming dialing the vanguard node.
	time.Sleep(reConPeriod)
	s.waitForConnection()
	// Reset run error in the event of a successful connection.
	s.runError = nil
}

// SubscribeChainHeadEvent registers a subscription of ChainHeadEvent.
func (s *Service) SubscribeMinConsensusInfoEvent(ch chan<- *types.MinimalEpochConsensusInfo) event.Subscription {
	return s.scope.Track(s.headerInfoFeed.Subscribe(ch))
}
