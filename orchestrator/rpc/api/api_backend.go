package api

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/rpc/api/events"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/iface"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

type Backend struct {
	ConsensusInfoFeed    iface.ConsensusInfoFeed
	ConsensusInfoDB      db.ROnlyConsensusInfoDB
	VanguardHeaderHashDB db.VanguardHeaderHashDB
	PandoraHeaderHashDB  db.PandoraHeaderHashDB
}

func (backend *Backend) FetchPanBlockStatus(slot uint64, hash common.Hash) (status events.Status, err error) {
	panic("implement me")
}

func (backend *Backend) FetchVanBlockStatus(slot uint64, hash common.Hash) (status events.Status, err error) {
	panic("implement me")
}

func (backend *Backend) InvalidatePendingQueue() {
	panic("implement me")
}

func (backend *Backend) SubscribeNewEpochEvent(ch chan<- *types.MinimalEpochConsensusInfo) event.Subscription {
	return backend.ConsensusInfoFeed.SubscribeMinConsensusInfoEvent(ch)
}

func (backend *Backend) CurrentEpoch() uint64 {
	return backend.ConsensusInfoDB.GetLatestEpoch()
}

func (backend *Backend) ConsensusInfoByEpochRange(fromEpoch uint64) []*types.MinimalEpochConsensusInfo {
	consensusInfos, err := backend.ConsensusInfoDB.ConsensusInfos(fromEpoch)
	if err != nil {
		return nil
	}
	return consensusInfos
}
