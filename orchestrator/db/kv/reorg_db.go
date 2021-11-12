package kv

import (
	"github.com/boltdb/bolt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/shared/bytesutil"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

func (s *Store) RevertConsensusInfo(reorgInfo *types.MinimalEpochConsensusInfoV2) error {
	// remove from verified database
	slotInfo := &types.SlotInfo{
		PandoraHeaderHash: common.BytesToHash(reorgInfo.ReorgInfo.PanParentHash),
		VanguardBlockHash: common.BytesToHash(reorgInfo.ReorgInfo.VanParentHash),
	}
	latestVerifiedSlot := s.LatestSavedVerifiedSlot()

	slotIndex := s.FindVerifiedSlotNumber(slotInfo, latestVerifiedSlot)
	if slotIndex > 0 {
		log.WithField("from", slotIndex+1).WithField("skip", reorgInfo.ReorgInfo.NewSlot).Debug("removing verified info")
		// remove from the dB
		err := s.RemoveRangeVerifiedInfo(slotIndex+1, reorgInfo.ReorgInfo.NewSlot)
		if err != nil {
			log.WithError(err).Error("failed to remove verified information")
			return err
		}
	}

	return nil
}

// SaveLatestFinalizedSlot
func (s *Store) SaveLatestFinalizedSlot(latestFinalizedSlot uint64) error {
	// storing latest finalized slot number into db
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(latestInfoMarkerBucket)
		slotBytes := bytesutil.Uint64ToBytesBigEndian(latestFinalizedSlot)
		if err := bkt.Put(latestFinalizedSlotKey, slotBytes); err != nil {
			return err
		}
		return nil
	})
}

// LatestLatestFinalizedSlot
func (s *Store) LatestLatestFinalizedSlot() uint64 {
	var latestFinalizedSlot uint64
	// Db is not prepared yet. Retrieve latest saved finalized slot number from db
	if !s.isRunning {
		s.db.View(func(tx *bolt.Tx) error {
			bkt := tx.Bucket(latestInfoMarkerBucket)
			slotBytes := bkt.Get(latestFinalizedSlotKey[:])
			// not found the latest finalized slot in db. so latest finalized slot will be zero
			if slotBytes == nil {
				latestFinalizedSlot = uint64(0)
				log.Trace("Latest finalized slot number could not find in db. It may happen for brand new DB")
				return nil
			}
			latestFinalizedSlot = bytesutil.BytesToUint64BigEndian(slotBytes)
			return nil
		})
	}
	// db is already started so latest finalized slot must be initialized in store
	return latestFinalizedSlot
}

// SaveLatestFinalizedEpoch
func (s *Store) SaveLatestFinalizedEpoch(latestFinalizedEpoch uint64) error {
	// storing latest finalized slot number into db
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(latestInfoMarkerBucket)
		epochBytes := bytesutil.Uint64ToBytesBigEndian(latestFinalizedEpoch)
		if err := bkt.Put(latestFinalizedEpochKey, epochBytes); err != nil {
			return err
		}
		return nil
	})
}

// LatestLatestFinalizedEpoch
func (s *Store) LatestLatestFinalizedEpoch() uint64 {
	var latestFinalizedEpoch uint64
	// Db is not prepared yet. Retrieve latest saved finalized slot number from db
	if !s.isRunning {
		s.db.View(func(tx *bolt.Tx) error {
			bkt := tx.Bucket(latestInfoMarkerBucket)
			epochBytes := bkt.Get(latestFinalizedEpochKey[:])
			// not found the latest finalized slot in db. so latest finalized slot will be zero
			if epochBytes == nil {
				latestFinalizedEpoch = uint64(0)
				log.Trace("Latest finalized epoch number could not find in db. It may happen for brand new DB")
				return nil
			}
			latestFinalizedEpoch = bytesutil.BytesToUint64BigEndian(epochBytes)
			return nil
		})
	}
	// db is already started so latest finalized slot must be initialized in store
	return latestFinalizedEpoch
}

func (s *Store) UpdateVerifiedSlotInfo(slot uint64) error {
	slotNumber, slotInfo, err := s.SeekSlotInfo(slot)
	if err != nil {
		return err
	}
	err = s.SaveLatestVerifiedSlot(s.ctx, slotNumber)
	if err != nil {
		return err
	}

	err = s.SaveLatestVerifiedHeaderHash(slotInfo.PandoraHeaderHash)
	if err != nil {
		return err
	}
	return nil
}
