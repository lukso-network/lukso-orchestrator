package consensus

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/cache"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/iface"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"sync"
)

type Config struct {
	VerifiedSlotInfo             db.ROnlyVerifiedSlotInfo
	InvalidSlotInfo              db.ROnlyInvalidSlotInfo
	VanguardPendingShardingCache cache.PandoraHeaderCache
	PandoraPendingHeaderCache    cache.PandoraHeaderCache
}

// Service This part could be moved to other place during refactor, might be registered as a service
type Service struct {
	isRunning      bool
	processingLock sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	runError       error

	verifiedSlotInfo             db.ROnlyVerifiedSlotInfo
	invalidSlotInfo              db.ROnlyInvalidSlotInfo
	vanguardPendingShardingCache cache.PandoraHeaderCache
	pandoraPendingHeaderCache    cache.PandoraHeaderCache

	vanguardShardInfoChan chan *types.VanguardShardInfo
	pandoraHeaderInfoChan chan *types.PandoraHeaderInfo

	vanguardService iface.VanguardShardInfoFeed
}

func New(ctx context.Context, cfg *Config) (service *Service) {
	ctx, cancel := context.WithCancel(ctx)
	_ = cancel // govet fix for lost cancel. Cancel is handled in service.Stop()

	return &Service{
		ctx:    ctx,
		cancel: cancel,

		verifiedSlotInfo:             cfg.VerifiedSlotInfo,
		invalidSlotInfo:              cfg.InvalidSlotInfo,
		vanguardPendingShardingCache: cfg.VanguardPendingShardingCache,
		pandoraPendingHeaderCache:    cfg.PandoraPendingHeaderCache,
		vanguardShardInfoChan:        make(chan *types.VanguardShardInfo),
		pandoraHeaderInfoChan:        make(chan *types.PandoraHeaderInfo),
	}
}

func (s *Service) Start() {
	if s.isRunning {
		log.Error("Attempted to start rpc server when it was already started")
		return
	}

	go func() {

	}()
}

func (s *Service) Stop() error {

}

func (s *Service) Status() error {

}
