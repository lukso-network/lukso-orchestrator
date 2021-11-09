package kv

import (
	"github.com/ethereum/go-ethereum/common"
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
