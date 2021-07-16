package kv

import (
	"github.com/boltdb/bolt"
	"github.com/lukso-network/lukso-orchestrator/shared/bytesutil"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

// InvalidSlotInfo
func (s *Store) InvalidSlotInfo(slot uint64) (*types.SlotInfo, error) {
	if v, ok := s.invalidSlotInfoCache.Get(slot); v != nil && ok {
		return v.(*types.SlotInfo), nil
	}
	var slotInfo *types.SlotInfo
	err := s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(invalidSlotInfosBucket)
		key := bytesutil.Uint64ToBytesBigEndian(slot)
		value := bkt.Get(key[:])
		if value == nil {
			return nil
		}
		return decode(value, &slotInfo)
	})
	return slotInfo, err
}

// SaveInvalidSlotInfo
func (s *Store) SaveInvalidSlotInfo(slot uint64, slotInfo *types.SlotInfo) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	// storing consensus info into cache and db
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(invalidSlotInfosBucket)
		slotBytes := bytesutil.Uint64ToBytesBigEndian(slot)
		enc, err := encode(slotInfo)
		if err != nil {
			return err
		}
		if status := s.invalidSlotInfoCache.Set(slot, slotInfo, 0); !status {
			log.WithField("slot", slot).Warn("could not store invalid slot info into cache")
		}
		if err := bkt.Put(slotBytes, enc); err != nil {
			return err
		}
		return nil
	})
}
