package utils

import (
	types2 "github.com/prysmaticlabs/eth2-types"
	"time"
)

func SlotStartTime(genesis uint64, slot types2.Slot, secondsPerSlot uint64) time.Time {
	duration := time.Second * time.Duration(slot.Mul(secondsPerSlot))
	startTime := time.Unix(int64(genesis), 0).Add(duration)
	return startTime
}
