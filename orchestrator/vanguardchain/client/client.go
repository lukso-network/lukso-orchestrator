package client

import (
	"context"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	types "github.com/prysmaticlabs/eth2-types"
	ethpb "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"time"
)

type VanguardClient interface {
	ChainHead() (*ethpb.ChainHead, error)
	StreamNewPendingBlocks(blockRoot []byte, fromSlot types.Slot) (ethpb.BeaconChain_StreamNewPendingBlocksClient, error)
	StreamMinimalConsensusInfo(epoch uint64) (stream ethpb.BeaconChain_StreamMinimalConsensusInfoClient, err error)
	Close()
	SyncStatus() (bool, error)
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
	nodeClient      ethpb.NodeClient
}

// Dial connects a client to the given URL.
func Dial(ctx context.Context, rawurl string, grpcRetryDelay time.Duration,
	grpcRetries uint, maxCallRecvMsgSize int) (VanguardClient, error) {

	dialOpts := constructDialOptions(
		maxCallRecvMsgSize,
		"",
		grpcRetries,
		grpcRetryDelay,
	)
	if dialOpts == nil {
		return nil, nil
	}

	c, err := grpc.DialContext(ctx, rawurl, dialOpts...)
	if err != nil {
		log.Errorf("Could not dial endpoint: %s, %v", rawurl, err)
		return nil, err
	}

	return &GRPCClient{
		ctx,
		c,
		dialOpts,
		ethpb.NewBeaconChainClient(c),
		ethpb.NewBeaconNodeValidatorClient(c),
		ethpb.NewNodeClient(c),
	}, nil
}

// Close
func (vanClient *GRPCClient) Close() {
	vanClient.c.Close()
}

// CanonicalHeadSlot returns the slot of canonical block currently found in the
// beacon chain via RPC.
func (vanClient *GRPCClient) ChainHead() (*ethpb.ChainHead, error) {
	head, err := vanClient.beaconClient.GetChainHead(vanClient.ctx, &emptypb.Empty{})
	if err != nil {
		log.WithError(err).Warn("Failed to get canonical chain head")
		return nil, err
	}
	return head, nil
}

// SyncStatus
func (vanClient *GRPCClient) SyncStatus() (bool, error) {
	status, err := vanClient.nodeClient.GetSyncStatus(vanClient.ctx, &emptypb.Empty{})
	if err != nil {
		log.WithError(err).Error("Could not fetch sync status")
		return false, err
	}
	if status != nil && !status.Syncing {
		log.Info("Beacon node is fully synced")
		return true, nil
	}
	return false, nil
}

// StreamNewPendingBlocks
func (vanClient *GRPCClient) StreamNewPendingBlocks(blockRoot []byte, fromSlot types.Slot) (
	ethpb.BeaconChain_StreamNewPendingBlocksClient,
	error,
) {
	stream, err := vanClient.beaconClient.StreamNewPendingBlocks(
		vanClient.ctx,
		&ethpb.StreamPendingBlocksRequest{BlockRoot: blockRoot, FromSlot: fromSlot},
	)
	if err != nil {
		log.WithError(err).Error("Failed to subscribe to pending vanguard blocks event")
		return nil, err
	}
	return stream, nil
}

// StreamMinimalConsensusInfo
func (vanClient *GRPCClient) StreamMinimalConsensusInfo(epoch uint64) (
	ethpb.BeaconChain_StreamMinimalConsensusInfoClient,
	error,
) {
	stream, err := vanClient.beaconClient.StreamMinimalConsensusInfo(
		vanClient.ctx,
		&ethpb.MinimalConsensusInfoRequest{FromEpoch: types.Epoch(epoch)},
	)
	if err != nil {
		log.WithError(err).Error("Failed to subscribe to epoch info event")
		return nil, err
	}
	log.WithField("fromEpoch", epoch).Info("Successfully subscribed to minimal consensus info event")
	return stream, nil
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
