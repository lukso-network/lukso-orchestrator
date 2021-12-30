package cache

import (
	"crypto/rand"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

// Returns a random hash
func randHash() common.Hash {
	var h common.Hash
	rand.Read(h[:])
	return h
}

func getTestHeader(blockNumber int64) *types.Header {
	return &types.Header{
		MixDigest:   randHash(),
		ReceiptHash: randHash(),
		TxHash:      randHash(),
		Nonce:       types.BlockNonce{},
		Extra:       []byte{},
		Bloom:       types.Bloom{},
		GasUsed:     0,
		Coinbase:    common.Address{},
		GasLimit:    0,
		UncleHash:   randHash(),
		Time:        1337,
		ParentHash:  randHash(),
		Root:        randHash(),
		Number:      big.NewInt(blockNumber),
		Difficulty:  big.NewInt(2),
	}
}
