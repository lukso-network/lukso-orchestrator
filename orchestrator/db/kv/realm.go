package kv

import (
	"github.com/boltdb/bolt"
	"github.com/lukso-network/lukso-orchestrator/shared/bytesutil"
)

func (s *Store) LatestVerifiedRealmSlot() (slot uint64) {
	if !s.isRunning {
		_ = s.db.View(func(tx *bolt.Tx) (dbErr error) {
			bkt := tx.Bucket(realmBucket)
			latestVerifiedSlot := bkt.Get(latestVerifiedRealmSlot[:])
			// not found the latest block number in db. so latest block number will be zero
			if latestVerifiedSlot == nil {
				log.Trace("Latest Verified Realm Slot header hash could not find in db. It may happen for brand new DB")
				return nil
			}

			if len(latestVerifiedSlot) != 8 {
				log.Errorf(
					"Data fetched from realm db is invalid, latestVerifiedSlot len: %d",
					len(latestVerifiedSlot),
				)
				return
			}

			slot = bytesutil.BytesToUint64BigEndian(latestVerifiedSlot)

			return nil
		})
	}

	return
}

func (s *Store) SaveLatestVerifiedRealmSlot(slot uint64) (err error) {
	err = s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(realmBucket)
		value := bytesutil.Uint64ToBytesBigEndian(slot)
		err = bkt.Put(latestVerifiedRealmSlot, value)

		return err
	})

	return
}
