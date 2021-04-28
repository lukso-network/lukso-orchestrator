package types

import (
	eth1Types "github.com/ethereum/go-ethereum/core/types"
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
