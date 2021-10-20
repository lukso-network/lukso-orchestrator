package client

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	types "github.com/prysmaticlabs/eth2-types"
	ethpb "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"time"
)

type VanguardClient interface {
	CanonicalHeadSlot() (types.Slot, error)
	StreamNewPendingBlocks(blockRoot []byte, fromSlot types.Slot) (ethpb.BeaconChain_StreamNewPendingBlocksClient, error)
	StreamMinimalConsensusInfo(epoch uint64) (stream ethpb.BeaconChain_StreamMinimalConsensusInfoClient, err error)
	Close()
}

// Assure that GRPCClient struct will implement VanguardClient interface
var _ VanguardClient = &GRPCClient{}

const (
	protocol = "unix"
)

// GRPCClient
type GRPCClient struct {
	ctx             context.Context
	c               *grpc.ClientConn
	dialOpts        []grpc.DialOption
	beaconClient    ethpb.BeaconChainClient
	validatorClient ethpb.BeaconNodeValidatorClient
}

// Dial connects a client to the given URL or socket path.
func Dial(ctx context.Context, rpcAddress string, grpcRetryDelay time.Duration,
	grpcRetries uint, maxCallRecvMsgSize int) (VanguardClient, error) {

	dialer := func(addr string, t time.Duration) (net.Conn, error) {
		return net.Dial(protocol, addr)
	}

	dialOpts := constructDialOptions(
		maxCallRecvMsgSize,
		"",
		grpcRetries,
		grpcRetryDelay,
		grpc.WithInsecure(),
		grpc.WithDialer(dialer),
	)
	if dialOpts == nil {
		return nil, nil
	}

	c, err := grpc.DialContext(ctx, rpcAddress, dialOpts...)
	if err != nil {
		log.Errorf("Could not connect to RPC endpoint or socket: %s, %v", rpcAddress, err)
		return nil, err
	}

	return &GRPCClient{
		ctx,
		c,
		dialOpts,
		ethpb.NewBeaconChainClient(c),
		ethpb.NewBeaconNodeValidatorClient(c),
	}, nil
}

// Close
func (ec *GRPCClient) Close() {
	ec.c.Close()
}

// CanonicalHeadSlot returns the slot of canonical block currently found in the
// beacon chain via RPC.
func (vanClient *GRPCClient) CanonicalHeadSlot() (types.Slot, error) {
	head, err := vanClient.beaconClient.GetChainHead(vanClient.ctx, &emptypb.Empty{})
	if err != nil {
		log.WithError(err).Warn("Failed to get canonical head")
		return types.Slot(0), err
	}
	return head.HeadSlot, nil
}

// StreamNewPendingBlocks
func (vanClient *GRPCClient) StreamNewPendingBlocks(blockRoot []byte, fromSlot types.Slot) (
	stream ethpb.BeaconChain_StreamNewPendingBlocksClient,
	err error,
) {
	stream, err = vanClient.beaconClient.StreamNewPendingBlocks(
		vanClient.ctx,
		&ethpb.StreamPendingBlocksRequest{BlockRoot: blockRoot, FromSlot: fromSlot},
	)
	if err != nil {
		log.WithError(err).Error("Failed to subscribe to StreamChainHead")
		return
	}
	log.WithField("fromSlot", fromSlot).
		WithField("blockRoot", hexutil.Encode(blockRoot)).
		Info("Successfully subscribed to chain header event")
	return
}

func (vanClient *GRPCClient) StreamMinimalConsensusInfo(epoch uint64) (
	stream ethpb.BeaconChain_StreamMinimalConsensusInfoClient,
	err error,
) {
	stream, err = vanClient.beaconClient.StreamMinimalConsensusInfo(
		vanClient.ctx,
		&ethpb.MinimalConsensusInfoRequest{FromEpoch: types.Epoch(epoch)},
	)
	if err != nil {
		log.WithError(err).Error("Failed to subscribe to StreamMinimalConsensusInfo")
		return
	}
	log.WithField("fromEpoch", epoch).Info("Successfully subscribed to StreamMinimalConsensusInfo event")
	return
}

// constructDialOptions constructs a list of grpc dial options
func constructDialOptions(
	maxCallRecvMsgSize int,
	withCert string,
	grpcRetries uint,
	grpcRetryDelay time.Duration,
	extraOpts ...grpc.DialOption,
) []grpc.DialOption {
	var transportSecurity grpc.DialOption
	if withCert != "" {
		creds, err := credentials.NewClientTLSFromFile(withCert, "")
		if err != nil {
			log.Errorf("Could not get valid credentials: %v", err)
			return nil
		}
		transportSecurity = grpc.WithTransportCredentials(creds)
	} else {
		transportSecurity = grpc.WithInsecure()
		log.Warn("You are using an insecure gRPC connection. If you are running your beacon node and " +
			"validator on the same machines, you can ignore this message. If you want to know " +
			"how to enable secure connections, see: https://docs.prylabs.network/docs/prysm-usage/secure-grpc")
	}

	if maxCallRecvMsgSize == 0 {
		maxCallRecvMsgSize = 10 * 5 << 20 // Default 50Mb
	}

	dialOpts := []grpc.DialOption{
		transportSecurity,
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxCallRecvMsgSize),
			grpc_retry.WithMax(grpcRetries),
			grpc_retry.WithBackoff(grpc_retry.BackoffLinear(grpcRetryDelay)),
		),
	}

	dialOpts = append(dialOpts, extraOpts...)
	return dialOpts
}
