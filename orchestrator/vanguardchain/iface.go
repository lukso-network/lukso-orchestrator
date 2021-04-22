package vanguardchain

import (
	"github.com/ethereum/go-ethereum/event"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

type ConsensusInfoFeed interface {
	SubscribeMinConsensusInfoEvent(chan<- *types.MinimalEpochConsensusInfo) event.Subscription
}
