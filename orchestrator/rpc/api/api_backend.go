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

// TODO: add err in return
// Idea is that it should be at least resource intensive as possible, because it could be triggered a lot
// Short circuits will prevent looping when logic says to not do so
func (backend *Backend) InvalidatePendingQueue() (vanguardErr error, pandoraErr error, realmErr error) {
	vanguardHashDB := backend.VanguardHeaderHashDB
	pandoraHeaderHashDB := backend.PandoraHeaderHashDB
	realmDB := backend.RealmDB

	// Short circuit, do not invalidate when databases are not present.
	// TODO: consider returning err when not ready
	if nil == vanguardHashDB || nil == pandoraHeaderHashDB || nil == realmDB {
		return
	}

	// If higher slot was found and is valid all the gaps between must me treated as invalid and discarded

	possibleInvalidPair := make([]*events.RealmPair, 0)

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

	// This is quite naive, but should work
	for index, vanguardBlockHash := range vanguardBlockHashes {
		slotToCheck := latestSavedVerifiedRealmSlot + uint64(index)
		pandoraHeaderHash := pandoraHeaderHashes[index]

		if len(pandoraHeaderHashes) < index {
			break
		}

		// Potentially skipped slot
		if nil == pandoraHeaderHash && nil == vanguardBlockHash {
			continue
		}

		// I dont know yet, if it is true.
		// In my opinion INVALID state is 100% accurate only with blockShard verification approach
		// TODO: add additional Sharding info check VanguardBlock -> PandoraHeaderHash when implementation on vanguard side will be ready
		if nil == pandoraHeaderHash {
			vanguardHeaderHash := &types.HeaderHash{
				HeaderHash: vanguardBlockHash.HeaderHash,
				Status:     types.Pending,
			}
			vanguardErr = vanguardHashDB.SaveVanguardHeaderHash(slotToCheck, vanguardHeaderHash)

			possibleInvalidPair = append(possibleInvalidPair, &events.RealmPair{
				Slot:          slotToCheck,
				VanguardHash:  vanguardHeaderHash,
				PandoraHashes: nil,
			})

			continue
		}

		if nil == vanguardBlockHash {
			currentPandoraHeaderHash := &types.HeaderHash{
				HeaderHash: pandoraHeaderHash.HeaderHash,
				Status:     types.Pending,
			}
			currentPandoraHeaderHashes := make([]*types.HeaderHash, 1)
			currentPandoraHeaderHashes[0] = currentPandoraHeaderHash
			pandoraErr = pandoraHeaderHashDB.SavePandoraHeaderHash(slotToCheck, currentPandoraHeaderHash)

			possibleInvalidPair = append(possibleInvalidPair, &events.RealmPair{
				Slot:          slotToCheck,
				VanguardHash:  nil,
				PandoraHashes: currentPandoraHeaderHashes,
			})

			continue
		}

		if types.Verified != vanguardBlockHash.Status {
			vanguardErr = vanguardHashDB.SaveVanguardHeaderHash(slotToCheck, &types.HeaderHash{
				HeaderHash: vanguardBlockHash.HeaderHash,
				Status:     types.Verified,
			})
		}

		if types.Verified != pandoraHeaderHash.Status {
			pandoraErr = pandoraHeaderHashDB.SavePandoraHeaderHash(slotToCheck, &types.HeaderHash{
				HeaderHash: pandoraHeaderHash.HeaderHash,
				Status:     types.Verified,
			})
		}

		if nil != vanguardErr || nil != pandoraErr {
			break
		}

		realmErr = realmDB.SaveLatestVerifiedRealmSlot(slotToCheck)
		pandoraErr = pandoraHeaderHashDB.SaveLatestPandoraSlot()
		vanguardErr = vanguardHashDB.SaveLatestVanguardSlot()

		if nil != realmErr || nil != pandoraErr || nil != vanguardErr {
			break
		}

		vanguardErr = vanguardHashDB.SaveLatestVanguardHeaderHash()
		pandoraErr = pandoraHeaderHashDB.SaveLatestPandoraHeaderHash()

		if nil != vanguardErr || nil != pandoraErr {
			break
		}
	}

	// LOG this out
	if nil != vanguardErr || nil != pandoraErr || nil != realmErr {
		return
	}

	// Resolve state of possible invalid pairs
	latestSavedVerifiedRealmSlot = realmDB.LatestVerifiedRealmSlot()

	for _, pair := range possibleInvalidPair {
		if nil == pair {
			continue
		}

		if pair.Slot > latestSavedVerifiedRealmSlot {
			continue
		}

		vanguardErr = vanguardHashDB.SaveVanguardHeaderHash(pair.Slot, &types.HeaderHash{
			Status: types.Skipped,
		})

		// TODO: when more shard will come we will need to maintain this information
		pandoraErr = pandoraHeaderHashDB.SavePandoraHeaderHash(pair.Slot, &types.HeaderHash{
			Status: types.Skipped,
		})

		if nil != vanguardErr || nil != pandoraErr {
			break
		}
	}

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
