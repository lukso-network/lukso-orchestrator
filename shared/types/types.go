package types

import (
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

type Status int

const (
	Pending Status = iota
	Verified
	Invalid
)

// ExtraData
type ExtraData struct {
	Slot          uint64
	Epoch         uint64
	ProposerIndex uint64
}

// PanBlockHeader
type PanBlockHeader struct {
	Header *eth1Types.Header `json:"header"`
	Status Status            `json:"status"`
}

// CopyPandoraHeader creates a deep copy of a pandora block header to prevent side effects from
// modifying a header variable.
func (panHeader *PanBlockHeader) Copy() *PanBlockHeader {
	cpy := *panHeader
	if cpy.Header.Difficulty = new(big.Int); panHeader.Header.Difficulty != nil {
		cpy.Header.Difficulty.Set(panHeader.Header.Difficulty)
	}
	if cpy.Header.Number = new(big.Int); panHeader.Header.Number != nil {
		cpy.Header.Number.Set(panHeader.Header.Number)
	}
	if len(panHeader.Header.Extra) > 0 {
		cpy.Header.Extra = make([]byte, len(panHeader.Header.Extra))
		copy(cpy.Header.Extra, panHeader.Header.Extra)
	}
	return &cpy
}
