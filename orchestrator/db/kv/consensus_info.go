package kv

import (
	"context"
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/lukso-network/lukso-orchestrator/shared/bytesutil"
	eventTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	"github.com/pkg/errors"
)

var errInvalidEpoch = errors.New("invalid epoch and not found any consensusInfo for the given epoch")

// ConsensusInfo
func (s *Store) ConsensusInfo(ctx context.Context, epoch uint64) (*eventTypes.MinimalEpochConsensusInfo, error) {
	// Return consensus info from cache if it exists.
	if v, ok := s.consensusInfoCache.Get(epoch); v != nil && ok {
		return v.(*eventTypes.MinimalEpochConsensusInfo), nil
	}
	// consensus info not found in cache so retrieve from db
	var consensusInfo *eventTypes.MinimalEpochConsensusInfo
	err := s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(consensusInfosBucket)
		key := bytesutil.Uint64ToBytesBigEndian(epoch)
		enc := bkt.Get(key[:])
		if enc == nil {
			return nil
		}
		return decode(enc, &consensusInfo)
	})
	return consensusInfo, err
}

// ConsensusInfos
func (s *Store) ConsensusInfos(fromEpoch uint64) ([]*eventTypes.MinimalEpochConsensusInfo, error) {
	latestEpoch := s.LatestSavedEpoch()
	// when requested epoch is greater than stored latest epoch
	if fromEpoch > latestEpoch {
		return nil, errors.Wrap(errInvalidEpoch, fmt.Sprintf("fromEpoch: %d", fromEpoch))
	}

	consensusInfos := make([]*eventTypes.MinimalEpochConsensusInfo, 0)
	err := s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(consensusInfosBucket)
		for epoch := fromEpoch; epoch <= latestEpoch; epoch++ {
			// fast finding into cache, if the value does not exist in cache, it starts finding into db
			if v, _ := s.consensusInfoCache.Get(epoch); v != nil {
				consensusInfos = append(consensusInfos, v.(*eventTypes.MinimalEpochConsensusInfo))
				continue
			}
			// preparing key bytes for searching into db
			key := bytesutil.Uint64ToBytesBigEndian(epoch)
			enc := bkt.Get(key[:])
			if enc == nil {
				return nil
			}
			var consensusInfo *eventTypes.MinimalEpochConsensusInfo
			decode(enc, &consensusInfo)
			consensusInfos = append(consensusInfos, consensusInfo)
		}
		return nil
	})
	// the query not successful
	if err != nil {
		return nil, err
	}

	return consensusInfos, nil
}

// SaveConsensusInfo
func (s *Store) SaveConsensusInfo(
	ctx context.Context,
	consensusInfo *eventTypes.MinimalEpochConsensusInfo,
) error {
	// storing consensus info into cache and db
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(consensusInfosBucket)
		epochBytes := bytesutil.Uint64ToBytesBigEndian(consensusInfo.Epoch)
		enc, err := encode(consensusInfo)
		if err != nil {
			return err
		}
		if status := s.consensusInfoCache.Set(consensusInfo.Epoch, consensusInfo, 0); !status {
			log.WithField("epoch", consensusInfo.Epoch).Warn("not set in cache")
		}
		if err := bkt.Put(epochBytes, enc); err != nil {
			return err
		}
		// update latest epoch
		return nil
	})
}

func (s *Store) RemoveRangeConsensusInfo(startEpoch, endEpoch uint64) error {

	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(consensusInfosBucket)
		for i := startEpoch; i <= endEpoch; i++ {
			s.consensusInfoCache.Del(i)
			epochBytes := bytesutil.Uint64ToBytesBigEndian(i)
			if err := bkt.Delete(epochBytes); err != nil {
				return err
			}
		}
		return nil
	})
}

// LatestSavedEpoch
func (s *Store) LatestSavedEpoch() uint64 {
	var latestSavedEpoch uint64
	// Db is not prepared yet. Retrieve latest saved epoch number from db
	if !s.isRunning {
		s.db.View(func(tx *bolt.Tx) error {
			bkt := tx.Bucket(latestInfoMarkerBucket)
			epochBytes := bkt.Get(lastStoredEpochKey[:])
			// not found the latest epoch in db. so latest epoch will be zero
			if epochBytes == nil {
				latestSavedEpoch = uint64(0)
				log.Trace("Latest epoch could not find in db. It may happen for brand new DB")
				return nil
			}
			latestSavedEpoch = bytesutil.BytesToUint64BigEndian(epochBytes)
			return nil
		})
	}
	// db is already started so latest epoch must be initialized in store
	return latestSavedEpoch
}

// SaveLatestEpoch
func (s *Store) SaveLatestEpoch(ctx context.Context, epoch uint64) error {
	// storing latest epoch number into db
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(latestInfoMarkerBucket)
		epochBytes := bytesutil.Uint64ToBytesBigEndian(epoch)
		if err := bkt.Put(lastStoredEpochKey, epochBytes); err != nil {
			return err
		}
		return nil
	})
}
