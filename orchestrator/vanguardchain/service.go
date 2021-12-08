package vanguardchain

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/event"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/cache"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	ethpb "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/emptypb"
)

// time to wait before trying to reconnect with the vanguard node.
var (
	reConPeriod = 2 * time.Second
	errDialNil  = errors.New("failed to construct dial options")
)

// Service
// 	- maintains connection with vanguard chain
//	- handles vanguard subscription for consensus info.
//  - sends new consensus info to all pandora subscribers.
//  - maintains consensusInfoDB to store the coming consensus info from vanguard.
type Service struct {
	// service maintenance related attributes
	isRunning      bool
	processingLock sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	runError       error

	// vanguard chain related attributes
	connectedVanguard bool
	vanGRPCEndpoint   string
	dialOpts          []grpc.DialOption
	beaconClient      ethpb.BeaconChainClient
	nodeClient        ethpb.NodeClient
	conn              *grpc.ClientConn

	// subscription
	consensusInfoFeed        event.Feed
	scope                    event.SubscriptionScope
	vanguardShardingInfoFeed event.Feed
	subscriptionShutdownFeed event.Feed

	verifiedShardInfoDB db.ROnlyVerifiedShardInfoDB // db support
	consensusInfoDB     db.ConsensusInfoAccessDB    // db support
	shardingInfoCache   cache.VanguardShardCache    // lru cache support
	stopPendingBlkSubCh chan struct{}
	stopEpochInfoSubCh  chan struct{}
	reorgInfo           *types.Reorg
}

// NewService creates new service with vanguard endpoint, vanguard namespace and consensusInfoDB
func NewService(
	ctx context.Context,
	vanGRPCEndpoint string,
	db db.Database,
	cache cache.VanguardShardCache,
) (*Service, error) {

	ctx, cancel := context.WithCancel(ctx)
	_ = cancel // govet fix for lost cancel. Cancel is handled in service.Stop()

	return &Service{
		ctx:                 ctx,
		cancel:              cancel,
		vanGRPCEndpoint:     vanGRPCEndpoint,
		verifiedShardInfoDB: db,
		consensusInfoDB:     db,
		shardingInfoCache:   cache,
		stopPendingBlkSubCh: make(chan struct{}),
		stopEpochInfoSubCh:  make(chan struct{}),
	}, nil
}

// Start a consensus info fetcher service's main event loop.
func (s *Service) Start() {
	// Exit early if endpoint is not set.
	if s.vanGRPCEndpoint == "" {
		log.Error("Missing vanguard node's endpoint")
		return
	}

	if err := s.dialConn(); err != nil {
		log.WithError(err).Error("Could not create connection with vanguard node")
		return
	}

	go s.run()
}

func (s *Service) Stop() error {
	if s.cancel != nil {
		defer s.cancel()
	}
	s.scope.Close()
	if s.conn != nil {
		s.conn.Close()
	}
	return nil
}

func (s *Service) Status() error {
	// Service don't start
	if !s.isRunning {
		return nil
	}
	// get error from run function
	if s.runError != nil {
		return s.runError
	}
	return nil
}

// run subscribes to all the services for the ETH1.0 chain.
func (s *Service) run() {

	s.waitForConnection()

	latestFinalizedEpoch := s.verifiedShardInfoDB.FinalizedEpoch()
	latestFinalizedSlot := s.verifiedShardInfoDB.FinalizedSlot()
	fromEpoch := latestFinalizedEpoch

	// checking consensus info db
	for i := latestFinalizedEpoch; i >= 0; {
		epochInfo, _ := s.consensusInfoDB.ConsensusInfo(s.ctx, i)
		if epochInfo == nil {
			// epoch info is missing. so subscribe from here. maybe db operation was wrong
			fromEpoch = i
			log.WithField("epoch", fromEpoch).Debug("Found missing epoch info in db, so subscription should " +
				"be started from this missing epoch")
		}
		if i == 0 {
			break
		}
		i--
	}

	go s.subscribeNewConsensusInfoGRPC(s.ctx, fromEpoch)
	go s.subscribeVanNewPendingBlockHash(s.ctx, latestFinalizedSlot)
}

// waitForConnection waits for a connection with vanguard chain. Until a successful with
// vanguard chain, it retries again and again.
func (s *Service) waitForConnection() {
	if err := s.dialConn(); err != nil {
		log.WithError(err).Error("Could not create connection with vanguard node during re-subscription")
		return
	}

	if _, err := s.beaconClient.GetChainHead(s.ctx, &emptypb.Empty{}); err == nil {
		log.WithField("vanguardEndpoint", s.vanGRPCEndpoint).Info("Connected vanguard chain")
		s.connectedVanguard = true
		return
	}

	ticker := time.NewTicker(reConPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if _, err := s.beaconClient.GetChainHead(s.ctx, &emptypb.Empty{}); err != nil {
				log.WithField("vanguardEndpoint", s.vanGRPCEndpoint).Warn("Could not connect or subscribe to vanguard chain")
				continue
			}
			s.connectedVanguard = true
			s.runError = nil
			log.WithField("vanguardEndpoint", s.vanGRPCEndpoint).Info("Connected vanguard chain")
			return
		case <-s.ctx.Done():
			log.Info("Received cancelled context, closing existing go routine: waitForConnection")
			return
		}
	}
}

// SubscribeMinConsensusInfoEvent registers a subscription of ChainHeadEvent.
func (s *Service) SubscribeMinConsensusInfoEvent(ch chan<- *types.MinimalEpochConsensusInfoV2) event.Subscription {
	return s.scope.Track(s.consensusInfoFeed.Subscribe(ch))
}

func (s *Service) SubscribeShardInfoEvent(ch chan<- *types.VanguardShardInfo) event.Subscription {
	return s.scope.Track(s.vanguardShardingInfoFeed.Subscribe(ch))
}

func (s *Service) SubscribeShutdownSignalEvent(ch chan<- *types.Reorg) event.Subscription {
	return s.scope.Track(s.subscriptionShutdownFeed.Subscribe(ch))
}

func (s *Service) StopSubscription(reorgInfo *types.Reorg) {
	defer s.processingLock.Unlock()
	s.processingLock.Lock()

	s.reorgInfo = reorgInfo
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
	log.Info("Stopped vanguard gRPC subscription due to reorg")
}

// dialConn method creates connection with vanguard grpc server
func (s *Service) dialConn() error {
	s.processingLock.Lock()
	defer s.processingLock.Unlock()

	if s.conn != nil {
		return nil
	}

	grpcAddress, protocol, err := resolveRpcAddressAndProtocol(s.vanGRPCEndpoint, "")
	if nil != err {
		return nil
	}

	dialOpts := constructDialOptions(math.MaxInt32, "", 32, time.Minute*6)
	if dialOpts == nil {
		return errDialNil
	}

	if "unix" == protocol {
		dialer := func(addr string, t time.Duration) (net.Conn, error) {
			return net.Dial(protocol, addr)
		}

		dialOpts = append(dialOpts, grpc.WithDialer(dialer))
	}

	c, err := grpc.DialContext(s.ctx, grpcAddress, dialOpts...)
	if err != nil {
		return err
	}

	s.conn = c
	s.beaconClient = ethpb.NewBeaconChainClient(c)
	s.nodeClient = ethpb.NewNodeClient(c)

	return nil
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
		log.Warn("You are using an insecure gRPC connection. If you are running your orchestrator, vanguard, pandora and " +
			"validator node on the same machines, you can ignore this message. If you want to know " +
			"how to enable secure connections, see: https://docs.lukso.tech/networks/l15-testnet")
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

// resolveRpcAddressAndProtocol returns a RPC address and protocol.
// It can be HTTP/S layer or IPC socket, tcp or unix socket.
func resolveRpcAddressAndProtocol(host, port string) (address string, protocol string, err error) {
	if port != "" {
		address = fmt.Sprintf("%s:%s", host, port)
	}

	if strings.Contains(address, ".ipc") || port == "" {
		address = host
	}

	u, err := url.Parse(address)
	if err != nil {
		host, _, err = net.SplitHostPort(address)
		if err != nil {
			return address, "", err
		}
	}
	if u != nil {
		host = u.Host
	}

	if net.ParseIP(host) != nil {
		return address, "tcp", nil
	}

	switch u.Scheme {
	case "http", "https":
		return address, "tcp", nil
	case "":
		if len(strings.TrimSpace(u.Path)) == 0 || !strings.Contains(address, ".ipc") {
			return address, "", fmt.Errorf("invalid socket path %q", address)
		}
		return address, "unix", nil
	default:
		return address, protocol, fmt.Errorf("no known network transport layer for address %q", address)
	}
}
