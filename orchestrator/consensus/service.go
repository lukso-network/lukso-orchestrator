package consensus

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/event"

	"github.com/lukso-network/lukso-orchestrator/orchestrator/cache"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	iface2 "github.com/lukso-network/lukso-orchestrator/orchestrator/pandorachain/iface"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/iface"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

var (
	errUnknownParent = errors.New("unknown parent")
)

const (
	TotalExecutionShardCount = 1
	ShardsPerVanBlock        = 1
)

type Config struct {
	VerifiedShardInfoDB db.VerifiedShardInfoDB
	PanHeaderCache      cache.PandoraInterface
	VanShardCache       cache.VanguardInterface
	VanguardShardFeed   iface.VanguardService
	PandoraHeaderFeed   iface2.PandoraService
}

// Service This part could be moved to other place during refactor, might be registered as a service
type Service struct {
	isRunning            bool
	processingLock       sync.Mutex
	ctx                  context.Context
	cancel               context.CancelFunc
	runError             error
	scope                event.SubscriptionScope
	db                   db.VerifiedShardInfoDB
	panHeaderCache       cache.PandoraInterface
	vanShardCache        cache.VanguardInterface
	vanguardService      iface.VanguardService
	pandoraService       iface2.PandoraService
	verifiedSlotInfoFeed event.Feed
	reorgInProgress      uint32
}

//
func New(ctx context.Context, cfg *Config) (service *Service) {
	ctx, cancel := context.WithCancel(ctx)
	_ = cancel // govet fix for lost cancel. Cancel is handled in service.Stop()

	return &Service{
		ctx:             ctx,
		cancel:          cancel,
		db:              cfg.VerifiedShardInfoDB,
		panHeaderCache:  cfg.PanHeaderCache,
		vanShardCache:   cfg.VanShardCache,
		vanguardService: cfg.VanguardShardFeed,
		pandoraService:  cfg.PandoraHeaderFeed,
	}
}

func (s *Service) panCacheClearanceLoop() {
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case currentTime := <-ticker.C:
			removedInfo := s.panHeaderCache.RemoveByTime(currentTime)
			// notify pandora node about this deletion because pandora is waiting for reply
			for _, panHeader := range removedInfo {
				if panHeader != nil {
					s.verifiedSlotInfoFeed.Send(&types.SlotInfoWithStatus{
						PandoraHeaderHash: panHeader.Hash(),
						Status:            types.Invalid,
					})
				}
			}

		case <-s.ctx.Done():
			log.Info("panCacheClearLoop terminating due to context close")
			return
		}
	}
}

func (s *Service) vanCacheClearanceLoop() {
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case currentTime := <-ticker.C:
			s.vanShardCache.RemoveByTime(currentTime)

		case <-s.ctx.Done():
			log.Info("vanCacheClearLoop terminating due to context close")
			return
		}
	}
}

func (s *Service) Start() {
	if s.isRunning {
		log.Error("Attempted to start consensus service when it was already started")
		return
	}
	go s.panCacheClearanceLoop()
	go s.vanCacheClearanceLoop()

	s.isRunning = true
	go func() {
		log.Info("Starting consensus service")
		vanShardInfoCh := make(chan *types.VanguardShardInfo, 1)
		panHeaderInfoCh := make(chan *types.PandoraHeaderInfo, 1)

		vanShardInfoSub := s.vanguardService.SubscribeShardInfoEvent(vanShardInfoCh)
		panHeaderInfoSub := s.pandoraService.SubscribeHeaderInfoEvent(panHeaderInfoCh)

		for {
			select {
			case newPanHeaderInfo := <-panHeaderInfoCh:
				if atomic.LoadUint32(&s.reorgInProgress) == 1 {
					log.WithField("slot", newPanHeaderInfo.Slot).Info("Reorg is progressing, so skipping new pandora header")
					continue
				}

				if err := s.processPandoraHeader(newPanHeaderInfo); err != nil {
					log.WithField("error", err).Error("Could not process pandora shard info, exiting consensus service")
					return
				}

			case newVanShardInfo := <-vanShardInfoCh:
				if atomic.LoadUint32(&s.reorgInProgress) == 1 {
					log.WithField("slot", newVanShardInfo.Slot).Info("Reorg is progressing, so skipping new vanguard shard")
					continue
				}

				if err := s.processVanguardShardInfo(newVanShardInfo); err != nil {
					log.WithField("error", err).Error("Could not process vanguard shard info, exiting consensus service")
					return
				}

			case <-s.ctx.Done():
				vanShardInfoSub.Unsubscribe()
				panHeaderInfoSub.Unsubscribe()
				log.Info("Received cancelled context, existing consensus service")
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
