package types

import (
	"github.com/ethereum/go-ethereum/common"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	eth2Types "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"time"
)

const BLSSignatureSize = 96

type MinimalEpochConsensusInfo struct {
	Epoch            uint64        `json:"epoch"`
	ValidatorList    []string      `json:"validatorList"`
	EpochStartTime   uint64        `json:"epochTimeStart"`
	SlotTimeDuration time.Duration `json:"slotTimeDuration"`
}

type BlockStatus struct {
	Hash   common.Hash `json:"hash"`
	Status Status      `json:"status"`
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

// VanguardShardInfo
type VanguardShardInfo struct {
	Slot      uint64
	ShardInfo *eth2Types.PandoraShard
	BlockHash []byte
}

type BlsSignatureBytes [BLSSignatureSize]byte

// SlotInfo
type SlotInfoWithStatus struct {
	VanguardBlockHash common.Hash
	PandoraHeaderHash common.Hash
	Status
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
