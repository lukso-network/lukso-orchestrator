package types

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	eth2Types "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
)

const BLSSignatureSize = 96

type Reorg struct {
	VanParentHash []byte `json:"van_parent_hash"`
	PanParentHash []byte `json:"pan_parent_hash"`
	NewSlot       uint64 `json:"new_slot"`
}

type MinimalEpochConsensusInfoV2 struct {
	Epoch            uint64        `json:"epoch"`
	ValidatorList    []string      `json:"validatorList"`
	EpochStartTime   uint64        `json:"epochTimeStart"`
	SlotTimeDuration time.Duration `json:"slotTimeDuration"`
	ReorgInfo        *Reorg        `json:"reorg_info"`
	FinalizedSlot    uint64        `json:"finalizedSlot"`
}

type MinimalEpochConsensusInfo struct {
	Epoch            uint64        `json:"epoch"`
	ValidatorList    []string      `json:"validatorList"`
	EpochStartTime   uint64        `json:"epochTimeStart"`
	SlotTimeDuration time.Duration `json:"slotTimeDuration"`
}

type BlockStatus struct {
	Hash          common.Hash `json:"hash"`
	Status        Status      `json:"status"`
	FinalizedSlot uint64      `json:"finalizedSlot"`
}

// PandoraPendingHeaderFilter
type PandoraPendingHeaderFilter struct {
	FromBlockHash common.Hash `json:"fromBlockHash"`
}

// PanExtraDataWithBLSSig
type PanExtraDataWithBLSSig struct {
	ExtraData
	BlsSignatureBytes BlsSignatureBytes
}

// PandoraHeaderInfo
type PandoraHeaderInfo struct {
	Slot   uint64
	Header *eth1Types.Header
}

type ShutDownSignal struct {
	Shutdown bool
}

// VanguardShardInfo
type VanguardShardInfo struct {
	Slot           uint64
	ShardInfo      *eth2Types.PandoraShard
	BlockRoot      [32]byte
	FinalizedSlot  uint64
	FinalizedEpoch uint64
	ParentRoot     [32]byte
}

func CopyVanguardShardInfo(data *VanguardShardInfo) *VanguardShardInfo {
	cpy := *data
	if cpy.ShardInfo = new(eth2Types.PandoraShard); data.ShardInfo != nil {
		cpy.ShardInfo.Hash = data.ShardInfo.GetHash()
		cpy.ShardInfo.ParentHash = data.ShardInfo.GetParentHash()
		cpy.ShardInfo.BlockNumber = data.ShardInfo.GetBlockNumber()
		cpy.ShardInfo.ReceiptHash = data.ShardInfo.GetReceiptHash()
		cpy.ShardInfo.SealHash = data.ShardInfo.GetSealHash()
		cpy.ShardInfo.TxHash = data.ShardInfo.GetTxHash()
		cpy.ShardInfo.StateRoot = data.ShardInfo.GetStateRoot()
		cpy.ShardInfo.Signature = data.ShardInfo.GetSignature()
	}
	return &cpy
}

type BlsSignatureBytes [BLSSignatureSize]byte

// SlotInfo
type SlotInfoWithStatus struct {
	VanguardBlockHash common.Hash
	PandoraHeaderHash common.Hash
	Status
}

func (info *MinimalEpochConsensusInfoV2) ConvertToEpochInfo() *MinimalEpochConsensusInfo {
	return &MinimalEpochConsensusInfo{
		Epoch:            info.Epoch,
		ValidatorList:    info.ValidatorList,
		EpochStartTime:   info.EpochStartTime,
		SlotTimeDuration: info.SlotTimeDuration,
	}
}

func (info *MinimalEpochConsensusInfo) ConvertToEpochInfoV2() *MinimalEpochConsensusInfoV2 {
	return &MinimalEpochConsensusInfoV2{
		Epoch:            info.Epoch,
		ValidatorList:    info.ValidatorList,
		EpochStartTime:   info.EpochStartTime,
		SlotTimeDuration: info.SlotTimeDuration,
		ReorgInfo:        nil,
	}
}

// Bytes gets the byte representation of the underlying hash.
func (h BlsSignatureBytes) Bytes() []byte { return h[:] }

func BytesToSig(b []byte) BlsSignatureBytes {
	var bls BlsSignatureBytes
	bls.SetBytes(b)
	return bls
}

func (bls *BlsSignatureBytes) SetBytes(b []byte) {
	if len(b) > len(bls) {
		b = b[len(b)-BLSSignatureSize:]
	}

	copy(bls[BLSSignatureSize-len(b):], b)
}
