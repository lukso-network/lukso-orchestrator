package api

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/rpc/api/events"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain/iface"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"sync"
)

type Backend struct {
	ConsensusInfoFeed    iface.ConsensusInfoFeed
	ConsensusInfoDB      db.ROnlyConsensusInfoDB
	VanguardHeaderHashDB db.VanguardHeaderHashDB
	PandoraHeaderHashDB  db.PandoraHeaderHashDB
	RealmDB              db.RealmDB
	sync.Locker
}

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

	emptyHash := common.Hash{}

	if emptyHash.String() == hash.String() {
		err = fmt.Errorf("hash cannot be empty")
		status = events.Invalid

		return
	}

	headerHash, err := pandoraHeaderHashDB.PandoraHeaderHash(slot)

	if nil != err {
		status = events.Invalid

		return
	}

	pandoraHash := headerHash.HeaderHash

	if pandoraHash.String() != hash.String() {
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

	emptyHash := common.Hash{}

	if emptyHash.String() == hash.String() {
		err = fmt.Errorf("hash cannot be empty")
		status = events.Invalid

		return
	}

	headerHash, err := vanHashDB.VanguardHeaderHash(slot)

	if nil != err {
		status = events.Invalid

		return
	}

	vanguardHash := headerHash.HeaderHash

	if vanguardHash.String() != hash.String() {
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

// TODO: add err in return
// Idea is that it should be at least resource intensive as possible, because it could be triggered a lot
// Short circuits will prevent looping when logic says to not do so
func (backend *Backend) InvalidatePendingQueue() {
	vanguardHashDB := backend.VanguardHeaderHashDB
	pandoraHeaderHashDB := backend.PandoraHeaderHashDB
	realmDB := backend.RealmDB

	// Short circuit, do not invalidate when databases are not present.
	// TODO: consider returning err when not ready
	if nil == vanguardHashDB || nil == pandoraHeaderHashDB || nil == realmDB {
		return
	}

	//backend.Lock()
	//defer backend.Unlock()

	latestSavedVerifiedRealmSlot := realmDB.LatestVerifiedRealmSlot()
	pandoraHeaderHashes, err := pandoraHeaderHashDB.PandoraHeaderHashes(latestSavedVerifiedRealmSlot)

	if nil != err {
		return
	}

	vanguardBlockHashes, err := vanguardHashDB.VanguardHeaderHashes(latestSavedVerifiedRealmSlot)

	if nil != err {
		return
	}

	pandoraRange := len(pandoraHeaderHashes)
	vanguardRange := len(vanguardBlockHashes)

	// You wont match anything, so short circuit
	if pandoraRange < 1 || vanguardRange < 1 {
		return
	}

	// TODO: consider having slot also present in header hash structure
	var (
		vanguardErr error
		pandoraErr  error
		realmErr    error
	)

	// This is quite naive, but should work
	slotToCheck := latestSavedVerifiedRealmSlot
	for index, vanguardBlockHash := range vanguardBlockHashes {
		slotToCheck++
		vanguardHash := vanguardBlockHash.HeaderHash

		if len(pandoraHeaderHashes) < index {
			break
		}

		pandoraHeaderHash := pandoraHeaderHashes[index]
		pandoraHash := pandoraHeaderHash.HeaderHash

		if vanguardHash != pandoraHash {
			vanguardErr = vanguardHashDB.SaveVanguardHeaderHash(slotToCheck, &types.HeaderHash{
				HeaderHash: vanguardHash,
				Status:     types.Invalid,
			})

			pandoraErr = pandoraHeaderHashDB.SavePandoraHeaderHash(slotToCheck, &types.HeaderHash{
				HeaderHash: pandoraHeaderHash.HeaderHash,
				Status:     types.Invalid,
			})

			break
		}

		if types.Verified != vanguardBlockHash.Status {
			vanguardErr = vanguardHashDB.SaveVanguardHeaderHash(slotToCheck, &types.HeaderHash{
				HeaderHash: vanguardHash,
				Status:     types.Verified,
			})
		}

		if types.Verified != pandoraHeaderHash.Status {
			pandoraErr = pandoraHeaderHashDB.SavePandoraHeaderHash(slotToCheck, &types.HeaderHash{
				HeaderHash: pandoraHash,
				Status:     types.Verified,
			})
		}

		if nil != vanguardErr || nil != pandoraErr {
			break
		}

		realmErr = realmDB.SaveLatestVerifiedRealmSlot(slotToCheck)

		if nil != realmErr {
			break
		}
	}

	// LOG this out
	if nil != vanguardErr || nil != pandoraErr || nil != realmErr {
		return
	}

	//latestPandoraHashes := pandoraHeaderHashDB.GetLatestHeaderHash()

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
