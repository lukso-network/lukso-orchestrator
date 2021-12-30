package consensus

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	eth2Types "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
)

func compareShardingInfo(ph *eth1Types.Header, vs *eth2Types.PandoraShard) (status bool) {
	defer func() {
		if !status {
			TotalMisMatchedShardCnt.Inc()
		}
	}()

	if ph == nil || vs == nil {
		return false
	}

	if vs.BlockNumber != ph.Number.Uint64() {
		log.WithField("panBlkNum", ph.Number.Uint64()).
			WithField("shardBlkNum", vs.BlockNumber).
			Error("Block number mismatched")
		return false
	}

	// match header hash
	if ph.Hash() != common.BytesToHash(vs.GetHash()) {
		log.WithField("panHeaderHash", ph.Hash()).
			WithField("shardHeaderHash", hexutil.Encode(vs.GetHash())).
			Error("Header hash mismatched")
		return false
	}

	// match parent hash
	if ph.ParentHash != common.BytesToHash(vs.GetParentHash()) {
		log.WithField("panParentHash", ph.ParentHash).
			WithField("shardParentHash", hexutil.Encode(vs.ParentHash)).
			Error("Parent hash mismatched")
		return false
	}

	// match state root hash
	if ph.Root != common.BytesToHash(vs.GetStateRoot()) {
		log.WithField("panStateRoot", ph.Root).
			WithField("shardStateRoot", hexutil.Encode(vs.StateRoot)).
			Error("State root hash mismatched")
		return false
	}

	// match TxHash
	if ph.TxHash != common.BytesToHash(vs.GetTxHash()) {
		log.WithField("panTxRoot", ph.TxHash).
			WithField("shardTxRoot", hexutil.Encode(vs.TxHash)).
			Error("Tx hash mismatched")
		return false
	}

	// match receiptHash
	if ph.ReceiptHash != common.BytesToHash(vs.GetReceiptHash()) {
		log.WithField("panReceiptHash", ph.ReceiptHash).
			WithField("shardReceiptHash", hexutil.Encode(vs.ReceiptHash)).
			Error("Receipt hash mismatched")
		return false
	}

	// retrieve extra data
	pandoraExtraDataWithSig := new(types.PanExtraDataWithBLSSig)
	err := rlp.DecodeBytes(ph.Extra, pandoraExtraDataWithSig)
	if nil != err {
		log.WithError(err).
			Error("Failed to convert extra data to extraDataWithSig")
		return false
	}

	// match signature
	if pandoraExtraDataWithSig.BlsSignatureBytes != types.BytesToSig(vs.GetSignature()) {
		log.WithField("panBlsSig", hexutil.Encode(pandoraExtraDataWithSig.BlsSignatureBytes.Bytes())).
			WithField("shardBlsSig", hexutil.Encode(vs.GetSignature())).
			Error("Signature mismatched")
		return false
	}

	return true
}

// verifyParentShardInfo method checks parent hash of vanguard and pandora current block header
// Retrieves latest verified slot info and then checks the incoming vanguard and pandora blocks parent hash
func (s *Service) verifyShardInfo(
	latestShardInfo *types.MultiShardInfo,
	panHeader *eth1Types.Header,
	vanShardInfo *types.VanguardShardInfo,
	latestStepId uint64,
) bool {

	if latestStepId == 0 {
		return true
	}

	if latestShardInfo == nil || latestShardInfo.IsNil() {
		log.Debug("Nil latest shard info so verification of shard info is failed")
		return false
	}

	vParentHash := common.BytesToHash(vanShardInfo.ParentRoot[:])
	pParentHash := panHeader.ParentHash

	if latestShardInfo.GetVanSlotRoot() != vParentHash {
		log.WithField("lastVerifiedVanHash", latestShardInfo.SlotInfo.BlockRoot).WithField("curVanParentHash", vParentHash).
			Info("Invalid vanguard parent hash")
		return false
	}

	if latestShardInfo.GetPanShardRoot() != pParentHash || latestShardInfo.GetPanBlockNumber()+1 != panHeader.Number.Uint64() {
		log.WithField("lastPanHash", latestShardInfo.GetPanShardRoot()).WithField("curPanParentHash", pParentHash).
			WithField("latestPanBlockNum", latestShardInfo.GetPanBlockNumber()).WithField("curPanBlockNum", panHeader.Number).
			Info("Invalid pandora parent hash")
		return false
	}

	return true
}
