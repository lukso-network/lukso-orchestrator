package api

import (
	"github.com/ethereum/go-ethereum/event"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	eth2Types "github.com/prysmaticlabs/eth2-types"
)

type APIBackend struct {
	EpochExtractor vanguardchain.EpochExtractor
}

func (backend *APIBackend) SubscribeNewEpochEvent(ch chan<- *types.MinimalEpochConsensusInfo) event.Subscription {
	return backend.EpochExtractor.SubscribeMinConsensusInfoEvent(ch)
}

func (backend *APIBackend) CurrentEpoch() eth2Types.Epoch {
	return backend.EpochExtractor.CurrentEpoch()
}

func (backend *APIBackend) ConsensusInfoByEpochRange(
	fromEpoch,
	toEpoch eth2Types.Epoch,
) map[eth2Types.Epoch]*types.MinimalEpochConsensusInfo {

	return backend.EpochExtractor.ConsensusInfoByEpochRange(fromEpoch, toEpoch)
}
