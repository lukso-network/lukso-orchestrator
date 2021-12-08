package types

import (
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

type ShardData struct {
	Number     uint64
	HeaderRoot common.Hash
}

type Shard struct {
	Id     uint64
	Blocks []*ShardData
}

// MultiShardInfo
type MultiShardInfo struct {
	SlotInfo *NewSlotInfo
	Shards   []*Shard
}

// NewSlotInfo contains slot info
type NewSlotInfo struct {
	Slot      uint64
	BlockRoot common.Hash
}

// SlotInfo
type SlotInfo struct {
	VanguardBlockHash common.Hash
	PandoraHeaderHash common.Hash
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
