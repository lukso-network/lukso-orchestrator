package testutil

import (
	"github.com/ethereum/go-ethereum/common"
	eventTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	"strconv"
	"time"
)

func NewMinimalConsensusInfo(epoch uint64) *eventTypes.MinimalEpochConsensusInfo {
	validatorList := make([]string, 32)

	for idx := 0; idx < 32; idx++ {
		bs := []byte(strconv.Itoa(31415926))
		pubKey := common.Bytes2Hex(bs)
		validatorList[idx] = pubKey
	}

	return &eventTypes.MinimalEpochConsensusInfo{
		Epoch:            epoch,
		ValidatorList:    validatorList,
		EpochStartTime:   765544433,
		SlotTimeDuration: time.Duration(6),
	}
}
