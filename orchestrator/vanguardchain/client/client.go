package client

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	types "github.com/prysmaticlabs/eth2-types"
	ethpb "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"net"
	"net/url"
	"strings"
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

	rpcAddress, protocol, err := prepareRpcAddressAndProtocol(rpcAddress)
	if nil != err {
		return nil, err
	}

	dialOpts := constructDialOptions(
		maxCallRecvMsgSize,
		"",
		grpcRetries,
		grpcRetryDelay,
		grpc.WithInsecure(),
	)
	if dialOpts == nil {
		return nil, nil
	}

	if "unix" == protocol {
		dialer := func(addr string, t time.Duration) (net.Conn, error) {
			return net.Dial(protocol, addr)
		}

		dialOpts = append(dialOpts, grpc.WithDialer(dialer))
	}

	c, err := grpc.DialContext(ctx, rpcAddress, dialOpts...)
	if err != nil {
		log.Errorf("Could not connect to gRPC endpoint or socket: %s, %v", rpcAddress, err)
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

// prepareRpcAddressAndProtocol returns a RPC address and protocol.
// It can be HTTP/S layer or IPC socket, tcp or unix socket.
func prepareRpcAddressAndProtocol(rpcAddress string) (string, string, error) {
	var host string
	u, err := url.Parse(rpcAddress)
	if err != nil {
		host, _, err = net.SplitHostPort(rpcAddress)
		if err != nil {
			return rpcAddress, "", err
		}
	}
	if u != nil {
		host = u.Host
	}

	if net.ParseIP(host) != nil {
		return rpcAddress, "tcp", nil
	}

	switch u.Scheme {
	case "http", "https":
		return rpcAddress, "tcp", nil
	case "":
		if len(strings.TrimSpace(u.Path)) == 0 {
			return rpcAddress, "", fmt.Errorf("invalid socket path %q", u.Path)
		}
		return rpcAddress, "unix", nil
	default:
		return rpcAddress, "", fmt.Errorf("no known transport for URL scheme %q", u.Scheme)
	}
}
