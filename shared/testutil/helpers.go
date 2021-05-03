package testutil

import (
	"github.com/ethereum/go-ethereum/common"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"math/big"
	"strconv"
	"time"
)

func NewMinimalConsensusInfo(epoch uint64) *types.MinimalEpochConsensusInfo {
	validatorList := make([]string, 32)

	for idx := 0; idx < 32; idx++ {
		bs := []byte(strconv.Itoa(31415926))
		pubKey := common.Bytes2Hex(bs)
		validatorList[idx] = pubKey
	}

	var validatorList32 [32]string
	copy(validatorList32[:], validatorList)
	return &types.MinimalEpochConsensusInfo{
		Epoch:            epoch,
		ValidatorList:    validatorList32,
		EpochStartTime:   765544433,
		SlotTimeDuration: time.Duration(6),
	}
}

// NewPandoraHeaderHash
func NewPandoraHeaderHash(slot uint64, status types.Status) *types.PanHeaderHash {
	return &types.PanHeaderHash{
		HeaderHash: NewEth1Header(slot).Hash(),
		Status:     status,
	}
}

// NewEth1Header
func NewEth1Header(slot uint64) *eth1Types.Header {
	epoch := slot / 32
	extraData := types.ExtraData{
		Slot:          slot,
		Epoch:         epoch,
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
