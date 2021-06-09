package api

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/consensus"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/rpc/api/events"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/iface"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	log "github.com/sirupsen/logrus"
	"sync"
)

type Backend struct {
	ConsensusInfoFeed    iface.ConsensusInfoFeed
	ConsensusInfoDB      db.ROnlyConsensusInfoDB
	VanguardHeaderHashDB db.VanguardHeaderHashDB
	PandoraHeaderHashDB  db.PandoraHeaderHashDB
	RealmDB              db.RealmDB
	// For now I will use singleton pattern instead of service registry
	// TODO: move into service registry
	consensusService *consensus.Service
	sync.Mutex
}

var _ events.Backend = &Backend{}

func (backend *Backend) FetchPanBlockStatus(slot uint64, hash common.Hash) (status events.Status, err error) {
	pandoraHeaderHashDB := backend.PandoraHeaderHashDB

	if nil == pandoraHeaderHashDB {
		err = fmt.Errorf("pandora database is empty")
		status = events.Invalid

		return
	}

	latestSlot := pandoraHeaderHashDB.LatestSavedPandoraSlot()

	if slot > latestSlot {
		status = events.Pending

		return
	}

	headerHash, err := pandoraHeaderHashDB.PandoraHeaderHash(slot)

	if nil != err {
		status = events.Invalid

		return
	}

	pandoraHash := headerHash.HeaderHash

	if pandoraHash.String() != hash.String() && types.Skipped != headerHash.Status {
		err = fmt.Errorf(
			"hashes does not match for slot: %d, provided: %s, proper: %s",
			slot,
			hash.String(),
			pandoraHash.String(),
		)
		status = events.Invalid

		return
	}

	status = events.FromDBStatus(headerHash.Status)

	return
}

func (backend *Backend) FetchVanBlockStatus(slot uint64, hash common.Hash) (status events.Status, err error) {
	vanHashDB := backend.VanguardHeaderHashDB

	if nil == vanHashDB {
		err = fmt.Errorf("vanguard database is empty")
		status = events.Invalid

		return
	}

	latestSlot := vanHashDB.LatestSavedVanguardSlot()

	if slot > latestSlot {
		status = events.Pending

		return
	}

	headerHash, err := vanHashDB.VanguardHeaderHash(slot)

	if nil != err {
		status = events.Invalid

		return
	}

	vanguardHash := headerHash.HeaderHash

	if vanguardHash.String() != hash.String() && types.Skipped != headerHash.Status {
		err = fmt.Errorf(
			"hashes does not match for slot: %d, provided: %s, proper: %s",
			slot,
			hash.String(),
			vanguardHash.String(),
		)
		status = events.Invalid

		return
	}

	status = events.FromDBStatus(headerHash.Status)

	return
}

// Idea is that it should be very little resource intensive as possible, because it could be triggered a lot
// Short circuits will prevent looping when logic says to not do so
func (backend *Backend) InvalidatePendingQueue() (
	vanguardErr error,
	pandoraErr error,
	realmErr error,
) {
	realmDB := backend.RealmDB

	// Invalidation does not need to rely on database, it can be done in multiple ways
	// IMHO we should extend this function by strategy pattern or just stick to one source of truth
	if nil == realmDB {
		log.Errorf("empty realm db")
		return
	}

	consensusService := backend.consensusService

	if nil == consensusService {
		// This service could be fetched from registry nor created during singleton pattern
		backend.consensusService = &consensus.Service{
			VanguardHeaderHashDB: backend.VanguardHeaderHashDB,
			PandoraHeaderHashDB:  backend.PandoraHeaderHashDB,
			RealmDB:              backend.RealmDB,
		}

		consensusService = backend.consensusService
	}

	latestSavedVerifiedRealmSlot := realmDB.LatestVerifiedRealmSlot()
	vanguardErr, pandoraErr, realmErr = consensusService.Canonicalize(
		latestSavedVerifiedRealmSlot,
		// TODO: pass it from cli context
		50000,
	)

	return
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
