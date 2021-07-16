package iface

import (
	"context"
	"github.com/ethereum/go-ethereum/event"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

type ConsensusInfoFeed interface {
	SubscribeMinConsensusInfoEvent(chan<- *types.MinimalEpochConsensusInfo) event.Subscription
}

type NewConsensusInfoHandler interface {
	OnNewConsensusInfo(ctx context.Context, consensusInfo *types.MinimalEpochConsensusInfo)
}

type VanguardShardInfoFeed interface {
	SubscribeShardInfoEvent(chan<- *types.VanguardShardInfo) event.Subscription
}

type NewHeaderHandler interface {
	OnNewHeader()
	OnHeaderSubError(err error)
}
