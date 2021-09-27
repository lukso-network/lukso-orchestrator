package kv

import (
	"github.com/boltdb/bolt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/shared/bytesutil"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

var (
	EmptyHash      = common.HexToHash("0000000000000000000000000000000000000000000000000000000000000000")
)

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
			return ErrValueNotFound
		}
		return decode(value, &slotInfo)
	})
	return slotInfo, err
}

// ConsensusInfos
func (s *Store) VerifiedSlotInfos(fromSlot uint64) (map[uint64]*types.SlotInfo, error) {

	slotInfos := make(map[uint64]*types.SlotInfo)
	err := s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(verifiedSlotInfosBucket)
		cursor := bkt.Cursor()

		key := bytesutil.Uint64ToBytesBigEndian(fromSlot)
		slotNumber, slotData := cursor.Seek(key)
		for slotNumber != nil && slotData != nil {
			var slotInfo *types.SlotInfo
			err := decode(slotData, &slotInfo)
			if err != nil {
				return err
			}
			slot := bytesutil.BytesToUint64BigEndian(slotNumber)
			slotInfos[slot] = slotInfo
			slotNumber, slotData = cursor.Next()
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

func (s *Store) GetLatestVerifiedSlotInfo() (uint64, *types.SlotInfo, error) {
	var slotInfo *types.SlotInfo
	var slotNumber uint64
	err := s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(verifiedSlotInfosBucket)
		cursor := bkt.Cursor()
		key, value := cursor.Last()
		if value == nil || key == nil {
			return ErrValueNotFound
		}
		slotNumber = bytesutil.BytesToUint64BigEndian(key)
		decodeError := decode(value, &slotInfo)
		for decodeError != nil {
			key, value := cursor.Prev()
			if key == nil {
				// first element has reached. So just return error
				return decodeError
			}
			slotNumber = bytesutil.BytesToUint64BigEndian(key)
			decodeError = decode(value, &slotInfo)
		}
		return nil
	})
	return slotNumber, slotInfo, err
}

func (s *Store) GetFirstVerifiedSlotNumber (fromSlot uint64) (uint64, error) {
	var derivedSlot uint64
	err := s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(verifiedSlotInfosBucket)
		cursor := bkt.Cursor()

		key := bytesutil.Uint64ToBytesBigEndian(fromSlot)
		slotNumber, _ := cursor.Seek(key)
		if slotNumber == nil {
			return ErrValueNotFound
		}
		derivedSlot = bytesutil.BytesToUint64BigEndian(slotNumber)
		return nil
	})
	return derivedSlot, err
}

// LatestSavedEpoch
func (s *Store) LatestSavedVerifiedSlot() uint64 {
	slot, _, err := s.GetLatestVerifiedSlotInfo()
	if err != nil {
		return 0
	}
	// db is already started so latest epoch must be initialized in store
	return slot
}

// LatestSavedEpoch
func (s *Store) LatestVerifiedHeaderHash() common.Hash {
	latestHeaderHash := EmptyHash
	_, slotInfo, err := s.GetLatestVerifiedSlotInfo()
	if err == nil {
		latestHeaderHash = slotInfo.PandoraHeaderHash
	}
	return latestHeaderHash
}

func (s *Store) removeSlotInfoFromVerifiedDB(slot uint64) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(verifiedSlotInfosBucket)
		key := bytesutil.Uint64ToBytesBigEndian(slot)
		s.verifiedSlotInfoCache.Del(slot)
		if err := bkt.Delete(key); err != nil {
			return err
		}
		return nil
	})
}
