package consensus

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/event"

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

	VanguardShardFeed iface.VanguardService
	PandoraHeaderFeed iface2.PandoraService
}

// Service This part could be moved to other place during refactor, might be registered as a service
type Service struct {
	isRunning      bool
	processingLock sync.Mutex
	ctx            context.Context
	cancel         context.CancelFunc
	runError       error

	scope                        event.SubscriptionScope
	verifiedSlotInfoDB           db.VerifiedSlotInfoDB
	invalidSlotInfoDB            db.InvalidSlotInfoDB
	vanguardPendingShardingCache cache.VanguardShardCache
	pandoraPendingHeaderCache    cache.PandoraHeaderCache

	vanguardService      iface.VanguardService
	pandoraService       iface2.PandoraService
	verifiedSlotInfoFeed event.Feed
}

//
func New(ctx context.Context, cfg *Config) (service *Service) {
	ctx, cancel := context.WithCancel(ctx)
	_ = cancel // govet fix for lost cancel. Cancel is handled in service.Stop()

	return &Service{
		ctx:                          ctx,
		cancel:                       cancel,
		verifiedSlotInfoDB:           cfg.VerifiedSlotInfoDB,
		invalidSlotInfoDB:            cfg.InvalidSlotInfoDB,
		vanguardPendingShardingCache: cfg.VanguardPendingShardingCache,
		pandoraPendingHeaderCache:    cfg.PandoraPendingHeaderCache,
		vanguardService:              cfg.VanguardShardFeed,
		pandoraService:               cfg.PandoraHeaderFeed,
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
		vanShutdownSignalCh := make(chan *types.ShutDownSignal)
		panHeaderInfoCh := make(chan *types.PandoraHeaderInfo)

		vanShardInfoSub := s.vanguardService.SubscribeShardInfoEvent(vanShardInfoCh)
		vanShutdownSub := s.vanguardService.SubscribeShutdownSignalEvent(vanShutdownSignalCh)
		panHeaderInfoSub := s.pandoraService.SubscribeHeaderInfoEvent(panHeaderInfoCh)

		for {
			select {
			case newPanHeaderInfo := <-panHeaderInfoCh:
				if slotInfo, _ := s.verifiedSlotInfoDB.VerifiedSlotInfo(newPanHeaderInfo.Slot); slotInfo != nil {
					if slotInfo.PandoraHeaderHash == newPanHeaderInfo.Header.Hash() {
						log.WithField("slot", newPanHeaderInfo.Slot).
							WithField("headerHash", newPanHeaderInfo.Header.Hash()).
							Info("Pandora header is already in verified slot info db")

						s.verifiedSlotInfoFeed.Send(&types.SlotInfoWithStatus{
							VanguardBlockHash: slotInfo.VanguardBlockHash,
							PandoraHeaderHash: slotInfo.PandoraHeaderHash,
							Status:            types.Verified,
						})

						continue
					}
				}
				if err := s.processPandoraHeader(newPanHeaderInfo); err != nil {
					log.WithField("error", err).Error("error found while processing pandora header")
					return
				}
			case newVanShardInfo := <-vanShardInfoCh:
				if slotInfo, _ := s.verifiedSlotInfoDB.VerifiedSlotInfo(newVanShardInfo.Slot); slotInfo != nil {
					blockHashHex := common.BytesToHash(newVanShardInfo.BlockHash[:])
					if slotInfo.VanguardBlockHash == blockHashHex {
						log.WithField("slot", newVanShardInfo.Slot).
							WithField("shardInfoHash", hexutil.Encode(newVanShardInfo.ShardInfo.Hash)).
							Info("Vanguard shard info is already in verified slot info db")

						continue
					}
				}
				if err := s.processVanguardShardInfo(newVanShardInfo); err != nil {
					log.WithField("error", err).Error("error found while processing vanguard sharding info")
					return
				}
			case shutdownSignal := <-vanShutdownSignalCh:
				if shutdownSignal == nil {
					log.Error("received shutdown signal but value not set. So we are doing nothing")
					continue
				}
				log.WithField("value", *shutdownSignal).Debug("received shut down signal from vanguard")
				if shutdownSignal.Shutdown == true {
					// disconnect subscription
					log.Debug("Stopping subscription for vanguard and pandora")
					s.vanguardService.StopSubscription()
					s.pandoraService.StopPandoraSubscription()
				} else {
					log.Debug("Starting subscription for vanguard and pandora")
					err := s.vanguardService.ReSubscribeBlocksEvent()
					if err != nil {
						log.WithError(err).Error("Error while subscribing block event")
						continue
					}
					err = s.pandoraService.ResumePandoraSubscription()
					if err != nil {
						log.WithError(err).Error("Error while resuming pandora block subscription")
						continue
					}
				}

			case <-s.ctx.Done():
				vanShardInfoSub.Unsubscribe()
				vanShutdownSub.Unsubscribe()
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
