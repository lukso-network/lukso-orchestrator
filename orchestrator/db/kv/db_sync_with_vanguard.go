package kv

import (
	"errors"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

func (s *Store) RemoveInfoFromAllDb(fromEpoch, toEpoch uint64) error {
	err := s.rangeRemoveConsensusInfoDb(fromEpoch, toEpoch)
	if err != nil {
		return err
	}
	startSlot := StartSlot(fromEpoch)
	endSlot, err := EndSlot(toEpoch)
	if err != nil {
		return err
	}
	log.WithField("start slot", startSlot).WithField("end slot", endSlot).Debug("removing info")
	err = s.RangeRemoveSlotInfo(startSlot, endSlot)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) GetFirstVerifiedSlotInAnEpoch (epoch uint64) (*types.SlotInfo, error) {
	startSlot := StartSlot(epoch)
	endSlot, err := EndSlot(epoch)
	if err != nil {
		return nil, err
	}
	slotNo, err := s.GetFirstVerifiedSlotNumber(startSlot)
	if err != nil {
		return nil, err
	}
	if slotNo >= startSlot && slotNo <= endSlot {
		// slot is within the range of the epoch
		info, err := s.VerifiedSlotInfo(slotNo)
		if err != nil {
			return nil, err
		}
		return info, nil
	}
	return nil, errors.New("no slot found in this epoch")
}

func (s *Store) RemoveSlotInfo (slot uint64) error {
	err := s.removeSlotInfoFromVerifiedDB(slot)
	if err != nil {
		return err
	}
	err = s.removeSlotInfoFromInvalidDB(slot)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) RangeRemoveSlotInfo (startSlot, endSlot uint64) error {
	err := s.rangeRemoveSlotInfoFromVerifiedDB(startSlot, endSlot)
	if err != nil {
		return err
	}
	err = s.rangeRemoveSlotInfoFromInvalidDB(startSlot, endSlot)
	if err != nil {
		return err
	}
	return nil
}