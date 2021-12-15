package node

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/ethereum/go-ethereum/common/math"
	ethRpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/cache"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/consensus"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db/kv"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/pandorachain"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/rpc"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/vanguardchain"
	"github.com/lukso-network/lukso-orchestrator/shared"
	"github.com/lukso-network/lukso-orchestrator/shared/cmd"
	"github.com/lukso-network/lukso-orchestrator/shared/fileutil"
	"github.com/lukso-network/lukso-orchestrator/shared/version"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// OrchestratorNode
type OrchestratorNode struct {
	// basic configuration
	cliCtx *cli.Context
	ctx    context.Context
	cancel context.CancelFunc

	// service storage
	services *shared.ServiceRegistry
	lock     sync.RWMutex
	stop     chan struct{} // Channel to wait for termination notifications.

	//kv database with cache
	db db.Database

	// lru caches
	pandoraPendingCache *cache.PandoraCache
	vanPendingCache     *cache.VanguardCache
}

// New creates a new node instance, sets up configuration options, and registers
// every required service to the node.
func New(cliCtx *cli.Context) (*OrchestratorNode, error) {
	registry := shared.NewServiceRegistry()
	ctx, cancel := context.WithCancel(cliCtx.Context)

	orchestrator := &OrchestratorNode{
		cliCtx:              cliCtx,
		ctx:                 ctx,
		cancel:              cancel,
		services:            registry,
		stop:                make(chan struct{}),
		pandoraPendingCache: cache.NewPandoraCache(math.MaxInt32, cliCtx.Uint64(cmd.VanguardGenesisTime.Name), cliCtx.Uint64(cmd.SecondsPerSlot.Name)),
		vanPendingCache:     cache.NewVanguardCache(math.MaxInt32, cliCtx.Uint64(cmd.VanguardGenesisTime.Name), cliCtx.Uint64(cmd.SecondsPerSlot.Name)),
	}

	if err := orchestrator.startDB(orchestrator.cliCtx); err != nil {
		return nil, err
	}

	// Reverting db to latest finalized slot
	finalizedSlot := orchestrator.db.FinalizedSlot()
	// Removing slot infos from verified slot info db
	stepId, err := orchestrator.db.GetStepIdBySlot(finalizedSlot)
	if err != nil {
		log.WithError(err).WithField("finalizedSlot", finalizedSlot).WithField("stepId", stepId).
			Error("Could not found step id from DB")
		return nil, err
	}

	if err := orchestrator.db.RemoveShardingInfos(stepId + 1); err != nil {
		log.WithError(err).Error("Failed to remove latest verified shard infos from db")
		return nil, err
	}

	if err := orchestrator.db.SaveLatestStepID(stepId); err != nil {
		log.WithError(err).Error("Failed to update latest step id into db")
		return nil, err
	}

	if err := orchestrator.registerVanguardChainService(cliCtx); err != nil {
		return nil, err
	}

	if err := orchestrator.registerPandoraChainService(cliCtx); err != nil {
		return nil, err
	}

	if err := orchestrator.registerConsensusService(cliCtx); err != nil {
		return nil, err
	}

	if err := orchestrator.registerRPCService(cliCtx); err != nil {
		return nil, err
	}

	return orchestrator, nil
}

// startDB initialize KV db and cache
func (o *OrchestratorNode) startDB(cliCtx *cli.Context) error {
	baseDir := cliCtx.String(cmd.DataDirFlag.Name)
	dbPath := filepath.Join(baseDir, kv.OrchestratorNodeDbDirName)
	clearDB := cliCtx.Bool(cmd.ClearDB.Name)
	forceClearDB := cliCtx.Bool(cmd.ForceClearDB.Name)

	log.WithField("database-path", dbPath).Info("Checking DB")

	d, err := db.NewDB(o.ctx, dbPath, &kv.Config{
		InitialMMapSize: cliCtx.Int(cmd.BoltMMapInitialSizeFlag.Name),
	})
	if err != nil {
		return err
	}

	clearDBConfirmed := false
	if clearDB && !forceClearDB {
		actionText := "This will delete your orchestrator database stored in your data directory. " +
			"Your database backups will not be removed - do you want to proceed? (Y/N)"
		deniedText := "Database will not be deleted. No changes have been made."
		clearDBConfirmed, err = cmd.ConfirmAction(actionText, deniedText)
		if err != nil {
			return err
		}
	}

	if clearDBConfirmed || forceClearDB {
		log.Warning("Removing database")
		if err := d.Close(); err != nil {
			return errors.Wrap(err, "could not close db prior to clearing")
		}
		if err := d.ClearDB(); err != nil {
			return errors.Wrap(err, "could not clear database")
		}
		d, err = db.NewDB(o.ctx, dbPath, &kv.Config{
			InitialMMapSize: cliCtx.Int(cmd.BoltMMapInitialSizeFlag.Name),
		})
		if err != nil {
			return errors.Wrap(err, "could not create new database")
		}
	}

	o.db = d
	return nil
}

// registerVanguardChainService
func (o *OrchestratorNode) registerVanguardChainService(cliCtx *cli.Context) error {
	vanguardGRPCUrl := cliCtx.String(cmd.VanguardGRPCEndpoint.Name)
	svc, err := vanguardchain.NewService(
		o.ctx,
		vanguardGRPCUrl,
		o.db,
	)
	if err != nil {
		return nil
	}
	log.WithField("vanguardGRPCUrl", vanguardGRPCUrl).Info("Registered vanguard chain service")
	return o.services.RegisterService(svc)
}

// registerPandoraChainService
func (o *OrchestratorNode) registerPandoraChainService(cliCtx *cli.Context) error {
	pandoraRPCUrl := cliCtx.String(cmd.PandoraRPCEndpoint.Name)
	dialRPCClient := func(endpoint string) (*ethRpc.Client, error) {
		rpcClient, err := ethRpc.Dial(endpoint)
		if err != nil {
			return nil, err
		}
		return rpcClient, nil
	}
	namespace := "eth"
	svc, err := pandorachain.NewService(o.ctx, pandoraRPCUrl, namespace, o.db, dialRPCClient)
	if err != nil {
		return nil
	}
	log.WithField("pandoraHttpUrl", pandoraRPCUrl).Info("Registered pandora chain service")
	return o.services.RegisterService(svc)
}

// registerConsensusService
func (o *OrchestratorNode) registerConsensusService(cliCtx *cli.Context) error {
	var vanguardShardFeed *vanguardchain.Service
	if err := o.services.FetchService(&vanguardShardFeed); err != nil {
		return err
	}

	var pandoraHeaderFeed *pandorachain.Service
	if err := o.services.FetchService(&pandoraHeaderFeed); err != nil {
		return err
	}

	svc := consensus.New(o.ctx, &consensus.Config{
		VerifiedShardInfoDB: o.db,
		PanHeaderCache:      o.pandoraPendingCache,
		VanShardCache:       o.vanPendingCache,
		VanguardShardFeed:   vanguardShardFeed,
		PandoraHeaderFeed:   pandoraHeaderFeed,
	})

	log.Info("Registered consensus service")
	return o.services.RegisterService(svc)
}

// register RPC server
func (o *OrchestratorNode) registerRPCService(cliCtx *cli.Context) error {
	var consensusInfoFeed *vanguardchain.Service
	if err := o.services.FetchService(&consensusInfoFeed); err != nil {
		return err
	}

	var verifiedSlotInfoFeed *consensus.Service
	if err := o.services.FetchService(&verifiedSlotInfoFeed); err != nil {
		return err
	}

	var ipcapiURL string
	if cliCtx.String(cmd.IPCPathFlag.Name) != "" {
		ipcFilePath := cliCtx.String(cmd.IPCPathFlag.Name)
		ipcapiURL = fileutil.IpcEndpoint(filepath.Join(ipcFilePath, cmd.DefaultIpcPath), "")

		log.WithField("ipcFilePath", ipcFilePath).WithField(
			"ipcPath", ipcapiURL).Info("ipc file path")
	}

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
		ConsensusInfoFeed: consensusInfoFeed,
		Db:                o.db,
		IPCPath:           ipcapiURL,
		HTTPEnable:        httpEnable,
		HTTPHost:          httpListenAddr,
		HTTPPort:          httpPort,
		WSEnable:          wsEnable,
		WSHost:            wsListenerAddr,
		WSPort:            wsPort,

		PandoraHeaderCache:   o.pandoraPendingCache,
		VangShardCache:       o.vanPendingCache,
		VerifiedSlotInfoFeed: verifiedSlotInfoFeed,
	})
	if err != nil {
		return nil
	}

	log.Info("Registered RPC service")
	return o.services.RegisterService(svc)
}

// Start the OrchestratorNode and kicks off every registered service.
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
		panic("Panic closing the orchestrator node")
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
	if err := b.db.Close(); err != nil {
		log.Errorf("Failed to close database: %v", err)
	}
	b.cancel()
	close(b.stop)
}
