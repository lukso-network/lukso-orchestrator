package node

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/epochextractor"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/rpc"
	"github.com/lukso-network/lukso-orchestrator/shared"
	"github.com/lukso-network/lukso-orchestrator/shared/cmd"
	"github.com/lukso-network/lukso-orchestrator/shared/version"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// OrchestratorNode
type OrchestratorNode struct {
	cliCtx   *cli.Context
	ctx      context.Context
	cancel   context.CancelFunc
	services *shared.ServiceRegistry
	lock     sync.RWMutex
	stop     chan struct{} // Channel to wait for termination notifications.
}

// New creates a new node instance, sets up configuration options, and registers
// every required service to the node.
func New(cliCtx *cli.Context) (*OrchestratorNode, error) {
	registry := shared.NewServiceRegistry()
	ctx, cancel := context.WithCancel(cliCtx.Context)
	orchestrator := &OrchestratorNode{
		cliCtx:   cliCtx,
		ctx:      ctx,
		cancel:   cancel,
		services: registry,
		stop:     make(chan struct{}),
	}
	//if err := orchestrator.registerEpochExtractor(cliCtx); err != nil {
	//	return nil, err
	//}

	if err := orchestrator.registerRPCService(cliCtx); err != nil {
		return nil, err
	}
	return orchestrator, nil
}

// registerEpochExtractor
func (o *OrchestratorNode) registerEpochExtractor(cliCtx *cli.Context) error {
	pandoraHttpUrl := cliCtx.String(cmd.PandoraRPCEndpoint.Name)
	vanguardHttpUrl := cliCtx.String(cmd.VanguardRPCEndpoint.Name)
	genesisTime := cliCtx.Uint64(cmd.GenesisTime.Name)

	log.WithField("pandoraHttpUrl", pandoraHttpUrl).WithField(
		"vanguardHttpUrl", vanguardHttpUrl).WithField("genesisTime", genesisTime).Debug("flag values")

	svc, err := epochextractor.NewService(o.ctx, pandoraHttpUrl, vanguardHttpUrl, genesisTime)
	if err != nil {
		return nil
	}
	return o.services.RegisterService(svc)
}

// register RPC server
func (o *OrchestratorNode) registerRPCService(cliCtx *cli.Context) error {
	log.Info("Registering rpc server")
	httpEnable := cliCtx.Bool(cmd.HTTPEnabledFlag.Name)
	httpListenAddr := cliCtx.String(cmd.HTTPListenAddrFlag.Name)
	httpPort := cliCtx.Int(cmd.HTTPPortFlag.Name)
	wsEnable := cliCtx.Bool(cmd.WSEnabledFlag.Name)
	wsListenerAddr := cliCtx.String(cmd.WSListenAddrFlag.Name)
	wsPort := cliCtx.Int(cmd.WSPortFlag.Name)

	log.WithField("httpEnable", httpEnable).WithField("httpListenAddr", httpListenAddr).WithField(
		"httpPort", httpPort).WithField("wsEnable", wsEnable).WithField(
		"wsListenerAddr", wsListenerAddr).WithField("wsPort", wsPort).Debug("rpc server configuration")

	svc, err := rpc.NewService(o.ctx, &rpc.Config{
		HTTPHost: httpListenAddr,
		HTTPPort: httpPort,
		WSHost:   wsListenerAddr,
		WSPort:   wsPort,
	})
	if err != nil {
		return nil
	}
	return o.services.RegisterService(svc)
}

// Start the BeaconNode and kicks off every registered service.
func (o *OrchestratorNode) Start() {
	o.lock.Lock()

	log.WithFields(logrus.Fields{
		"version": version.Version(),
	}).Info("Starting orchestrator node")

	o.services.StartAll()

	stop := o.stop
	o.lock.Unlock()

	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigc)
		<-sigc
		log.Info("Got interrupt, shutting down...")
		go o.Close()
		for i := 10; i > 0; i-- {
			<-sigc
			if i > 1 {
				log.WithField("times", i-1).Info("Already shutting down, interrupt more to panic")
			}
		}
		panic("Panic closing the beacon node")
	}()

	// Wait for stop channel to be closed.
	<-stop
}

// Close handles graceful shutdown of the system.
func (b *OrchestratorNode) Close() {
	b.lock.Lock()
	defer b.lock.Unlock()

	log.Info("Stopping orchestrator node")
	b.services.StopAll()
	b.cancel()
	close(b.stop)
}
