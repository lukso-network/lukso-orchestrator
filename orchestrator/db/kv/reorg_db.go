package kv

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

func (s *Store) RevertConsensusInfo(reorgInfo *types.MinimalEpochConsensusInfoV2) error {
	// remove minimal consensus info
	latestEpoch := s.LatestSavedEpoch()
	if reorgInfo.Epoch <= latestEpoch {
		log.WithField("from", reorgInfo.Epoch).WithField("to", latestEpoch).Debug("removing consensus info")
		err := s.RemoveRangeConsensusInfo(reorgInfo.Epoch, latestEpoch)
		if err != nil {
			log.WithError(err).Error("failed to remove consensus info from database")
			return err
		}
		if reorgInfo.Epoch-1 >= 0 {
			s.latestEpoch = reorgInfo.Epoch - 1
			err := s.SaveLatestEpoch(s.ctx)
			if err != nil {
				log.WithError(err).Error("failed to save latest epoch info")
				return err
			}
		}
	}

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
