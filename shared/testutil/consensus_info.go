package testutil

import (
	"github.com/ethereum/go-ethereum/common"
	eventTypes "github.com/lukso-network/lukso-orchestrator/shared/types"
	types "github.com/prysmaticlabs/eth2-types"
	"math/rand"
	"strconv"
)

func NewMinimalConsensusInfo(epoch types.Epoch) *eventTypes.MinConsensusInfoEvent {
	validatorList := make([]string, 32)

	for idx := 0; idx < 32; idx++ {
		bs := []byte(strconv.Itoa(31415926))
		pubKey := common.Bytes2Hex(bs)
		validatorList[idx] = pubKey
	}
	return &eventTypes.MinConsensusInfoEvent{
		Epoch:            epoch,
		ValidatorList:    validatorList,
		EpochStartTime:   rand.Uint64(),
		SlotTimeDuration: uint64(6),
	}
}