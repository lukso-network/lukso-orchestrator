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
	EmptyHash                = common.HexToHash("0000000000000000000000000000000000000000000000000000000000000000")
	InvalidExtraDataErr      = errors.New("Invalid extra data")
	InvalidSlot              = errors.New("Invalid slot")
	PandoraHeaderNotFoundErr = errors.New("Pandora header not found")
)

// PanHeader
func (s *Store) PandoraHeaderHash(slot uint64) (*types.PanHeaderHash, error) {
	if v, ok := s.panHeaderCache.Get(slot); v != nil && ok {
		return v.(*types.PanHeaderHash), nil
	}
	var headerHash *types.PanHeaderHash
	err := s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(pandoraHeaderHashesBucket)
		key := bytesutil.Uint64ToBytesBigEndian(slot)
		value := bkt.Get(key[:])
		if value == nil {
			return nil
		}
		return decode(value, &headerHash)
	})
	return headerHash, err
}

// PanHeaders
func (s *Store) PandoraHeaderHashes(fromSlot uint64) ([]*types.PanHeaderHash, error) {
	// when requested epoch is greater than stored latest epoch
	if fromSlot > s.latestPanSlot {
		return nil, errors.Wrap(InvalidSlot, fmt.Sprintf("Got invalid fromSlot: %d", fromSlot))
	}
	pandoraHeaderHashes := make([]*types.PanHeaderHash, 0)
	err := s.db.View(func(tx *bolt.Tx) error {
		for slot := fromSlot; slot <= s.latestPanSlot; slot++ {
			// fast finding into cache, if the value does not exist in cache, it starts finding into db
			headerHash, err := s.PandoraHeaderHash(slot)
			if err != nil {
				return errors.Wrap(PandoraHeaderNotFoundErr, fmt.Sprintf("Could not found pandora header for slot: %d", slot))
			}
			pandoraHeaderHashes = append(pandoraHeaderHashes, headerHash)
		}
		return nil
	})
	// the query not successful
	if err != nil {
		return nil, err
	}
	return pandoraHeaderHashes, nil
}

// SavePanHeader
func (s *Store) SavePandoraHeaderHash(slot uint64, headerHash *types.PanHeaderHash) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(pandoraHeaderHashesBucket)
		if status := s.panHeaderCache.Set(latestSavedPanHeaderHashKey, headerHash, 0); !status {
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
		s.latestPanSlot = slot
		s.latestPanHeaderHash = headerHash.HeaderHash
		return nil
	})
}

// SaveLatestPanSlot
func (s *Store) SaveLatestPandoraSlot() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(pandoraHeaderHashesBucket)
		val := bytesutil.Uint64ToBytesBigEndian(s.latestPanSlot)
		if err := bkt.Put(latestSavedPanSlotKey, val); err != nil {
			return err
		}
		return nil
	})
}

// LatestSavedPandoraSlot
func (s *Store) LatestSavedPandoraSlot() uint64 {
	var latestSlot uint64
	// Db is not prepared yet. Retrieve latest saved block number from db
	if !s.isRunning {
		s.db.View(func(tx *bolt.Tx) error {
			bkt := tx.Bucket(pandoraHeaderHashesBucket)
			latestSlotBytes := bkt.Get(latestSavedPanSlotKey[:])
			// not found the latest block number in db. so latest block number will be zero
			if latestSlotBytes == nil {
				latestSlot = 0
				log.Trace("Latest slot number could not find in db. It may happen for brand new DB")
				return nil
			}
			latestSlot = bytesutil.BytesToUint64BigEndian(latestSlotBytes)
			return nil
		})
	}
	return latestSlot
}

// SaveLatestPandoraHeaderHash
func (s *Store) SaveLatestPandoraHeaderHash() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(pandoraHeaderHashesBucket)
		val := s.latestPanHeaderHash.Bytes()
		if err := bkt.Put(latestSavedPanHeaderHashKey, val); err != nil {
			return err
		}
		return nil
	})
}

// LatestSavedPandoraHeaderHash
func (s *Store) LatestSavedPandoraHeaderHash() common.Hash {
	var latestHeaderHash common.Hash
	// Db is not prepared yet. Retrieve latest saved block number from db
	if !s.isRunning {
		s.db.View(func(tx *bolt.Tx) error {
			bkt := tx.Bucket(pandoraHeaderHashesBucket)
			latestHeaderHashBytes := bkt.Get(latestSavedPanHeaderHashKey[:])
			// not found the latest block number in db. so latest block number will be zero
			if latestHeaderHashBytes == nil {
				latestHeaderHash = EmptyHash
				log.Trace("Latest header hash could not find in db. It may happen for brand new DB")
				return nil
			}
			latestHeaderHash = common.BytesToHash(latestHeaderHashBytes)
			return nil
		})
	}
	return latestHeaderHash
}

// GetLatestHeaderHash
func (s *Store) GetLatestHeaderHash() common.Hash {
	return s.latestPanHeaderHash
}
