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
func (s *Store) ConsensusInfos(fromEpoch uint64) (
	[]*eventTypes.MinimalEpochConsensusInfo, error,
) {
	// when requested epoch is greater than stored latest epoch
	if fromEpoch > s.latestEpoch {
		return nil, errors.Wrap(errInvalidEpoch, fmt.Sprintf("fromEpoch: %d", fromEpoch))
	}

	consensusInfos := make([]*eventTypes.MinimalEpochConsensusInfo, 0)
	err := s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(consensusInfosBucket)
		for epoch := fromEpoch; epoch <= s.latestEpoch; epoch++ {
			// fast finding into cache, if the value does not exist in cache, it starts finding into db
			if v, _ := s.consensusInfoCache.Get(epoch); v != nil {
				consensusInfos = append(consensusInfos, v.(*eventTypes.MinimalEpochConsensusInfo))
				continue
			}
			// preparing key bytes for searching into db
			key := bytesutil.Uint64ToBytesBigEndian(epoch)
			enc := bkt.Get(key[:])
			if enc == nil {
				return errors.Wrap(errInvalidEpoch, fmt.Sprintf("epoch: %d", epoch))
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
		if status := s.consensusInfoCache.Set(consensusInfo.Epoch, consensusInfo, int64(len(enc))); !status {
			log.WithField("epoch", consensusInfo.Epoch).Warn("not set in cache")
		}
		if err := bkt.Put(epochBytes, enc); err != nil {
			return err
		}
		// update latest epoch
		s.latestEpoch = consensusInfo.Epoch
		return nil
	})
}

// LatestSavedEpoch
func (s *Store) LatestSavedEpoch() (uint64, error) {
	// Db is not prepared yet. Retrieve latest saved epoch number from db
	if !s.isRunning {
		err := s.db.View(func(tx *bolt.Tx) error {
			bkt := tx.Bucket(consensusInfosBucket)
			epochBytes := bkt.Get(lastStoredEpochKey[:])
			// not found the latest epoch in db. so latest epoch will be zero
			if epochBytes == nil {
				s.latestEpoch = uint64(0)
				return nil
			}
			s.latestEpoch = bytesutil.BytesToUint64BigEndian(epochBytes)
			return nil
		})
		// got decoding error
		if err != nil {
			return 0, err
		}
	}
	// db is already started so latest epoch must be initialized in store
	return s.latestEpoch, nil
}

// SaveLatestEpoch
func (s *Store) SaveLatestEpoch(ctx context.Context) error {
	// storing latest epoch number into db
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(consensusInfosBucket)
		epochBytes := bytesutil.Uint64ToBytesBigEndian(s.latestEpoch)
		if err := bkt.Put(lastStoredEpochKey, epochBytes); err != nil {
			return err
		}
		return nil
	})
}
