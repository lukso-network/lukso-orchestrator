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
	reorgInProgress      bool
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
		vanShardInfoCh := make(chan *types.VanguardShardInfo, 1)
		reorgSignalCh := make(chan *types.Reorg, 1)
		panHeaderInfoCh := make(chan *types.PandoraHeaderInfo, 1)

		vanShardInfoSub := s.vanguardService.SubscribeShardInfoEvent(vanShardInfoCh)
		vanShutdownSub := s.vanguardService.SubscribeShutdownSignalEvent(reorgSignalCh)
		panHeaderInfoSub := s.pandoraService.SubscribeHeaderInfoEvent(panHeaderInfoCh)

		for {
			select {
			case newPanHeaderInfo := <-panHeaderInfoCh:

				if s.reorgInProgress {
					log.WithField("slot", newPanHeaderInfo.Slot).Info("Reorg is progressing, so skipping new pandora header")
					continue
				}

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

				if s.reorgInProgress {
					log.WithField("slot", newVanShardInfo.Slot).Info("Reorg is progressing, so skipping new vanguard shard")
					continue
				}

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
			case reorgInfo := <-reorgSignalCh:
				if reorgInfo == nil {
					log.Error("received shutdown signal but value not set. So we are doing nothing")
					continue
				}
				s.reorgInProgress = true
				// reorg happened. So remove info from database
				finalizedSlot := s.verifiedSlotInfoDB.LatestLatestFinalizedSlot()
				finalizedEpoch := s.verifiedSlotInfoDB.LatestLatestFinalizedEpoch()
				log.WithField("curSlot", reorgInfo.NewSlot).WithField("revertSlot", finalizedSlot).
					WithField("finalizedEpoch", finalizedEpoch).Warn("Triggered reorg event")

				if err := s.reorgDB(finalizedSlot); err != nil {
					log.WithError(err).Warn("Failed to revert verified info db, exiting consensus go routine")
					return
				}
				// Removing slot infos from vanguard cache and pandora cache
				s.vanguardPendingShardingCache.Purge()
				s.pandoraPendingHeaderCache.Purge()
				log.Debug("Starting subscription for vanguard and pandora")

				// disconnect subscription
				log.Debug("Stopping subscription for vanguard and pandora")
				s.vanguardService.StopSubscription()
				s.pandoraService.StopPandoraSubscription()

				//if err := s.vanguardService.ReSubscribeBlocksEvent(); err != nil {
				//	log.WithError(err).Error("Error while subscribing block event, exiting consensus go routine")
				//	return
				//}
				//
				//if err := s.pandoraService.ResumePandoraSubscription(); err != nil {
				//	log.WithError(err).Error("Error while resuming pandora block subscription, exiting consensus go routine")
				//	return
				//}
				// resetting reorgInProgress to false
				s.reorgInProgress = false
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
