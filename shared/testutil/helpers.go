package testutil

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	eth2Types "github.com/prysmaticlabs/eth2-types"
	ethpb "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
	"golang.org/x/crypto/sha3"
	"math/big"
	"time"
)

func NewMinimalConsensusInfo(epoch uint64) *types.MinimalEpochConsensusInfoV2 {
	validatorList := make([]string, 32)

	for idx := 0; idx < 32; idx++ {
		pubKey := make([]byte, 48)
		validatorList[idx] = hexutil.Encode(pubKey)
	}

	var validatorList32 [32]string
	copy(validatorList32[:], validatorList)
	return &types.MinimalEpochConsensusInfoV2{
		Epoch:            epoch,
		ValidatorList:    validatorList32[:],
		EpochStartTime:   765544433,
		SlotTimeDuration: time.Duration(6),
	}
}

// NewEth1Header
func NewEth1Header(slot uint64) *eth1Types.Header {
	blockNumber := int64(slot)
	epoch := slot / 32
	extraData := types.ExtraData{
		Slot:  slot,
		Epoch: epoch,
		// TODO: remove this, we do not have this information
		ProposerIndex: 786,
	}

	signatureBytes := []byte("df7284286281db4c0bea60b338a62ddfde0d34736ad2657f2bea159fc8c6675cd5bbb68373e9f3d4bba017a82ed0d9b9")
	var blsSignatureBytes types.BlsSignatureBytes
	copy(blsSignatureBytes[:], signatureBytes[:])
	extraDataWithSig := types.PanExtraDataWithBLSSig{
		extraData,
		blsSignatureBytes,
	}
	extraDataByte, _ := rlp.EncodeToBytes(extraDataWithSig)
	header := &eth1Types.Header{
		ParentHash:  eth1Types.EmptyRootHash,
		UncleHash:   eth1Types.EmptyUncleHash,
		Coinbase:    common.HexToAddress("8888f1f195afa192cfee860698584c030f4c9db1"),
		Root:        common.HexToHash("ef1552a40b7165c3cd773806b9e0c165b75356e0314bf0706f279c729f51e017"),
		TxHash:      eth1Types.EmptyRootHash,
		ReceiptHash: eth1Types.EmptyRootHash,
		Difficulty:  big.NewInt(131072),
		Number:      big.NewInt(blockNumber),
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

// NewBeaconBlock
func NewVanguardShardInfo(slot uint64, header *eth1Types.Header) *types.VanguardShardInfo {
	return &types.VanguardShardInfo{
		Slot:           slot,
		ShardInfo:      NewPandoraShard(header),
		BlockHash:      []byte("0xd2302fac5c5f370575a70bcbab9fdaeb8f7e892f381d648ce1f2ad07ad17f20e"),
		FinalizedEpoch: 0,
		FinalizedSlot:  slot,
	}
}

func NewPandoraShard(panHeader *eth1Types.Header) *ethpb.PandoraShard {
	return &ethpb.PandoraShard{
		BlockNumber: panHeader.Number.Uint64(),
		Hash:        panHeader.Hash().Bytes(),
		ParentHash:  panHeader.ParentHash.Bytes(),
		StateRoot:   panHeader.Root.Bytes(),
		TxHash:      panHeader.TxHash.Bytes(),
		ReceiptHash: panHeader.ReceiptHash.Bytes(),
		Signature:   []byte("df7284286281db4c0bea60b338a62ddfde0d34736ad2657f2bea159fc8c6675cd5bbb68373e9f3d4bba017a82ed0d9b9"),
	}
}

// NewBeaconBlock creates a beacon block with minimum marshalable fields.
func NewBeaconBlock(slot uint64) *ethpb.BeaconBlock {
	return &ethpb.BeaconBlock{
		ParentRoot: make([]byte, 32),
		StateRoot:  make([]byte, 32),
		Slot:       eth2Types.Slot(slot),
		Body: &ethpb.BeaconBlockBody{
			RandaoReveal: make([]byte, 96),
			Eth1Data: &ethpb.Eth1Data{
				DepositRoot: make([]byte, 32),
				BlockHash:   make([]byte, 32),
			},
			Graffiti:          make([]byte, 32),
			Attestations:      []*ethpb.Attestation{},
			AttesterSlashings: []*ethpb.AttesterSlashing{},
			Deposits:          []*ethpb.Deposit{},
			ProposerSlashings: []*ethpb.ProposerSlashing{},
			VoluntaryExits:    []*ethpb.SignedVoluntaryExit{},
			PandoraShard:      []*ethpb.PandoraShard{NewPandoraShard(NewEth1Header(slot))},
		},
	}
}

func NewReOrg(slot uint64) *types.Reorg {
	return &types.Reorg{
		VanParentHash: []byte("0xd2302fac5c5f370575a70bcbab9fdaeb8f7e892f381d648ce1f2ad07ad17f20e"),
		PanParentHash: []byte("0xd2302fac5c5f370575a70bcbab9fdaeb8f7e892f381d648ce1f2ad07ad17f20e"),
		NewSlot:       slot,
	}
}
