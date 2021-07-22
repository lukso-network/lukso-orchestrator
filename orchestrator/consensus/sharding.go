package consensus

import (
	"reflect"

	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	eth2Types "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
)

func CompareShardingInfo(pandoraHeaderData *eth1Types.Header, vanguardShardInfo *eth2Types.PandoraShard) bool {
	if pandoraHeaderData == nil && vanguardShardInfo == nil {
		// in existing code this will happen. as some part may have no sharding info for testing.
		return true
	}

	if vanguardShardInfo.BlockNumber != pandoraHeaderData.Number.Uint64() {
		log.WithField("pandora data block number", pandoraHeaderData.Number.Uint64()).
			WithField("vanguard block number", vanguardShardInfo.BlockNumber).
			Error("block number mismatched")
		return false
	}

	// match header hash
	if !reflect.DeepEqual(pandoraHeaderData.Hash().Bytes(), vanguardShardInfo.GetHash()) {
		log.WithField("pandora header hash", pandoraHeaderData.Hash().Bytes()).
			WithField("vanguard header hash", vanguardShardInfo.GetHash()).
			Error("header hash mismatched")
		return false
	}

	// match parent hash
	if !reflect.DeepEqual(pandoraHeaderData.ParentHash.Bytes(), vanguardShardInfo.GetParentHash()) {
		log.WithField("pandora data parent hash", pandoraHeaderData.ParentHash.Bytes()).
			WithField("vanguard parent hash", vanguardShardInfo.ParentHash).
			Error("parent hash mismatched")
		return false
	}

	// match state root hash
	if !reflect.DeepEqual(pandoraHeaderData.Root.Bytes(), vanguardShardInfo.GetStateRoot()) {
		log.WithField("pandora data root hash", pandoraHeaderData.Root.Bytes()).
			WithField("vanguard state root hash", vanguardShardInfo.StateRoot).
			Error("state root hash mismatched")
		return false
	}

	// match TxHash
	if !reflect.DeepEqual(pandoraHeaderData.TxHash.Bytes(), vanguardShardInfo.GetTxHash()) {
		log.WithField("pandora data tx hash", pandoraHeaderData.TxHash.Bytes()).
			WithField("vanguard tx hash", vanguardShardInfo.TxHash).
			Error("tx hash mismatched")
		return false
	}

	// match receiptHash
	if !reflect.DeepEqual(pandoraHeaderData.ReceiptHash.Bytes(), vanguardShardInfo.GetReceiptHash()) {
		log.WithField("pandora data receipt hash", pandoraHeaderData.ReceiptHash.Bytes()).
			WithField("vanguard receipt hash", vanguardShardInfo.ReceiptHash).
			Error("receipt hash mismatched")
		return false
	}

	// retrieve extra data
	pandoraExtraDataWithSig := new(types.PanExtraDataWithBLSSig)
	err := rlp.DecodeBytes(pandoraHeaderData.Extra, pandoraExtraDataWithSig)
	if nil != err {
		log.WithField("error", err).
			Error("error converting extra data to extraDataWithSig")
		return false
	}

	// match signature
	if !reflect.DeepEqual(pandoraExtraDataWithSig.BlsSignatureBytes.Bytes(), vanguardShardInfo.GetSignature()) {
		log.WithField("pandora data signature", pandoraExtraDataWithSig.BlsSignatureBytes.Bytes()).
			WithField("vanguard signature", vanguardShardInfo.GetSignature()).
			Error("signature mismatched")
		return false
	}

	return true
}
