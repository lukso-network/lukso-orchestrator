package kv

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/lukso-network/lukso-orchestrator/shared/bytesutil"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"github.com/pkg/errors"
)

var (
	InvalidExtraDataErr      = errors.New("Invalid extra data")
	InvalidSlot              = errors.New("Invalid slot")
	PandoraHeaderNotFoundErr = errors.New("Pandora header not found")
)

// PanHeader
func (s *Store) PanHeader(slot uint64) (*types.PanBlockHeader, error) {
	if v, ok := s.panHeaderCache.Get(slot); v != nil && ok {
		return v.(*types.PanBlockHeader), nil
	}
	var panHeader *types.PanBlockHeader
	err := s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(pandoraHeadersBucket)
		key := bytesutil.Uint64ToBytesBigEndian(slot)
		enc := bkt.Get(key[:])
		if enc == nil {
			return nil
		}
		return decode(enc, &panHeader)
	})
	return panHeader, err
}

// PanHeaders
func (s *Store) PanHeaders(fromSlot uint64) ([]*types.PanBlockHeader, error) {
	// when requested epoch is greater than stored latest epoch
	if fromSlot > s.latestPanSlot {
		return nil, errors.Wrap(InvalidSlot, fmt.Sprintf("Got invalid fromSlot: %d", fromSlot))
	}

	pandoraHeaders := make([]*types.PanBlockHeader, 0)
	err := s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(consensusInfosBucket)
		for slot := fromSlot; slot <= s.latestPanSlot; slot++ {
			// fast finding into cache, if the value does not exist in cache, it starts finding into db
			if v, _ := s.panHeaderCache.Get(slot); v != nil {
				pandoraHeaders = append(pandoraHeaders, v.(*types.PanBlockHeader))
				continue
			}
			// preparing key bytes for searching into db
			key := bytesutil.Uint64ToBytesBigEndian(slot)
			enc := bkt.Get(key[:])
			if enc == nil {
				return errors.Wrap(PandoraHeaderNotFoundErr, fmt.Sprintf("Could not found pandora header for slot: %d", slot))
			}
			var panHeader *types.PanBlockHeader
			decode(enc, &panHeader)
			pandoraHeaders = append(pandoraHeaders, panHeader)
		}
		return nil
	})
	// the query not successful
	if err != nil {
		return nil, err
	}

	return pandoraHeaders, nil
}

func (s *Store) LatestSavedPanBlockNum() (uint64, error) {
	// Db is not prepared yet. Retrieve latest saved block number from db
	if !s.isRunning {
		err := s.db.View(func(tx *bolt.Tx) error {
			bkt := tx.Bucket(pandoraHeadersBucket)
			latestBlkNumBytes := bkt.Get(latestSavedPanBlockNumKey[:])
			// not found the latest block number in db. so latest block number will be zero
			if latestBlkNumBytes == nil {
				s.latestPanBlockNum = uint64(0)
				return nil
			}
			s.latestPanBlockNum = bytesutil.BytesToUint64BigEndian(latestBlkNumBytes)
			return nil
		})
		// got decoding error
		if err != nil {
			return 0, err
		}
	}
	return s.latestPanBlockNum, nil
}

// SavePanHeader
func (s *Store) SavePanHeader(header *types.PanBlockHeader) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(pandoraHeadersBucket)
		slot, err := SlotFromExtraData(header)
		if err != nil {
			return err
		}

		keyBytes := bytesutil.Uint64ToBytesBigEndian(slot)
		enc, err := encode(header)
		if err != nil {
			return err
		}
		var status bool
		if status = s.panHeaderCache.Set(slot, header, 7); !status {
			log.WithField("slot", slot).Warn("failed to set pandora header into cache")
		}
		if err := bkt.Put(keyBytes, enc); err != nil {
			return err
		}
		// update latest epoch
		s.latestPanSlot = slot
		s.latestPanBlockNum = header.Header.Number.Uint64()
		return nil
	})
}

// SaveLatestPanSlot
func (s *Store) SaveLatestPanSlot() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(pandoraHeadersBucket)
		val := bytesutil.Uint64ToBytesBigEndian(s.latestPanSlot)
		if err := bkt.Put(latestSavedPanSlot, val); err != nil {
			return err
		}
		return nil
	})
}

// SaveLatestPanSlot
func (s *Store) SaveLatestPanBlockNum() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(pandoraHeadersBucket)
		val := bytesutil.Uint64ToBytesBigEndian(s.latestPanBlockNum)
		if err := bkt.Put(latestSavedPanBlockNumKey, val); err != nil {
			return err
		}
		return nil
	})
}
