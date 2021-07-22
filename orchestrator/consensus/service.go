package consensus

import (
	"context"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/cache"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	iface2 "github.com/lukso-network/lukso-orchestrator/orchestrator/pandorachain/iface"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/iface"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	eth2Types "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"sync"
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

	latestVerifiedSlot uint64
	curSlot            uint64
	curHeader          *eth1Types.Header
	curShardInfo       *eth2Types.PandoraShard
	curSlotInfo        types.CurrentSlotInfo

	verifiedSlotInfoDB           db.VerifiedSlotInfoDB
	invalidSlotInfoDB            db.InvalidSlotInfoDB
	vanguardPendingShardingCache cache.VanguardShardCache
	pandoraPendingHeaderCache    cache.PandoraHeaderCache

	vanguardShardFeed iface.VanguardShardInfoFeed
	pandoraHeaderFeed iface2.PandoraHeaderFeed
}

//
func New(ctx context.Context, cfg *Config) (service *Service) {
	ctx, cancel := context.WithCancel(ctx)
	_ = cancel // govet fix for lost cancel. Cancel is handled in service.Stop()

	latestVerifiedSlot := cfg.VerifiedSlotInfoDB.InMemoryLatestVerifiedSlot()
	log.WithField("latestVerifiedSlot", latestVerifiedSlot).Debug("Initializing consensus service")

	return &Service{
		ctx:                ctx,
		cancel:             cancel,
		latestVerifiedSlot: latestVerifiedSlot,
		curSlot:            latestVerifiedSlot,
		curSlotInfo: types.CurrentSlotInfo{
			Slot: latestVerifiedSlot,
		},
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
				log.WithField("slot", newPanHeaderInfo.Slot).Debug("New pandora header is validating")
				s.curHeader = newPanHeaderInfo.Header
				s.curSlot = newPanHeaderInfo.Slot
				s.assignCurrentSlotInfo(newPanHeaderInfo.Slot)
				s.processingLock.Lock()
				//if s.curSlotInfo.Slot != newPanHeaderInfo.Slot {
				//	s.curSlotInfo.Slot = newPanHeaderInfo.Slot
				//	s.curSlotInfo.Header = newPanHeaderInfo.Header
				//	s.curSlotInfo.Status = types.Pending
				//}
				//s.processingLock.Unlock()

				s.processPandoraHeader(newPanHeaderInfo)
			case newVanShardInfo := <-vanShardInfoCh:
				log.WithField("slot", newVanShardInfo.Slot).Debug("New vanguard shard info is validating")
				s.curShardInfo = newVanShardInfo.ShardInfo
				s.curSlot = newVanShardInfo.Slot
				s.assignCurrentSlotInfo(newPanHeaderInfo.Slot)

				s.processVanguardShardInfo(newVanShardInfo)
			case <-s.ctx.Done():
				vanShardInfoSub.Unsubscribe()
				panHeaderInfoSub.Unsubscribe()
				log.Debug("Received cancelled context,closing existing consensus service")
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

func (s *Service) assignCurrentSlotInfo(curSlot uint64) {
	if s.curSlotInfo.Slot != curSlot {
		curSlotInfo := new(types.CurrentSlotInfo)
		//if
	}
}
