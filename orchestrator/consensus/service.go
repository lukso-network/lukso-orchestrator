package consensus

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/cache"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	iface2 "github.com/lukso-network/lukso-orchestrator/orchestrator/pandorachain/iface"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/iface"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"sync"
)

type Config struct {
	VerifiedSlotInfoDB           db.VerifiedSlotInfoDB
	InvalidSlotInfoDB            db.InvalidSlotInfoDB
	VanguardPendingShardingCache cache.PandoraHeaderCache
	PandoraPendingHeaderCache    cache.PandoraHeaderCache

	VanguardShardFeed iface.VanguardShardInfoFeed
	PandoraHeaderFeed iface2.PandoraHeaderFeed
}

// Service This part could be moved to other place during refactor, might be registered as a service
type Service struct {
	isRunning      bool
	processingLock sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	runError       error

	verifiedSlotInfoDB           db.VerifiedSlotInfoDB
	invalidSlotInfoDB            db.InvalidSlotInfoDB
	vanguardPendingShardingCache cache.PandoraHeaderCache
	pandoraPendingHeaderCache    cache.PandoraHeaderCache

	vanguardShardFeed iface.VanguardShardInfoFeed
	pandoraHeaderFeed iface2.PandoraHeaderFeed
}

func New(ctx context.Context, cfg *Config) (service *Service) {
	ctx, cancel := context.WithCancel(ctx)
	_ = cancel // govet fix for lost cancel. Cancel is handled in service.Stop()

	return &Service{
		ctx:    ctx,
		cancel: cancel,

		verifiedSlotInfoDB:           cfg.VerifiedSlotInfoDB,
		invalidSlotInfoDB:            cfg.InvalidSlotInfoDB,
		vanguardPendingShardingCache: cfg.VanguardPendingShardingCache,
		pandoraPendingHeaderCache:    cfg.PandoraPendingHeaderCache,
		vanguardShardFeed:            cfg.VanguardShardFeed,
		pandoraHeaderFeed:            cfg.PandoraHeaderFeed,
	}
}

func (s *Service) Start() {
	if s.isRunning {
		log.Error("Attempted to start rpc server when it was already started")
		return
	}

	go func() {
		vanShardInfoCh := make(chan *types.VanguardShardInfo)
		panHeaderInfoCh := make(chan *types.PandoraHeaderInfo)

		s.vanguardShardFeed.SubscribeShardInfoEvent(vanShardInfoCh)
		s.pandoraHeaderFeed.SubscribeHeaderInfoEvent(panHeaderInfoCh)

		for {
			select {
			case newPanHeaderInfo := <-panHeaderInfoCh:
				log.WithField("slot", newPanHeaderInfo.Slot).Debug("New pandora header is validating")

			case newVanShardInfo := <-vanShardInfoCh:
				log.WithField("slot", newVanShardInfo.Slot).Debug("New vanguard shard info is validating")

			case <-s.ctx.Done():
				log.Debug("Received cancelled context,closing existing pandora client service")
				return
			}
		}
	}()
}

func (s *Service) Stop() error {
	if s.cancel != nil {
		defer s.cancel()
	}
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
