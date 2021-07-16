package api

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/cache"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/consensus"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/rpc/api/events"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/iface"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Backend struct {
	ConsensusInfoFeed    iface.ConsensusInfoFeed
	ConsensusInfoDB      db.ROnlyConsensusInfoDB
	VanguardHeaderHashDB db.VanguardHeaderHashDB
	PandoraHeaderHashDB  db.PandoraHeaderHashDB
	RealmDB              db.RealmDB
	consensusService     *consensus.Service

	VerifiedSlotInfo             db.ROnlyVerifiedSlotInfo
	InvalidSlotInfo              db.ROnlyInvalidSlotInfo
	VanguardPendingShardingCache cache.PandoraHeaderCache
	PandoraPendingHeaderCache    cache.PandoraHeaderCache

	sync.Mutex
}

func (backend *Backend) GetPendingHashes() (response *events.PendingHashesResponse, err error) {
	vanguardHashes, err := backend.VanguardHeaderHashDB.VanguardHeaderHashes(0, 15000)

	if nil != err {
		return
	}

	pandoraHashes, err := backend.PandoraHeaderHashDB.PandoraHeaderHashes(0, 15000)

	if nil != err {
		return
	}

	timestamp := time.Now().Unix()

	response = &events.PendingHashesResponse{
		VanguardHashes:    vanguardHashes,
		PandoraHashes:     pandoraHashes,
		VanguardHashesLen: int64(len(vanguardHashes)),
		PandoraHashesLen:  int64(len(pandoraHashes)),
		UnixTime:          timestamp,
	}

	return
}

var _ events.Backend = &Backend{}

func (backend *Backend) FetchPanBlockStatus(slot uint64, hash common.Hash) (status events.Status, err error) {
	pandoraHeaderHashDB := backend.PandoraHeaderHashDB
	realmDB := backend.RealmDB

	if nil == pandoraHeaderHashDB {
		err = fmt.Errorf("pandora database is empty")
		status = events.Invalid

		return
	}

	if nil == realmDB {
		err = fmt.Errorf("realm database is empty")
		status = events.Invalid

		return
	}

	headerHash, err := pandoraHeaderHashDB.PandoraHeaderHash(slot)

	if nil != err {
		status = events.Invalid

		return
	}

	latestSlot := realmDB.LatestVerifiedRealmSlot()

	if slot > latestSlot {
		status = events.Pending

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
	realmDB := backend.RealmDB

	if nil == vanHashDB {
		err = fmt.Errorf("vanguard database is empty")
		status = events.Invalid

		return
	}

	if nil == realmDB {
		err = fmt.Errorf("realm database is empty")
		status = events.Invalid

		return
	}

	headerHash, err := vanHashDB.VanguardHeaderHash(slot)

	if nil != err {
		status = events.Invalid

		return
	}

	latestSlot := realmDB.LatestVerifiedRealmSlot()

	if slot > latestSlot {
		status = events.Pending

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

// GetSlotStatus
func (b *Backend) GetSlotStatus(ctx context.Context, slot uint64, requestType bool) events.Status {
	if requestType {
		if headerInfo, _ := b.PandoraPendingHeaderCache.Get(ctx, slot); headerInfo != nil {
			log.WithField("slot", slot).WithField("api", "ConfirmPanBlockHashes").Debug("Pending slot")
			return events.Pending
		}
	} else {
		if headerInfo, _ := b.VanguardPendingShardingCache.Get(ctx, slot); headerInfo != nil {
			log.WithField("slot", slot).WithField("api", "ConfirmPanBlockHashes").Debug("Pending slot")
			return events.Pending
		}
	}

	if slotInfo, _ := b.VerifiedSlotInfo.VerifiedSlotInfo(slot); slotInfo != nil {
		log.WithField("slot", slot).WithField("api", "ConfirmPanBlockHashes").Debug("Verified slot")
		return events.Verified
	}

	if slotInfo, _ := b.InvalidSlotInfo.InvalidSlotInfo(slot); slotInfo != nil {
		log.WithField("slot", slot).WithField("api", "ConfirmPanBlockHashes").Debug("Invalid slot")
		return events.Invalid
	}

	return events.Skipped
}
