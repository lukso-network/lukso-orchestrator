package kv

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/shared/bytesutil"
	"github.com/pkg/errors"
)

var (
	InvalidExtraDataErr      = errors.New("Invalid extra data")
	InvalidSlot              = errors.New("Invalid slot")
	PandoraHeaderNotFoundErr = errors.New("Pandora header not found")
)

// PanHeader
func (s *Store) PandoraHeaderHash(slot uint64) (common.Hash, error) {
	if v, ok := s.panHeaderCache.Get(slot); v != nil && ok {
		return v.(common.Hash), nil
	}
	var headerHash common.Hash
	err := s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(pandoraHeaderHashesBucket)
		key := bytesutil.Uint64ToBytesBigEndian(slot)
		value := bkt.Get(key[:])
		if value == nil {
			return nil
		}
		headerHash = common.BytesToHash(value)
		return nil
	})
	return headerHash, err
}

// PanHeaders
func (s *Store) PandoraHeaderHashes(fromSlot uint64) ([]common.Hash, error) {
	// when requested epoch is greater than stored latest epoch
	if fromSlot > s.latestPanSlot {
		return nil, errors.Wrap(InvalidSlot, fmt.Sprintf("Got invalid fromSlot: %d", fromSlot))
	}
	pandoraHeaderHashes := make([]common.Hash, 0)
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
func (s *Store) SavePandoraHeaderHash(slot uint64, headerHash common.Hash) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(pandoraHeaderHashesBucket)
		if status := s.panHeaderCache.Set(slot, headerHash, 0); !status {
			log.WithField("slot", slot).Warn("failed to set pandora header into cache")
		}

		key := bytesutil.Uint64ToBytesBigEndian(slot)
		value := headerHash.Bytes()
		if err := bkt.Put(key, value); err != nil {
			return err
		}
		// update latest epoch
		s.latestPanSlot = slot
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
func (s *Store) LatestSavedPandoraSlot() (uint64, error) {
	// Db is not prepared yet. Retrieve latest saved block number from db
	if !s.isRunning {
		err := s.db.View(func(tx *bolt.Tx) error {
			bkt := tx.Bucket(pandoraHeaderHashesBucket)
			latestSlotBytes := bkt.Get(latestSavedPanSlotKey[:])
			// not found the latest block number in db. so latest block number will be zero
			if latestSlotBytes == nil {
				s.latestPanSlot = uint64(0)
				return errors.New("Latest slot number could not find in db")
			}
			s.latestPanSlot = bytesutil.BytesToUint64BigEndian(latestSlotBytes)
			return nil
		})
		// got decoding error
		if err != nil {
			return 0, err
		}
	}
	return s.latestPanSlot, nil
}
