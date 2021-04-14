package kv

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"github.com/pkg/errors"
)

var EmptyExtraData = types.ExtraData{}

func SlotFromExtraData(header *types.PanBlockHeader) (uint64, error) {
	var extraData types.ExtraData
	if err := rlp.DecodeBytes(header.Header.Extra, &extraData); err != nil {
		return 0, errors.Wrap(err, "Failed to decode extra data fields")
	}
	if extraData == EmptyExtraData {
		return 0, errors.Wrap(InvalidExtraDataErr, "Got nil in extra data field")
	}
	return extraData.Slot, nil
}
