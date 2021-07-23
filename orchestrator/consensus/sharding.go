package consensus

import (
	"github.com/ethereum/go-ethereum/common"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	eth2Types "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
)

func CompareShardingInfo(ph *eth1Types.Header, vs *eth2Types.PandoraShard) bool {
	if ph == nil && vs == nil {
		// in existing code this will happen. as some part may have no sharding info for testing.
		return true
	}

	if vs.BlockNumber != ph.Number.Uint64() {
		log.WithField("pandora data block number", ph.Number.Uint64()).
			WithField("vanguard block number", vs.BlockNumber).
			Error("block number mismatched")
		return false
	}

	// match header hash
	if ph.Hash() != common.BytesToHash(vs.GetHash()) {
		log.WithField("pandora header hash", ph.Hash().Bytes()).
			WithField("vanguard header hash", vs.GetHash()).
			Error("header hash mismatched")
		return false
	}

	// match parent hash
	if ph.ParentHash != common.BytesToHash(vs.GetParentHash()) {
		log.WithField("pandora data parent hash", ph.ParentHash.Bytes()).
			WithField("vanguard parent hash", vs.ParentHash).
			Error("parent hash mismatched")
		return false
	}

	// match state root hash
	if ph.Root != common.BytesToHash(vs.GetStateRoot()) {
		log.WithField("pandora data root hash", ph.Root.Bytes()).
			WithField("vanguard state root hash", vs.StateRoot).
			Error("state root hash mismatched")
		return false
	}

	// match TxHash
	if ph.TxHash != common.BytesToHash(vs.GetTxHash()) {
		log.WithField("pandora data tx hash", ph.TxHash.Bytes()).
			WithField("vanguard tx hash", vs.TxHash).
			Error("tx hash mismatched")
		return false
	}

	// match receiptHash
	if ph.ReceiptHash != common.BytesToHash(vs.GetReceiptHash()) {
		log.WithField("pandora data receipt hash", ph.ReceiptHash.Bytes()).
			WithField("vanguard receipt hash", vs.ReceiptHash).
			Error("receipt hash mismatched")
		return false
	}

	// retrieve extra data
	pandoraExtraDataWithSig := new(types.PanExtraDataWithBLSSig)
	err := rlp.DecodeBytes(ph.Extra, pandoraExtraDataWithSig)
	if nil != err {
		log.WithField("error", err).
			Error("error converting extra data to extraDataWithSig")
		return false
	}

	// match signature
	if pandoraExtraDataWithSig.BlsSignatureBytes != types.BytesToSig(vs.GetSignature()) {
		log.WithField("pandora data signature", pandoraExtraDataWithSig.BlsSignatureBytes.Bytes()).
			WithField("vanguard signature", vs.GetSignature()).
			Error("signature mismatched")
		return false
	}

	return true
}
