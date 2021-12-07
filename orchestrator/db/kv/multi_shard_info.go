package kv

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/lukso-network/lukso-orchestrator/shared/bytesutil"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"github.com/pkg/errors"
)

var (
	errShardInfoNotFound = errors.New("shard info not found in db")
)

// VerifiedSlotInfo
func (s *Store) VerifiedShardInfo(stepId uint64) (*types.MultiShardInfo, error) {
	if v, ok := s.multiShardsInfoCache.Get(stepId); v != nil && ok {
		return v.(*types.MultiShardInfo), nil
	}

	var shardInfo *types.MultiShardInfo
	err := s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(multiShardsBucket)
		key := bytesutil.Uint64ToBytesBigEndian(stepId)
		value := bkt.Get(key[:])
		if value == nil {
			return nil
		}
		return decode(value, &shardInfo)
	})
	return shardInfo, err
}

// ConsensusInfos
func (s *Store) VerifiedShardInfos(fromStepId uint64) (map[uint64]*types.MultiShardInfo, error) {
	latestStepId := s.LatestStepID()
	// when requested epoch is greater than stored latest epoch
	if fromStepId > latestStepId {
		return nil, errors.Wrap(errInvalidSlot, fmt.Sprintf("fromStepId: %d", fromStepId))
	}

	shards := make(map[uint64]*types.MultiShardInfo)
	err := s.db.View(func(tx *bolt.Tx) error {
		for step := fromStepId; step <= latestStepId; step++ {
			shardInfo, err := s.VerifiedShardInfo(step)
			if err != nil {
				log.WithField("step", step).Warn("DB is corrupted")
				return err
			}
			shards[step] = shardInfo
		}
		return nil
	})
	// the query not successful
	if err != nil {
		return nil, err
	}

	return shards, nil
}

// SaveVerifiedSlotInfo will insert slot information to particular slot to db and cache
// After save operations you must call SaveLatestVerifiedSlot to push in memory slot height to db
func (s *Store) SaveVerifiedShardInfo(stepId uint64, shardInfo *types.MultiShardInfo) error {
	// storing consensus info into cache and db
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(multiShardsBucket)
		stepIdBytes := bytesutil.Uint64ToBytesBigEndian(stepId)
		enc, err := encode(shardInfo)
		if err != nil {
			return err
		}
		if status := s.multiShardsInfoCache.Set(stepId, shardInfo, 0); !status {
			log.WithField("stepId", stepId).WithField("shardInfo", shardInfo).
				Warn("could not store verified shard info into cache")
		}
		if err := bkt.Put(stepIdBytes, enc); err != nil {
			return err
		}
		return nil
	})
}

// SaveLatestStepID
func (s *Store) SaveLatestStepID(stepID uint64) error {
	// storing latest epoch number into db
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(latestInfoMarkerBucket)
		stepBytes := bytesutil.Uint64ToBytesBigEndian(stepID)
		if err := bkt.Put(curStepId, stepBytes); err != nil {
			return err
		}
		return nil
	})
}

// LatestStepID
func (s *Store) LatestStepID() uint64 {
	var latestStepId uint64
	// Db is not prepared yet. Retrieve latest saved epoch number from db
	if !s.isRunning {
		s.db.View(func(tx *bolt.Tx) error {
			bkt := tx.Bucket(latestInfoMarkerBucket)
			stepIdBytes := bkt.Get(curStepId[:])
			// not found the latest epoch in db. so latest epoch will be zero
			if stepIdBytes == nil {
				latestStepId = uint64(0)
				log.Warn("Could not find latest step id in db. It may happen for brand new DB")
				return nil
			}
			latestStepId = bytesutil.BytesToUint64BigEndian(stepIdBytes)
			return nil
		})
	}
	// db is already started so latest epoch must be initialized in store
	return latestStepId
}

// RemoveShardingInfos removes shard infos from db
func (s *Store) RemoveShardingInfos(fromStepId uint64) error {
	latestStepId := s.LatestStepID()
	// when requested epoch is greater than stored latest epoch
	if fromStepId > latestStepId {
		return errors.Wrap(errInvalidSlot, fmt.Sprintf("fromStepId: %d", fromStepId))
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(multiShardsBucket)
		for step := fromStepId; step <= latestStepId; step++ {
			stepBytes := bytesutil.Uint64ToBytesBigEndian(step)
			s.multiShardsInfoCache.Del(step)
			err := bkt.Delete(stepBytes)
			if err != nil {
				return err
			}
		}
		log.WithField("fromStep", fromStepId).WithField("latestStepId", latestStepId).Info("Reverted shard infos from DB")
		return nil
	})
}

// StoreSlotStepIndex
func (s *Store) SaveSlotStepIndex(slot, stepId uint64) error {
	// storing latest epoch number into db
	return s.db.Update(func(tx *bolt.Tx) error {
		if status := s.slotStepIndexCache.Set(slot, stepId, 0); !status {
			log.WithField("stepId", stepId).WithField("slot", slot).Warn("could not store step id into cache")
		}

		bkt := tx.Bucket(slotStepIndicesBucket)
		key := bytesutil.Uint64ToBytesBigEndian(slot)
		value := bytesutil.Uint64ToBytesBigEndian(stepId)
		if err := bkt.Put(key, value); err != nil {
			return err
		}
		return nil
	})
}

// GetStepIdBySlot
func (s *Store) GetStepIdBySlot(slot uint64) (uint64, error) {
	if v, ok := s.multiShardsInfoCache.Get(slot); v != nil && ok {
		return v.(uint64), nil
	}

	var stepId uint64
	err := s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(slotStepIndicesBucket)
		key := bytesutil.Uint64ToBytesBigEndian(slot)
		stepIdBytes := bkt.Get(key[:])
		// not found the latest epoch in db. so latest epoch will be zero
		if stepIdBytes == nil {
			stepId = uint64(0)
			log.Warn("Could not find latest step id in db. It may happen for brand new DB")
			return nil
		}
		stepId = bytesutil.BytesToUint64BigEndian(stepIdBytes)
		return nil
	})

	if err != nil {
		return stepId, err
	}

	return stepId, nil
}
