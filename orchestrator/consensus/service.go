package consensus

import (
	"context"

	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db/iface"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"sync"
)

type Config struct {
}

// Service This part could be moved to other place during refactor, might be registered as a service
type Service struct {
	VanguardHeaderHashDB        iface.VanHeaderAccessDatabase
	PandoraHeaderHashDB         iface.PanHeaderAccessDatabase
	RealmDB                     iface.RealmAccessDatabase
	VanguardHeadersChan         chan *types.HeaderHash
	VanguardConsensusInfoChan   chan *types.MinimalEpochConsensusInfo
	PandoraHeadersChan          chan *types.HeaderHash
	stopChan                    chan struct{}
	canonicalizeChan            chan uint64
	isWorking                   bool
	canonicalizeLock            *sync.Mutex
	invalidationWorkPayloadChan chan *invalidationWorkPayload
	errChan                     chan databaseErrors
	started                     bool
	ctx                         context.Context
}

func New(ctx context.Context) (service *Service) {

}
