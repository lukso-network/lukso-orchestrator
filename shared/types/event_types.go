package types

import "time"

type MinimalEpochConsensusInfo struct {
	Epoch            uint64        `json:"epoch"`
	ValidatorList    [32]string    `json:"validatorList"`
	EpochStartTime   uint64        `json:"epochTimeStart"`
	SlotTimeDuration time.Duration `json:"slotTimeDuration"`
}
