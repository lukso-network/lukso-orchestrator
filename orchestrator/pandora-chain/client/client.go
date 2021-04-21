package client

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	types "github.com/prysmaticlabs/eth2-types"
	"math/big"
)

type PandoraClient interface {
	ChainID(ctx context.Context) (*big.Int, error)
	InsertConsensusInfo(ctx context.Context, curEpoch types.Epoch, validatorPubKeys []string, epochStartTime uint64) (bool, error)
	Close()
}

// Client defines typed wrappers for the Ethereum RPC API.
type RPCClient struct {
	c *rpc.Client
}

// Dial connects a client to the given URL.
func Dial(ctx context.Context, rawurl string) (*RPCClient, error) {
	return DialContext(ctx, rawurl)
}

func DialContext(ctx context.Context, rawurl string) (*RPCClient, error) {
	c, err := rpc.DialContext(ctx, rawurl)
	if err != nil {
		return nil, err
	}
	return NewClient(c), nil
}

// NewClient creates a client that uses the given RPC client.
func NewClient(c *rpc.Client) *RPCClient {
	return &RPCClient{c}
}

func (ec *RPCClient) Close() {
	ec.c.Close()
}

// ChainId retrieves the current chain ID for transaction replay protection.
func (ec *RPCClient) ChainID(ctx context.Context) (*big.Int, error) {
	var result hexutil.Big
	err := ec.c.CallContext(ctx, &result, "eth_chainId")
	if err != nil {
		return nil, err
	}
	return (*big.Int)(&result), err
}

//
func (ec *RPCClient) InsertConsensusInfo(ctx context.Context,
	curEpoch types.Epoch, validatorPubKeys []string, epochStartTime uint64) (bool, error) {

	var status bool
	if err := ec.c.Call(
		&status,
		"eth_insertMinimalConsensusInfo",
		uint64(curEpoch),
		validatorPubKeys,
		epochStartTime,
	); err != nil {
		return false, err
	}

	return status, nil
}
