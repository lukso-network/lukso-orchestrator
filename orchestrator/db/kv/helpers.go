package kv

import (
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/pkg/errors"
)

func StartSlot (epoch uint64) uint64 {
	slot ,_ := math.SafeMul(uint64(32), epoch)
	return slot
}

// EndSlot returns the last slot number of the
// current epoch.
func EndSlot(epoch uint64) (uint64, error) {
	if epoch == math.MaxUint64 {
		return 0, errors.New("start slot calculation overflows")
	}
	slot := StartSlot(uint64(epoch) + 1)
	return slot - 1, nil
}
