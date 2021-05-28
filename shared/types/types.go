package types

import (
	"github.com/ethereum/go-ethereum/common"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

type Status int

const (
	Pending Status = iota
	Verified
	Invalid
	Skipped
)

// ExtraData
type ExtraData struct {
	Slot          uint64
	Epoch         uint64
	ProposerIndex uint64
}

// generic header hash
type HeaderHash struct {
	HeaderHash common.Hash `json:"headerHash"`
	Status     Status      `json:"status"`
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
