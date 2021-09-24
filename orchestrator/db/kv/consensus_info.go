package kv

import (
	"context"
	"github.com/boltdb/bolt"
	"github.com/lukso-network/lukso-orchestrator/shared/bytesutil"
	eventTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
)


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
			return ErrValueNotFound
		}
		return decode(enc, &consensusInfo)
	})
	return consensusInfo, err
}

func (s *Store) GetLatestConsensusInfo() (*eventTypes.MinimalEpochConsensusInfo, error) {
	var consensusInfo *eventTypes.MinimalEpochConsensusInfo
	err := s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(consensusInfosBucket)
		cursor := bkt.Cursor()
		_, enc := cursor.Last()
		if enc == nil {
			return ErrValueNotFound
		}
		return decode(enc, &consensusInfo)
	})
	return consensusInfo, err
}

// ConsensusInfos
func (s *Store) ConsensusInfos(fromEpoch uint64) (
	[]*eventTypes.MinimalEpochConsensusInfo, error,
) {
	consensusInfos := make([]*eventTypes.MinimalEpochConsensusInfo, 0)
	err := s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(consensusInfosBucket)
		cursor := bkt.Cursor()
		key := bytesutil.Uint64ToBytesBigEndian(fromEpoch)
		epoch, mininalConsInfo := cursor.Seek(key)
		for epoch != nil {
			var consensusInfo *eventTypes.MinimalEpochConsensusInfo
			err := decode(mininalConsInfo, &consensusInfo)
			if err != nil {
				return err
			}
			consensusInfos = append(consensusInfos, consensusInfo)
			epoch, mininalConsInfo = cursor.Next()
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
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

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
		s.latestEpoch = consensusInfo.Epoch
		return nil
	})
}

// LatestSavedEpoch
func (s *Store) LatestSavedEpoch() uint64 {
	info, err := s.GetLatestConsensusInfo()
	if err != nil || info == nil {
		return 0
	}
	// db is already started so latest epoch must be initialized in store
	return info.Epoch
}

func (s *Store) removeConsensusInfoDb (epoch uint64)  error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	// storing consensus info into cache and db
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(consensusInfosBucket)
		epochBytes := bytesutil.Uint64ToBytesBigEndian(epoch)
		s.consensusInfoCache.Del(epoch)
		if err := bkt.Delete(epochBytes); err != nil {
			return err
		}
		return nil
	})
}

