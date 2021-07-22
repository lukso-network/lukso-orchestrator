package types

import (
	eth2Types "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
)

type Status string

const (
	Pending  Status = "Pending"
	Verified Status = "Verified"
	Invalid  Status = "Invalid"
	Skipped  Status = "Skipped"
	Unknown  Status = "Unknown"
)

// ExtraData
type ExtraData struct {
	Slot          uint64
	Epoch         uint64
	ProposerIndex uint64
}

// SlotInfo
type SlotInfo struct {
	VanguardBlockHash common.Hash
	PandoraHeaderHash common.Hash
}

// CurrentSlotInfo
type CurrentSlotInfo struct {
	Slot      uint64
	Header    *eth1Types.Header
	ShardInfo *eth2Types.PandoraShard
	Status    Status
}

// CopyHeader creates a deep copy of a block header to prevent side effects from
// modifying a header variable.
func CopyHeader(h *eth1Types.Header) *eth1Types.Header {
	cpy := *h
	if cpy.Difficulty = new(big.Int); h.Difficulty != nil {
		cpy.Difficulty.Set(h.Difficulty)
	}
	if cpy.Number = new(big.Int); h.Number != nil {
		cpy.Number.Set(h.Number)
	}
	if len(h.Extra) > 0 {
		cpy.Extra = make([]byte, len(h.Extra))
		copy(cpy.Extra, h.Extra)
	}
	return &cpy
}
