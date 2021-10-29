package kv

import (
	"context"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/shared/bytesutil"
	"github.com/lukso-network/lukso-orchestrator/shared/fork"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"github.com/pkg/errors"
)

var (
	EmptyHash      = common.HexToHash("0000000000000000000000000000000000000000000000000000000000000000")
	errInvalidSlot = errors.New("invalid slot and not found any verified slot info for the given slot")
)

// VerifiedSlotInfo
func (s *Store) VerifiedSlotInfo(slot uint64) (*types.SlotInfo, error) {
	if v, ok := s.verifiedSlotInfoCache.Get(slot); v != nil && ok {
		slotInfo := v.(*types.SlotInfo)
		err := fork.GuardAllUnsupportedPandoraForks(slotInfo.PandoraHeaderHash, slot)

		if nil != err {
			log.Warn("Deleting a fork slot")
			_ = s.DeleteVerifiedSlotInfo(slot)

			return nil, err
		}

		return slotInfo, nil
	}
	var slotInfo *types.SlotInfo
	err := s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(verifiedSlotInfosBucket)
		key := bytesutil.Uint64ToBytesBigEndian(slot)
		value := bkt.Get(key[:])
		if value == nil {
			return nil
		}
		err := decode(value, &slotInfo)

		if nil != err {
			slotInfo = nil
			return err
		}

		err = fork.GuardAllUnsupportedPandoraForks(slotInfo.PandoraHeaderHash, slot)

		if nil != err {
			log.WithField("store", "unsupported fork").
				WithField("hash", slotInfo.PandoraHeaderHash).
				WithField("slot", slot).
				Warn("Fork detected")
			log.Warn("Deleting a fork slot")

			_ = s.DeleteVerifiedSlotInfo(slot)
			slotInfo = nil
		}

		return err
	})
	return slotInfo, err
}

func (s *Store) DeleteVerifiedSlotInfo(slot uint64) (err error) {
	s.verifiedSlotInfoCache.Del(slot)

	err = s.db.Update(func(tx *bolt.Tx) (dbErr error) {
		bucket := tx.Bucket(verifiedSlotInfosBucket)
		slotBytes := bytesutil.Uint64ToBytesBigEndian(slot)
		dbErr = bucket.Delete(slotBytes)

		if nil != dbErr {
			err = dbErr
		}

		return
	})

	return
}

// ConsensusInfos
func (s *Store) VerifiedSlotInfos(fromSlot uint64) (map[uint64]*types.SlotInfo, error) {
	latestVerifiedSlot := s.LatestSavedVerifiedSlot()
	// when requested epoch is greater than stored latest epoch
	if fromSlot > latestVerifiedSlot {
		return nil, errors.Wrap(errInvalidSlot, fmt.Sprintf("fromSlot: %d", fromSlot))
	}

	slotInfos := make(map[uint64]*types.SlotInfo)
	err := s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(verifiedSlotInfosBucket)
		for slot := fromSlot; slot <= latestVerifiedSlot; slot++ {
			// fast finding into cache, if the value does not exist in cache, it starts finding into db
			if v, _ := s.verifiedSlotInfoCache.Get(slot); v != nil {
				slotInfo := v.(*types.SlotInfo)
				err := fork.GuardAllUnsupportedPandoraForks(slotInfo.PandoraHeaderHash, slot)

				if nil != err {
					log.WithField("store", "unsupported fork").
						WithField("hash", slotInfo.PandoraHeaderHash).
						WithField("slot", slot).
						Warn("Fork detected")
					_ = s.DeleteVerifiedSlotInfo(slot)

					continue
				}

				slotInfos[slot] = v.(*types.SlotInfo)
				continue
			}
			// preparing key bytes for searching into db
			key := bytesutil.Uint64ToBytesBigEndian(slot)
			enc := bkt.Get(key[:])
			if enc == nil {
				// no data found for the associated slot. So just find for other slot
				continue
			}

			var slotInfo *types.SlotInfo
			decode(enc, &slotInfo)

			err := fork.GuardAllUnsupportedPandoraForks(slotInfo.PandoraHeaderHash, slot)

			if nil != err {
				log.WithField("store", "unsupported fork").
					WithField("hash", slotInfo.PandoraHeaderHash).
					WithField("slot", slot).
					Warn("Fork detected")
				_ = s.DeleteVerifiedSlotInfo(slot)

				continue
			}

			slotInfos[slot] = slotInfo
		}
		return nil
	})
	// the query not successful
	if err != nil {
		return nil, err
	}

	return slotInfos, nil
}

// SaveVerifiedSlotInfo
func (s *Store) SaveVerifiedSlotInfo(slot uint64, slotInfo *types.SlotInfo) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	// storing consensus info into cache and db
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(verifiedSlotInfosBucket)
		slotBytes := bytesutil.Uint64ToBytesBigEndian(slot)
		enc, err := encode(slotInfo)
		if err != nil {
			return err
		}
		if status := s.verifiedSlotInfoCache.Set(slot, slotInfo, 0); !status {
			log.WithField("slot", slot).Warn("could not store verified slot info into cache")
		}
		if err := bkt.Put(slotBytes, enc); err != nil {
			return err
		}
		// store latest verified slot and latest header hash in in-memory
		s.latestVerifiedSlot = slot
		s.latestHeaderHash = slotInfo.PandoraHeaderHash

		return nil
	})
}

// SaveLatestEpoch
func (s *Store) SaveLatestVerifiedSlot(ctx context.Context) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	// storing latest epoch number into db
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(verifiedSlotInfosBucket)
		slotBytes := bytesutil.Uint64ToBytesBigEndian(s.latestVerifiedSlot)
		if err := bkt.Put(latestSavedVerifiedSlotKey, slotBytes); err != nil {
			return err
		}
		return nil
	})
}

// LatestSavedEpoch
func (s *Store) LatestSavedVerifiedSlot() uint64 {
	var latestSavedVerifiedSlot uint64
	// Db is not prepared yet. Retrieve latest saved epoch number from db
	if !s.isRunning {
		s.db.View(func(tx *bolt.Tx) error {
			bkt := tx.Bucket(verifiedSlotInfosBucket)
			slotBytes := bkt.Get(latestSavedVerifiedSlotKey[:])
			// not found the latest epoch in db. so latest epoch will be zero
			if slotBytes == nil {
				latestSavedVerifiedSlot = uint64(0)
				log.Trace("Latest verified slot number could not find in db. It may happen for brand new DB")
				return nil
			}
			latestSavedVerifiedSlot = bytesutil.BytesToUint64BigEndian(slotBytes)
			return nil
		})
	}
	// db is already started so latest epoch must be initialized in store
	return latestSavedVerifiedSlot
}

func (s *Store) InMemoryLatestVerifiedSlot() uint64 {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	return s.latestVerifiedSlot
}

// SaveLatestEpoch
func (s *Store) SaveLatestVerifiedHeaderHash() error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	// storing latest epoch number into db
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(verifiedSlotInfosBucket)
		headerHashBytes := s.latestHeaderHash.Bytes()
		if err := bkt.Put(latestHeaderHashKey, headerHashBytes); err != nil {
			return err
		}
		return nil
	})
}

// LatestSavedEpoch
func (s *Store) LatestVerifiedHeaderHash() common.Hash {
	var latestHeaderHash common.Hash
	// Db is not prepared yet. Retrieve latest saved epoch number from db
	if !s.isRunning {
		s.db.View(func(tx *bolt.Tx) error {
			bkt := tx.Bucket(verifiedSlotInfosBucket)
			latestHeaderHashBytes := bkt.Get(latestHeaderHashKey[:])
			// not found the latest epoch in db. so latest epoch will be zero
			if latestHeaderHashBytes == nil {
				latestHeaderHash = EmptyHash
				log.Trace("Latest verified header hash could not find in db. Brand new DB.")
				return nil
			}
			latestHeaderHash = common.BytesToHash(latestHeaderHashBytes)
			return nil
		})
	}
	// db is already started so latest epoch must be initialized in store
	return latestHeaderHash
}

func (s *Store) InMemoryLatestVerifiedHeaderHash() common.Hash {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	return s.latestHeaderHash
}

func (s *Store) FindVerifiedSlotNumber(info *types.SlotInfo, fromSlot uint64) uint64 {

	for i := fromSlot; i > 0; i-- {
		slotInfo, err := s.VerifiedSlotInfo(i)
		if err != nil {
			log.WithError(err).Error("failed to find slot info")
			return 0
		}
		if slotInfo.PandoraHeaderHash == info.PandoraHeaderHash && slotInfo.VanguardBlockHash == info.VanguardBlockHash {
			return i
		}
	}
	return 0
}

func (s *Store) RemoveRangeVerifiedInfo(fromSlot, toSlot uint64) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	// storing latest epoch number into db
	return s.db.Update(func(tx *bolt.Tx) error {
		for i := fromSlot; i <= toSlot; i++ {
			bkt := tx.Bucket(verifiedSlotInfosBucket)
			s.verifiedSlotInfoCache.Del(i)
			slotBytes := bytesutil.Uint64ToBytesBigEndian(i)

			if err := bkt.Delete(slotBytes); err != nil {
				return err
			}

		}
		return nil
	})
}
