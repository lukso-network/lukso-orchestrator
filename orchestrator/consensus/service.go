package consensus

import (
	"context"
	"github.com/ethereum/go-ethereum/event"
	"sync"

	"github.com/lukso-network/lukso-orchestrator/orchestrator/cache"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	iface2 "github.com/lukso-network/lukso-orchestrator/orchestrator/pandorachain/iface"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/iface"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

type Config struct {
	VerifiedSlotInfoDB           db.VerifiedSlotInfoDB
	InvalidSlotInfoDB            db.InvalidSlotInfoDB
	VanguardPendingShardingCache cache.VanguardShardCache
	PandoraPendingHeaderCache    cache.PandoraHeaderCache

	VanguardShardFeed iface.VanguardShardInfoFeed
	PandoraHeaderFeed iface2.PandoraHeaderFeed
}

// Service This part could be moved to other place during refactor, might be registered as a service
type Service struct {
	isRunning      bool
	processingLock sync.Mutex
	ctx            context.Context
	cancel         context.CancelFunc
	runError       error

	curSlot                      uint64
	scope                        event.SubscriptionScope
	verifiedSlotInfoDB           db.VerifiedSlotInfoDB
	invalidSlotInfoDB            db.InvalidSlotInfoDB
	vanguardPendingShardingCache cache.VanguardShardCache
	pandoraPendingHeaderCache    cache.PandoraHeaderCache

	vanguardShardFeed    iface.VanguardShardInfoFeed
	pandoraHeaderFeed    iface2.PandoraHeaderFeed
	verifiedSlotInfoFeed event.Feed
}

//
func New(ctx context.Context, cfg *Config) (service *Service) {
	ctx, cancel := context.WithCancel(ctx)
	_ = cancel // govet fix for lost cancel. Cancel is handled in service.Stop()

	latestVerifiedSlot := cfg.VerifiedSlotInfoDB.InMemoryLatestVerifiedSlot()
	log.WithField("latestVerifiedSlot", latestVerifiedSlot).Debug("Initializing consensus service")

	return &Service{
		ctx:                          ctx,
		cancel:                       cancel,
		curSlot:                      latestVerifiedSlot,
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
		log.Error("Attempted to start consensus service when it was already started")
		return
	}
	s.isRunning = true
	go func() {
		log.Info("Starting consensus service")
		vanShardInfoCh := make(chan *types.VanguardShardInfo)
		panHeaderInfoCh := make(chan *types.PandoraHeaderInfo)

		vanShardInfoSub := s.vanguardShardFeed.SubscribeShardInfoEvent(vanShardInfoCh)
		panHeaderInfoSub := s.pandoraHeaderFeed.SubscribeHeaderInfoEvent(panHeaderInfoCh)

		for {
			select {
			case newPanHeaderInfo := <-panHeaderInfoCh:
				s.curSlot = newPanHeaderInfo.Slot
				err := s.processPandoraHeader(newPanHeaderInfo)
				if err != nil {
					log.WithField("error", err).Error("error found while processing pandora header")
					return
				}
			case newVanShardInfo := <-vanShardInfoCh:
				s.curSlot = newVanShardInfo.Slot
				err := s.processVanguardShardInfo(newVanShardInfo)
				if err != nil {
					log.WithField("error", err).Error("error found while processing vanguard sharding info")
					return
				}
			case <-s.ctx.Done():
				vanShardInfoSub.Unsubscribe()
				panHeaderInfoSub.Unsubscribe()
				log.Info("Received cancelled context,closing existing consensus service")
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

func (s *Service) SubscribeVerifiedSlotInfoEvent(ch chan<- *types.SlotInfoWithStatus) event.Subscription {
	return s.scope.Track(s.verifiedSlotInfoFeed.Subscribe(ch))
}
