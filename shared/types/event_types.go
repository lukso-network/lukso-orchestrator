package types

import types "github.com/prysmaticlabs/eth2-types"

type MinConsensusInfoEvent struct {
	Epoch				types.Epoch
	ValidatorList 		[]string
	EpochStartTime 		uint64
	SlotTimeDuration 	uint64
}
