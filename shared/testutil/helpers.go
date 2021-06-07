package testutil

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"
	"math/big"
	"time"
)

func NewMinimalConsensusInfo(epoch uint64) *types.MinimalEpochConsensusInfo {
	validatorList := make([]string, 32)

	for idx := 0; idx < 32; idx++ {
		pubKey := make([]byte, 48)
		validatorList[idx] = hexutil.Encode(pubKey)
	}

	var validatorList32 [32]string
	copy(validatorList32[:], validatorList)
	return &types.MinimalEpochConsensusInfo{
		Epoch:            epoch,
		ValidatorList:    validatorList32[:],
		EpochStartTime:   765544433,
		SlotTimeDuration: time.Duration(6),
	}
}

// NewPandoraHeaderHash
func NewPandoraHeaderHash(slot uint64, status types.Status) *types.HeaderHash {
	return &types.HeaderHash{
		HeaderHash: NewEth1Header(slot).Hash(),
		Status:     status,
	}
}

// NewEth1Header
func NewEth1Header(slot uint64) *eth1Types.Header {
	epoch := slot / 32
	extraData := types.ExtraData{
		Slot:  slot,
		Epoch: epoch,
		// TODO: remove this, we do not have this information
		ProposerIndex: 786,
	}
	extraDataByte, _ := rlp.EncodeToBytes(extraData)
	header := &eth1Types.Header{
		ParentHash:  eth1Types.EmptyRootHash,
		UncleHash:   eth1Types.EmptyUncleHash,
		Coinbase:    common.HexToAddress("8888f1f195afa192cfee860698584c030f4c9db1"),
		Root:        common.HexToHash("ef1552a40b7165c3cd773806b9e0c165b75356e0314bf0706f279c729f51e017"),
		TxHash:      eth1Types.EmptyRootHash,
		ReceiptHash: eth1Types.EmptyRootHash,
		Difficulty:  big.NewInt(131072),
		Number:      big.NewInt(314),
		GasLimit:    uint64(3141592),
		GasUsed:     uint64(21000),
		Time:        uint64(1426516743),
		Extra:       extraDataByte,
		MixDigest:   eth1Types.EmptyRootHash,
		Nonce:       eth1Types.BlockNonce{0x01, 0x02, 0x03},
	}
	return header
}

// SealHash returns the hash of a block prior to it being sealed.
func SealHash(header *eth1Types.Header) (hash common.Hash) {
	hasher := sha3.NewLegacyKeccak256()

	if err := rlp.Encode(hasher, []interface{}{
		header.ParentHash,
		header.UncleHash,
		header.Coinbase,
		header.Root,
		header.TxHash,
		header.ReceiptHash,
		header.Bloom,
		header.Difficulty,
		header.Number,
		header.GasLimit,
		header.GasUsed,
		header.Time,
		header.Extra,
	}); err != nil {
		return eth1Types.EmptyRootHash
	}
	hasher.Sum(hash[:0])
	return hash
}

// GenerateExtraDataWithBLSSig generates pandora extra data with header hash signature
func GenerateExtraDataWithBLSSig(header *eth1Types.Header) (*types.PanExtraDataWithBLSSig, error) {
	extraData := new(types.ExtraData)
	if err := rlp.DecodeBytes(header.Extra, extraData); err != nil {
		return nil, errors.Wrap(err, "Failed to decode extra data fields")
	}
	var blsSignatureBytes types.BlsSignatureBytes
	signatureBytes := make([]byte, types.BLSSignatureSize)
	copy(blsSignatureBytes[:], signatureBytes[:])
	extraDataWithBlsSig := new(types.PanExtraDataWithBLSSig)
	extraDataWithBlsSig.ExtraData = *extraData
	extraDataWithBlsSig.BlsSignatureBytes = &blsSignatureBytes
	return extraDataWithBlsSig, nil
}
