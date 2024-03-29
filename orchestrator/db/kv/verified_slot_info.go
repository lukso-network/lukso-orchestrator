package kv

import (
	"context"
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/shared/bytesutil"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"github.com/pkg/errors"
)

var (
	EmptyHash      = common.HexToHash("0000000000000000000000000000000000000000000000000000000000000000")
	errInvalidSlot = errors.New("invalid slot and not found any verified slot info for the given slot")
)

func (s *Store) SeekSlotInfo(slot uint64) (uint64, *types.SlotInfo, error) {
	var slotInfo *types.SlotInfo
	var foundSlot uint64
	err := s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(verifiedSlotInfosBucket)
		for i := int64(slot); i > 0; i-- {
			slotInBytes := bytesutil.Uint64ToBytesBigEndian(uint64(i))
			info := bkt.Get(slotInBytes)
			if info == nil {
				continue
			}
			err := decode(info, &slotInfo)
			if err != nil {
				return err
			}
			foundSlot = uint64(i)
			break
		}
		return nil
	})
	if slotInfo != nil {
		log.Debug("seekSlotInfo is returning", "foundSlot", foundSlot, "slotInfo.PandoraHash", slotInfo.PandoraHeaderHash, "slotInfo.VanHash", slotInfo.VanguardBlockHash, "error", err)
	}
	return foundSlot, slotInfo, err
}

// VerifiedSlotInfo
func (s *Store) VerifiedSlotInfo(slot uint64) (*types.SlotInfo, error) {
	if v, ok := s.verifiedSlotInfoCache.Get(slot); v != nil && ok {
		return v.(*types.SlotInfo), nil
	}
	var slotInfo *types.SlotInfo
	err := s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(verifiedSlotInfosBucket)
		key := bytesutil.Uint64ToBytesBigEndian(slot)
		value := bkt.Get(key[:])
		if value == nil {
			return nil
		}
		return decode(value, &slotInfo)
	})
	return slotInfo, err
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

// SaveVerifiedSlotInfo will insert slot information to particular slot to db and cache
// After save operations you must call SaveLatestVerifiedSlot to push in memory slot height to db
func (s *Store) SaveVerifiedSlotInfo(slot uint64, slotInfo *types.SlotInfo) error {
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
		return nil
	})
}

// SaveLatestEpoch
func (s *Store) SaveLatestVerifiedSlot(ctx context.Context, slot uint64) error {
	// storing latest epoch number into db
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(latestInfoMarkerBucket)
		slotBytes := bytesutil.Uint64ToBytesBigEndian(slot)
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
			bkt := tx.Bucket(latestInfoMarkerBucket)
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

// SaveLatestEpoch
func (s *Store) SaveLatestVerifiedHeaderHash(hash common.Hash) error {
	// storing latest epoch number into db
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(latestInfoMarkerBucket)
		headerHashBytes := hash.Bytes()
		if err := bkt.Put(latestHeaderHashKey, headerHashBytes); err != nil {
			return err
		}
		return nil
	})
}

// LatestVerifiedHeaderHash should return latest verified header hash but I really dont know which (pandora or vanguard?)
// It should say explicitly which hash its returning, it looks like its pandora hash
func (s *Store) LatestVerifiedHeaderHash() common.Hash {
	var latestHeaderHash common.Hash
	// Db is not prepared yet. Retrieve latest saved epoch number from db
	if !s.isRunning {
		s.db.View(func(tx *bolt.Tx) error {
			bkt := tx.Bucket(latestInfoMarkerBucket)
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

// FindVerifiedSlotNumber will try to find matching of verified slot info
// fromSlot must be higher or equal slot number that is present in db
// TODO: consider not returning 0 when slot was not found, instead extend this function with multiple return
func (s *Store) FindVerifiedSlotNumber(info *types.SlotInfo, fromSlot uint64) uint64 {
	for i := fromSlot; i > 0; i-- {
		slotInfo, err := s.VerifiedSlotInfo(i)
		if err != nil {
			log.WithError(err).Error("failed to find slot info")
			return 0
		}
		if slotInfo != nil && slotInfo.PandoraHeaderHash == info.PandoraHeaderHash && slotInfo.VanguardBlockHash == info.VanguardBlockHash {
			return i
		}
	}
	return 0
}

// RemoveRangeVerifiedInfo method deletes [fromSlot, latestVerifiedSlot]
func (s *Store) RemoveRangeVerifiedInfo(fromSlot, toSlot uint64) error {
	log.WithField("fromSlot", fromSlot).WithField("toSlot", toSlot).
		Debug("Start removing slot infos from verified db!")

	// storing latest epoch number into db
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(verifiedSlotInfosBucket)

		for slotNum := fromSlot; slotNum <= toSlot; slotNum++ {
			removingSlotNumber := bytesutil.Uint64ToBytesBigEndian(slotNum)
			s.verifiedSlotInfoCache.Del(slotNum)
			err := bkt.Delete(removingSlotNumber)
			if err != nil {
				return err
			}
		}
		log.Debug("success:: all slots are removed from the verified database")
		return nil
	})
}
