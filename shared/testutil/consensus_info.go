package testutil

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	eventTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	"time"
)

func NewMinimalConsensusInfo(epoch uint64) *eventTypes.MinimalEpochConsensusInfo {
	validatorList := make([]string, 32)

	for idx := 0; idx < 32; idx++ {
		pubKey := make([]byte, 48)
		validatorList[idx] = hexutil.Encode(pubKey)
	}

	return &eventTypes.MinimalEpochConsensusInfo{
		Epoch:            epoch,
		ValidatorList:    validatorList,
		EpochStartTime:   765544433,
		SlotTimeDuration: time.Duration(6),
	}
}
