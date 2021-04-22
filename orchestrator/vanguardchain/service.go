package vanguardchain

import (
	"context"
	"github.com/ethereum/go-ethereum/event"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/client"
	"sync"
	"time"
)

// time to wait before trying to reconnect with the eth1 node.
var reconnectPeriod = 15 * time.Second

type Config struct {
	VanguardHttpEndpoint string
}

type Service struct {
	// service maintenance related attributes
	isRunning      bool
	processingLock sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	runError       error

	// vanguard chain related attributes
	connectedVanguard    bool
	vanguardHttpEndpoint string
	vanguardClient       client.VanguardClient

	// subscription
	consensusInfoFeed event.Feed
	scope             event.SubscriptionScope
	conInfoSubErrCh   chan error

	// db support
	db db.Database
}

func NewService(ctx context.Context, vanguardHttpEndpoint string, db db.Database) (*Service, error) {
	ctx, cancel := context.WithCancel(ctx)
	_ = cancel // govet fix for lost cancel. Cancel is handled in service.Stop()
	return &Service{
		ctx:                  ctx,
		cancel:               cancel,
		vanguardHttpEndpoint: vanguardHttpEndpoint,
		db:                   db,
	}, nil
}

// Start a consensus info fetcher service's main event loop.
func (s *Service) Start() {
	// Exit early if eth1 endpoint is not set.
	if s.vanguardHttpEndpoint == "" {
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
	vanguardClient, ok := s.vanguardClient.(*client.GRPCClient)
	if ok {
		vanguardClient.Close()
	}
}

// waitForConnection
func (s *Service) waitForConnection() {
	var err error
	if err = s.connectToVanguardChain(); err == nil {
		s.connectedVanguard = true
		return
	}

	s.runError = err
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			log.Debugf("Trying to dial vanguard endpoint: %s", s.vanguardHttpEndpoint)
			var errConnect error
			if errConnect = s.connectToVanguardChain(); errConnect != nil {
				log.Debug("Could not connect to vanguard endpoint")
				s.runError = errConnect
				continue
			}
			s.connectedVanguard = true
			s.runError = nil
			log.WithField("vanguardHttp", s.vanguardHttpEndpoint).Info("Connected vanguard chain")
			return
		case <-s.ctx.Done():
			log.Debug("Received cancelled context,closing existing pandora and vanguard client service")
			return
		}
	}
}

// run subscribes to all the services for the ETH1.0 chain.
func (s *Service) run(done <-chan struct{}) {
	s.runError = nil

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
				s.retryVanguardNode(err)
				latestSavedEpochInDb, err := s.db.LatestSavedEpoch()
				if err != nil {
					log.WithError(err).Warn("Failed to retrieve latest saved epoch information")
					continue
				}
				s.subscribeNewConsensusInfo(s.ctx, latestSavedEpochInDb, "van")
				continue
			}
		}
	}
}

// connectToVanguardChain
func (s *Service) connectToVanguardChain() error {
	if s.vanguardClient == nil {
		vanguardClient, err := client.Dial(
			s.ctx,
			s.vanguardHttpEndpoint,
			1*time.Second,
			5,
			4194304)
		if err != nil {
			return err
		}
		s.vanguardClient = vanguardClient
	}

	// Make a simple call to ensure we are actually connected to a working node.
	_, err := s.vanguardClient.CanonicalHeadSlot()
	if err != nil {
		s.vanguardClient.Close()
		return err
	}
	return nil
}

// Reconnect to eth1 node in case of any failure.
func (s *Service) retryVanguardNode(err error) {
	s.runError = err
	s.connectedVanguard = false
	// Back off for a while before resuming dialing the vanguard node.
	time.Sleep(reconnectPeriod)
	s.waitForConnection()
	// Reset run error in the event of a successful connection.
	s.runError = nil
}
