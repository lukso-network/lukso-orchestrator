package pandorachain

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/event"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

// time to wait before trying to reconnect.
var (
	reConPeriod = 2 * time.Second
	EmptyHash   = common.HexToHash("0000000000000000000000000000000000000000000000000000000000000000")
)

// DialRPCFn dials to the given endpoint
type DialRPCFn func(endpoint string) (*rpc.Client, error)

// Service
// 	- maintains connection with pandora chain
//  - maintains db and cache to store the in-coming headers from pandora.
type Service struct {
	// service maintenance related attributes
	isRunning      bool
	processingLock sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	runError       error

	// pandora chain related attributes
	connected bool
	endpoint  string
	rpcClient *rpc.Client
	dialRPCFn DialRPCFn
	namespace string

	// subscription
	conInfoSubErrCh      chan error
	conInfoSub           *rpc.ClientSubscription
	vanguardSubscription event.Subscription

	// db support
	db db.ROnlyVerifiedShardInfoDB

	scope                 event.SubscriptionScope
	pandoraHeaderInfoFeed event.Feed
}

// NewService creates new service with pandora ws or ipc endpoint, pandora service namespace and db
func NewService(
	ctx context.Context,
	endpoint string,
	namespace string,
	db db.ROnlyVerifiedShardInfoDB,
	dialRPCFn DialRPCFn,
) (*Service, error) {

	ctx, cancel := context.WithCancel(ctx)
	_ = cancel // govet fix for lost cancel. Cancel is handled in service.Stop()
	return &Service{
		ctx:             ctx,
		cancel:          cancel,
		endpoint:        endpoint,
		dialRPCFn:       dialRPCFn,
		namespace:       namespace,
		conInfoSubErrCh: make(chan error),
		db:              db,
	}, nil
}

// Start a consensus info fetcher service's main event loop.
func (s *Service) Start() {
	// Exit early if pandora endpoint is not set.
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
	s.closeClients()
	s.scope.Close()

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

// waitForConnection waits for a connection with pandora chain. Until a successful connection and subscription with
// pandora chain, it retries again and again.
func (s *Service) waitForConnection() {
	log.Debug("Waiting for the connection")
	var err error
	if err = s.connectToChain(); err == nil {
		log.WithField("endpoint", s.endpoint).Info("Connected and subscribed to pandora chain")
		s.connected = true
		s.runError = nil
		return
	}
	log.WithError(err).Warn("Could not connect or subscribe to pandora chain")
	s.runError = err
	ticker := time.NewTicker(reConPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.WithField("endpoint", s.endpoint).Debug("Dialing pandora node")
			var errConnect error
			if errConnect = s.connectToChain(); errConnect != nil {
				log.WithError(errConnect).Warn("Could not connect or subscribe to pandora chain")
				s.runError = errConnect
				continue
			}
			s.connected = true
			s.runError = nil
			log.WithField("endpoint", s.endpoint).Info("Connected and subscribed to pandora chain")
			return
		case <-s.ctx.Done():
			log.Info("Received cancelled context, closing existing pandora client connection service")
			return
		}
	}
}

// run subscribes to all the services for the ETH1.0 chain.
func (s *Service) run(done <-chan struct{}) {
	log.Debug("Pandora chain service is starting")
	s.runError = nil

	// the loop waits for any error which comes from consensus info subscription
	// if any subscription error happens, it will try to reconnect and re-subscribe with pandora chain again.
	for {
		select {
		case <-done:
			s.isRunning = false
			s.runError = nil
			log.Info("Context closed, exiting pandora chain service goroutine")
			return
		case err := <-s.conInfoSubErrCh:
			log.WithError(err).Debug("Got subscription error")
			log.Debug("Starting retry to connect and subscribe to pandora chain")
			// Try to check the connection and retry to establish the connection
			s.retryToConnectAndSubscribe(err)
			continue
		}
	}
}

func (s *Service) Resubscribe() {
	if s.conInfoSub != nil {
		s.conInfoSub.Unsubscribe()
		// resubscribing from latest finalised slot
		s.retryToConnectAndSubscribe(nil)
	}
}

// connectToChain dials to pandora chain and creates rpcClient and subscribe
func (s *Service) connectToChain() error {
	if s.rpcClient == nil {
		panRPCClient, err := s.dialRPCFn(s.endpoint)
		if err != nil {
			return err
		}
		s.rpcClient = panRPCClient
	}

	// connect to pandora subscription
	if err := s.subscribe(); err != nil {
		return err
	}
	return nil
}

// retryToConnectAndSubscribe retries to pandora chain in case of any failure.
func (s *Service) retryToConnectAndSubscribe(err error) {
	s.runError = err
	s.connected = false
	// Back off for a while before resuming dialing the pandora node.
	time.Sleep(reConPeriod)
	go s.waitForConnection()
}

// subscribe subscribes to pandora events
func (s *Service) subscribe() error {
	filter := &types.PandoraPendingHeaderFilter{
		FromBlockHash: EmptyHash,
	}

	curStepId := s.db.LatestStepID()
	latestShardInfo, _ := s.db.VerifiedShardInfo(curStepId)
	finalizedSlot := s.db.FinalizedSlot()
	finalizedStepId, _ := s.db.GetStepIdBySlot(finalizedSlot)
	finalizedShardInfo, _ := s.db.VerifiedShardInfo(finalizedStepId)

	if latestShardInfo != nil && finalizedShardInfo != nil {
		if latestShardInfo.SlotInfo.Slot < finalizedShardInfo.SlotInfo.Slot {
			shards := latestShardInfo.Shards
			if len(shards) > 0 && len(shards[0].Blocks) > 0 {
				filter.FromBlockHash = shards[0].Blocks[0].HeaderRoot
			}
		} else {
			shards := finalizedShardInfo.Shards
			if len(shards) > 0 && len(shards[0].Blocks) > 0 {
				filter.FromBlockHash = shards[0].Blocks[0].HeaderRoot
			}
		}
	}

	log.WithField("fromPanHash", filter.FromBlockHash).Debug("Start subscribing to pandora client for pending headers")
	// subscribe to pandora client for pending headers
	sub, err := s.SubscribePendingHeaders(s.ctx, filter, s.namespace, s.rpcClient)
	if err != nil {
		log.WithError(err).Warn("Could not subscribe to pandora client for new pending headers")
		return err
	}

	s.conInfoSub = sub
	return nil
}

func (s *Service) SubscribeHeaderInfoEvent(ch chan<- *types.PandoraHeaderInfo) event.Subscription {
	return s.scope.Track(s.pandoraHeaderInfoFeed.Subscribe(ch))
}
