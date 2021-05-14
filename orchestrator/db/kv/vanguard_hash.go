package kv

import (
	"github.com/boltdb/bolt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/shared/bytesutil"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

func (s *Store) LatestSavedVanguardHeaderHash() (hash common.Hash) {
	if !s.isRunning {
		_ = s.db.View(func(tx *bolt.Tx) (dbErr error) {
			bkt := tx.Bucket(pandoraHeaderHashesBucket)
			latestHeaderHashBytes := bkt.Get(latestSavedVanHashKey[:])
			// not found the latest block number in db. so latest block number will be zero
			if latestHeaderHashBytes == nil {
				hash = EmptyHash
				log.Trace("Latest header hash could not find in db. It may happen for brand new DB")
				return nil
			}
			hash = common.BytesToHash(latestHeaderHashBytes)
			return nil
		})
	}

	return
}

func (s *Store) LatestSavedVanguardSlot() (latestSlot uint64) {
	if !s.isRunning {
		_ = s.db.View(func(tx *bolt.Tx) (dbErr error) {
			bkt := tx.Bucket(vanguardHeaderHashesBucket)
			latestSlotBytes := bkt.Get(latestSavedVanSlotKey[:])
			// not found the latest block number in db. so latest block number will be zero
			if latestSlotBytes == nil {
				latestSlot = 0
				log.Trace("Latest slot number could not find in db. It may happen for brand new DB")

				return
			}

			latestSlot = bytesutil.BytesToUint64BigEndian(latestSlotBytes)

			return
		})
	}

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
	err = s.db.Update(func(tx *bolt.Tx) (dbErr error) {
		bkt := tx.Bucket(vanguardHeaderHashesBucket)
		val := bytesutil.Uint64ToBytesBigEndian(s.latestVanSlot)
		dbErr = bkt.Put(latestSavedVanSlotKey, val)

		return
	})

	return
}

func (s *Store) SaveLatestVanguardHeaderHash() (err error) {
	err = s.db.Update(func(tx *bolt.Tx) (dbErr error) {
		bkt := tx.Bucket(vanguardHeaderHashesBucket)
		val := s.latestVanHash.Bytes()
		dbErr = bkt.Put(latestSavedVanHashKey, val)

		return
	})

	return
}
