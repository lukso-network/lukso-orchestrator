package cmd

import (
	"github.com/urfave/cli/v2"
	"time"
)

var (
	Now = time.Now().Unix()

	// DataDirFlag defines a path on disk.
	DataDirFlag = &cli.StringFlag{
		Name:  "datadir",
		Usage: "Data directory for storing metadata",
		Value: DefaultConfigDir(),
	}

	IPCPathFlag = &cli.StringFlag{
		Name:  "ipcpath",
		Usage: "Filename for IPC socket/pipe within the datadir (explicit paths escape it)",
	}

	HTTPEnabledFlag = &cli.BoolFlag{
		Name:  "http",
		Usage: "Enable the HTTP-RPC server",
	}

	HTTPListenAddrFlag = &cli.StringFlag{
		Name:  "http.addr",
		Usage: "HTTP-RPC server listening interface",
		Value: DefaultHTTPHost,
	}

	HTTPPortFlag = &cli.IntFlag{
		Name:  "http.port",
		Usage: "HTTP-RPC server listening port",
		Value: DefaultHTTPPort,
	}

	WSEnabledFlag = &cli.BoolFlag{
		Name:  "ws",
		Usage: "Enable the WS-RPC server",
	}

	WSListenAddrFlag = &cli.StringFlag{
		Name:  "ws.addr",
		Usage: "WS-RPC server listening interface",
		Value: DefaultWSHost,
	}

	WSPortFlag = &cli.IntFlag{
		Name:  "ws.port",
		Usage: "WS-RPC server listening port",
		Value: DefaultWSPort,
	}

	// HTTPWeb3ProviderFlag provides an HTTP access endpoint to an ETH 1.0 RPC.
	VanguardRPCEndpoint = &cli.StringFlag{
		Name:  "vanguard-rpc-endpoint",
		Usage: "Vanguard node RPC provider endpoint",
		Value: DefaultVanguardRPCEndpoint,
	}

	PandoraRPCEndpoint = &cli.StringFlag{
		Name:  "pandora-rpc-endpoint",
		Usage: "Pandora node RP provider endpoint",
		Value: DefaultPandoraRPCEndpoint,
	}

	GenesisTime = &cli.Uint64Flag{
		Name:  "genesis-time",
		Usage: "Genesis time of the network",
		Value: uint64(Now),
	}

	// VerbosityFlag defines the logrus configuration.
	VerbosityFlag = &cli.StringFlag{
		Name:  "verbosity",
		Usage: "Logging verbosity (trace, debug, info=default, warn, error, fatal, panic)",
		Value: "info",
	}
)
