package api

import (
	"github.com/ethereum/go-ethereum/event"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/epochextractor"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

type APIBackend struct {
	EpochExtractor epochextractor.EpochExtractor
}

func (backend *APIBackend) SubscribeNewEpochEvent(ch chan<- *types.MinConsensusInfoEvent) event.Subscription {
	return backend.EpochExtractor.SubscribeMinConsensusInfoEvent(ch)
}
