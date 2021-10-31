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
		log.WithField("from", slotIndex+1).WithField("to", latestVerifiedSlot).Debug("removing consensus info")
		// remove from the dB
		err := s.RemoveRangeVerifiedInfo(slotIndex+1, latestVerifiedSlot)
		if err != nil {
			log.WithError(err).Error("failed to remove verified information")
			return err
		}
		s.latestVerifiedSlot = slotIndex
		info, err := s.VerifiedSlotInfo(slotIndex)
		if err != nil {
			log.WithError(err).Error("failed to retrieve verified slot info")
			return err
		}
		s.latestHeaderHash = info.PandoraHeaderHash

		err = s.SaveLatestVerifiedSlot(s.ctx)
		if err != nil {
			log.WithError(err).Error("failed to save latest verified slot info")
			return err
		}

		err = s.SaveLatestVerifiedHeaderHash()
		if err != nil {
			log.WithError(err).Error("failed to save latest verified header hash")
			return err
		}
	}

	return nil
}

// 5272: skipped
//5273: common.HexToHash("0x92dd18789abb7fbb3a9b7e7b2eee3ee99f6717068a9db6da2f18327f75b10304"),
//5274: common.HexToHash("0x479088bbb0f4a1577f2528fbea41ff3d99373f323769375888ac0bdf9fe506e8"),
//5275: common.HexToHash("0x138d937d02c386e1c47d5550a87deff8652f20683dfd771afa39e3553f596490"),
//5276: common.HexToHash("0xa137cb9c9a22f678994b82c09f31de79994222470c1a77dfc7c78268bb5d0bbc"),
//5278: common.HexToHash("0xcbd44e37125c599ff218b966658d720665e1dfa9fd3db0230ad8743e754495d5"),
//5279: common.HexToHash("0x0acf03ae123dc232e181d3273114b4fc1ae570f469c64655ccb7bc8c6b6aaa28"),
// RemoveForkedL15ProdSlots is an one time execution for resume l15 production testnet
func (s *Store) RemoveForkedL15ProdSlots() error {
	log.WithField("fromInvalidSlot", 5272).WithField("toInvalidSlot", 5279).
		Warn("Discarding verified slots if those slots exist")
	return s.RemoveRangeVerifiedInfo(5272, 5279)
}
