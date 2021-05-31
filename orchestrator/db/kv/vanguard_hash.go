package kv

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/shared/bytesutil"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"github.com/pkg/errors"
)

var (
	VanguardHeaderNotFoundErr = errors.New("Vanguard header not found")
)

func (s *Store) LatestSavedVanguardHeaderHash() (hash common.Hash) {
	// Quick return
	if s.latestVanHash != hash {
		return s.latestVanHash
	}

	if !s.isRunning {
		_ = s.db.View(func(tx *bolt.Tx) (dbErr error) {
			bkt := tx.Bucket(vanguardHeaderHashesBucket)
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

func (s *Store) VanguardHeaderHash(slot uint64) (headerHash *types.HeaderHash, err error) {
	if v, ok := s.vanHeaderCache.Get(slot); v != nil && ok {
		return v.(*types.HeaderHash), nil
	}

	err = s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(vanguardHeaderHashesBucket)
		key := bytesutil.Uint64ToBytesBigEndian(slot)
		value := bkt.Get(key[:])
		if value == nil {
			return nil
		}
		return decode(value, &headerHash)
	})

	return
}

func (s *Store) VanguardHeaderHashes(fromSlot uint64) (vanguardHeaderHashes []*types.HeaderHash, err error) {
	// when requested epoch is greater than stored latest epoch
	if fromSlot > s.latestVanSlot {
		return nil, errors.Wrap(InvalidSlot, fmt.Sprintf(
			"Got invalid fromSlot: %d, latestVanSlot: %d",
			fromSlot,
			s.latestVanSlot,
		))
	}
	err = s.db.View(func(tx *bolt.Tx) error {
		for slot := fromSlot; slot <= s.latestVanSlot; slot++ {
			// fast finding into cache, if the value does not exist in cache, it starts finding into db
			headerHash, err := s.VanguardHeaderHash(slot)
			// TODO: This cannot be that way, slots can be skipped so they won't be present in database.
			if err != nil {
				return errors.Wrap(VanguardHeaderNotFoundErr, fmt.Sprintf("Could not found pandora header for slot: %d", slot))
			}
			vanguardHeaderHashes = append(vanguardHeaderHashes, headerHash)
		}
		return nil
	})
	// the query not successful
	if err != nil {
		return nil, err
	}
	return vanguardHeaderHashes, nil
}

func (s *Store) SaveVanguardHeaderHash(slot uint64, headerHash *types.HeaderHash) (err error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	err = s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(vanguardHeaderHashesBucket)
		if status := s.vanHeaderCache.Set(latestSavedVanHashKey, headerHash, 0); !status {
			log.WithField("slot", slot).Warn("failed to set pandora header hash into cache")
		}

		key := bytesutil.Uint64ToBytesBigEndian(slot)
		value, err := encode(headerHash)
		if err != nil {
			return err
		}
		if err := bkt.Put(key, value); err != nil {
			return err
		}
		// update latest epoch
		s.latestVanSlot = slot
		s.latestVanHash = headerHash.HeaderHash
		return nil
	})

	return
}

func (s *Store) SaveLatestVanguardSlot() (err error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	err = s.db.Update(func(tx *bolt.Tx) (dbErr error) {
		bkt := tx.Bucket(vanguardHeaderHashesBucket)
		val := bytesutil.Uint64ToBytesBigEndian(s.latestVanSlot)
		dbErr = bkt.Put(latestSavedVanSlotKey, val)

		return
	})

	return
}

func (s *Store) SaveLatestVanguardHeaderHash() (err error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	err = s.db.Update(func(tx *bolt.Tx) (dbErr error) {
		bkt := tx.Bucket(vanguardHeaderHashesBucket)
		val := s.latestVanHash.Bytes()
		dbErr = bkt.Put(latestSavedVanHashKey, val)

		return
	})

	return
}
