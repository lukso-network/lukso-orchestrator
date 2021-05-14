package kv

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

func (s *Store) LatestSavedVanguardHeaderHash() (hash common.Hash) {
	return
}

func (s *Store) LatestSavedVanguardSlot() (slot uint64) {
	return
}

func (s *Store) VanguardHeaderHash(slot uint64) (hash *types.HeaderHash, err error) {
	return
}

func (s *Store) VanguardHeaderHashes(fromSlot uint64) (hashes []*types.HeaderHash, err error) {
	return
}

func (s *Store) SaveVanguardHeaderHash(slot uint64, headerHash *types.HeaderHash) (err error) {
	return
}

func (s *Store) SaveLatestVanguardSlot() (err error) {
	return
}

func (s *Store) SaveLatestVanguardHeaderHash() (err error) {
	return
}
