package api

import (
	"github.com/ethereum/go-ethereum/event"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/iface"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

type APIBackend struct {
	ConsensusInfoFeed iface.ConsensusInfoFeed
	ConsensusInfoDB   db.ReadOnlyDatabase
}

func (backend *APIBackend) SubscribeNewEpochEvent(ch chan<- *types.MinimalEpochConsensusInfo) event.Subscription {
	return backend.ConsensusInfoFeed.SubscribeMinConsensusInfoEvent(ch)
}

func (backend *APIBackend) CurrentEpoch() uint64 {
	curEpoch, err := backend.ConsensusInfoDB.LatestSavedEpoch()
	if err != nil {
		return 0
	}
	return curEpoch
}

func (backend *APIBackend) ConsensusInfoByEpochRange(fromEpoch uint64) []*types.MinimalEpochConsensusInfo {
	consensusInfos, err := backend.ConsensusInfoDB.ConsensusInfos(fromEpoch)
	if err != nil {
		return nil
	}
	return consensusInfos
}
