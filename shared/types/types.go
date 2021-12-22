package types

import (
	"fmt"
	"math/big"
	"strings"

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

func (si *MultiShardInfo) DeepEqual(nsi *MultiShardInfo) bool {
	if nsi == nil {
		return false
	}

	if nsi.IsNil() && si.NotNil() {
		return false
	}

	if nsi.NotNil() && si.IsNil() {
		return false
	}

	if nsi.SlotInfo.Slot != si.SlotInfo.Slot || nsi.SlotInfo.BlockRoot != si.SlotInfo.BlockRoot {
		return false
	}

	if len(nsi.Shards) != len(si.Shards) {
		return false
	}

	nsiShard := nsi.Shards[0]
	siShard := si.Shards[0]

	if nsiShard.Id != siShard.Id || len(nsiShard.Blocks) != len(siShard.Blocks) {
		return false
	}

	nsiShardBlock := nsiShard.Blocks[0]
	siShardBlock := siShard.Blocks[0]

	if nsiShardBlock.Number != siShardBlock.Number || siShardBlock.HeaderRoot != nsiShardBlock.HeaderRoot {
		return false
	}

	return true
}

func (si *MultiShardInfo) NotNil() bool {
	return si.SlotInfo != nil && len(si.Shards) > 0 && len(si.Shards[0].Blocks) > 0
}

func (si *MultiShardInfo) IsNil() bool {
	if si.SlotInfo == nil || len(si.Shards) == 0 {
		return true
	}

	if len(si.Shards[0].Blocks) == 0 {
		return true
	}

	if si.Shards[0].Blocks[0] == nil {
		return true
	}
	return false
}

func (si *MultiShardInfo) FormattedStr() string {
	if si.IsNil() {
		return ""
	}
	s := strings.Join([]string{`shardInfo: { `,
		`slotInfo: { slot: ` + fmt.Sprintf("%d", si.SlotInfo.Slot) + `, blockRoot: ` + fmt.Sprintf("%v", si.SlotInfo.BlockRoot) + `} `,
		`shards: { shardId: ` + fmt.Sprintf("%v", si.Shards[0].Id) + `, shardData: { `,
		`{ number: ` + fmt.Sprintf("%d", si.Shards[0].Blocks[0].Number) + `, headerRoot: ` + fmt.Sprintf("%v", si.Shards[0].Blocks[0].HeaderRoot),
		` }`,
	}, "")
	return s
}

func (si *MultiShardInfo) GetPanShardRootBytes() []byte {
	return si.Shards[0].Blocks[0].HeaderRoot.Bytes()
}

func (si *MultiShardInfo) GetPanShardRoot() common.Hash {
	return si.Shards[0].Blocks[0].HeaderRoot
}

func (si *MultiShardInfo) GetVanSlotRootBytes() []byte {
	return si.SlotInfo.BlockRoot.Bytes()
}

func (si *MultiShardInfo) GetVanSlotRoot() common.Hash {
	return si.SlotInfo.BlockRoot
}

func (si *MultiShardInfo) GetSlot() uint64 {
	return si.SlotInfo.Slot
}

func (si *MultiShardInfo) GetPanBlockNumber() uint64 {
	return si.Shards[0].Blocks[0].Number
}

// reorgStatus holds current reorg status for a certain slot
type ReorgStatus struct {
	Slot            uint64
	BlockRoot       [32]byte
	ParentStepId    uint64
	ParentShardInfo *MultiShardInfo
	PandoraHash     common.Hash
	HasResolved     bool
}

func (rs *ReorgStatus) FormattedStr() string {
	s := strings.Join([]string{`reorgStatus: { `,
		`slot: ` + fmt.Sprintf("%d", rs.Slot) +
			`, blockRoot: ` + fmt.Sprintf("%v", common.BytesToHash(rs.BlockRoot[:])) +
			`, parentStepId: ` + fmt.Sprintf("%v", rs.ParentStepId) +
			`, parentShardInfo: ` + fmt.Sprintf("%v", rs.ParentShardInfo.FormattedStr()) +
			`, pandoraHash: ` + fmt.Sprintf("%v", rs.PandoraHash) +
			`, hasResolved: ` + fmt.Sprintf("%v", rs.HasResolved) + `} `,
	}, "")
	return s
}
