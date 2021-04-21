package vanguardchain

import (
	"github.com/ethereum/go-ethereum/event"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	eth2Types "github.com/prysmaticlabs/eth2-types"
)

type EpochExtractor interface {
	SubscribeMinConsensusInfoEvent(chan<- *types.MinimalEpochConsensusInfo) event.Subscription
	CurrentEpoch() eth2Types.Epoch
	ConsensusInfoByEpochRange(fromEpoch, toEpoch eth2Types.Epoch) map[eth2Types.Epoch]*types.MinimalEpochConsensusInfo
}
