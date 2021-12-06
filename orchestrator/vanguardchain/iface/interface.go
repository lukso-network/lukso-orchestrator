package iface

import (
	"github.com/ethereum/go-ethereum/event"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

type ConsensusInfoFeed interface {
	SubscribeMinConsensusInfoEvent(chan<- *types.MinimalEpochConsensusInfoV2) event.Subscription
}

type VanguardService interface {
	SubscribeShardInfoEvent(chan<- *types.VanguardShardInfo) event.Subscription
	SubscribeShutdownSignalEvent(chan<- *types.Reorg) event.Subscription
	StopSubscription()
}
