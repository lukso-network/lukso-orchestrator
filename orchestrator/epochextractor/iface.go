package epochextractor

import (
	"github.com/ethereum/go-ethereum/event"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

type EpochExtractor interface {
	SubscribeMinConsensusInfoEvent(chan<- *types.MinConsensusInfoEvent) event.Subscription
}
