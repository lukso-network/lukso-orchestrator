package kv

import (
	"context"
	"github.com/boltdb/bolt"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db/filters"
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
		epochBytes := bytesutil.Uint64ToBytesBigEndian(epoch)
		enc := bkt.Get(epochBytes[:])
		if enc == nil {
			return nil
		}
		return decode(enc, &consensusInfo)
	})
	return consensusInfo, err
}

// ConsensusInfos
func (s *Store) ConsensusInfos(ctx context.Context, f *filters.QueryFilter) (
	[]*eventTypes.MinimalEpochConsensusInfo, error,
) {

	return nil, nil
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
		s.consensusInfoCache.Set(consensusInfo.Epoch, consensusInfo, int64(len(enc)))
		if err := bkt.Put(epochBytes, enc); err != nil {
			return err
		}
		return nil
	})
}
