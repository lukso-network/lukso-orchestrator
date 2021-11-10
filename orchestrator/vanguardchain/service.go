package vanguardchain

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/cache"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/event"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/client"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

// time to wait before trying to reconnect with the vanguard node.
var (
	reConPeriod               = 2 * time.Second
	syncStatusPollingInterval = 30 * time.Second
)

type DIALGRPCFn func(endpoint string) (client.VanguardClient, error)

// Service
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
	vanGRPCEndpoint   string
	vanGRPCClient     client.VanguardClient
	dialGRPCFn        DIALGRPCFn

	// subscription
	consensusInfoFeed        event.Feed
	scope                    event.SubscriptionScope
	vanguardShardingInfoFeed event.Feed
	// db support
	orchestratorDB db.Database
	// lru cache support
	shardingInfoCache       cache.VanguardShardCache
	resubscribePendingBlkCh chan struct{}
	resubscribeEpochInfoCh  chan struct{}
}

// NewService creates new service with vanguard endpoint, vanguard namespace and consensusInfoDB
func NewService(
	ctx context.Context,
	vanGRPCEndpoint string,
	db db.Database,
	cache cache.VanguardShardCache,
	dialGRPCFn DIALGRPCFn,
) (*Service, error) {

	ctx, cancel := context.WithCancel(ctx)
	_ = cancel // govet fix for lost cancel. Cancel is handled in service.Stop()
	return &Service{
		ctx:                     ctx,
		cancel:                  cancel,
		vanGRPCEndpoint:         vanGRPCEndpoint,
		dialGRPCFn:              dialGRPCFn,
		orchestratorDB:          db,
		shardingInfoCache:       cache,
		resubscribeEpochInfoCh:  make(chan struct{}),
		resubscribePendingBlkCh: make(chan struct{}),
	}, nil
}

// Start a consensus info fetcher service's main event loop.
func (s *Service) Start() {
	// Exit early if eth1 endpoint is not set.
	if s.vanGRPCEndpoint == "" {
		return
	}
	go s.run()
}

func (s *Service) Stop() error {
	if s.cancel != nil {
		defer s.cancel()
	}
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

// run subscribes to all the services for the ETH1.0 chain.
func (s *Service) run() {
	if s.vanGRPCClient == nil {
		log.Error("Vanguard client has not successfully initiated, exiting vanguard chain service!")
		return
	}

	s.waitForConnection()

	latestFinalizedEpoch := s.orchestratorDB.LatestLatestFinalizedEpoch()
	fromEpoch := latestFinalizedEpoch

	// checking consensus info db
	for i := latestFinalizedEpoch; i >= 0; {
		epochInfo, _ := s.orchestratorDB.ConsensusInfo(s.ctx, i)
		if epochInfo == nil {
			// epoch info is missing. so subscribe from here. maybe db operation was wrong
			latestFinalizedEpoch = i
			log.WithField("epoch", fromEpoch).Debug("Found missing epoch info in db, so subscription should " +
				"be started from this missing epoch")
		}
		if i == 0 {
			break
		}
		i--
	}

	latestFinalizedSlot := s.orchestratorDB.LatestLatestFinalizedSlot()

	go s.subscribeNewConsensusInfoGRPC(s.vanGRPCClient, fromEpoch)
	go s.subscribeVanNewPendingBlockHash(s.vanGRPCClient, latestFinalizedSlot)
	go s.syncWithVanguardHead()
}

// waitForConnection waits for a connection with vanguard chain. Until a successful with
// vanguard chain, it retries again and again.
func (s *Service) waitForConnection() {
	if _, err := s.vanGRPCClient.ChainHead(); err == nil {
		log.WithField("vanguardEndpoint", s.vanGRPCEndpoint).Info("Connected vanguard chain")
		s.connectedVanguard = true
		return
	}

	ticker := time.NewTicker(reConPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if _, err := s.vanGRPCClient.ChainHead(); err != nil {
				log.WithField("vanguardEndpoint", s.vanGRPCEndpoint).Warn("Could not connect or subscribe to vanguard chain")
				continue
			}
			s.connectedVanguard = true
			s.runError = nil
			log.WithField("vanguardEndpoint", s.vanGRPCEndpoint).Info("Connected vanguard chain")
			return
		case <-s.ctx.Done():
			log.Info("Received cancelled context, closing existing go routine: waitForConnection")
			return
		}
	}
}

// SubscribeMinConsensusInfoEvent registers a subscription of ChainHeadEvent.
func (s *Service) SubscribeMinConsensusInfoEvent(ch chan<- *types.MinimalEpochConsensusInfoV2) event.Subscription {
	return s.scope.Track(s.consensusInfoFeed.Subscribe(ch))
}

func (s *Service) SubscribeShardInfoEvent(ch chan<- *types.VanguardShardInfo) event.Subscription {
	return s.scope.Track(s.vanguardShardingInfoFeed.Subscribe(ch))
}
