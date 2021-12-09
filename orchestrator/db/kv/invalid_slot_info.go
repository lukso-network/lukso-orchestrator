package kv

import (
	"github.com/boltdb/bolt"
	"github.com/lukso-network/lukso-orchestrator/shared/bytesutil"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

// InvalidSlotInfo
func (s *Store) InvalidSlotInfo(slot uint64) (*types.SlotInfo, error) {
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
	// storing consensus info into cache and db
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(invalidSlotInfosBucket)
		slotBytes := bytesutil.Uint64ToBytesBigEndian(slot)
		enc, err := encode(slotInfo)
		if err != nil {
			return err
		}
		if err := bkt.Put(slotBytes, enc); err != nil {
			return err
		}
		return nil
	})
}
