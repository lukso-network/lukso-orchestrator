package cmd

import (
	"github.com/ethereum/go-ethereum/node"
	"github.com/urfave/cli/v2"
	"strings"
	"time"
)

var (
	Now = time.Now().Unix()

	// RPC settings
	IPCDisabledFlag = &cli.BoolFlag{
		Name:  "ipcdisable",
		Usage: "Disable the IPC-RPC server",
	}

	HTTPEnabledFlag = &cli.BoolFlag{
		Name:  "http",
		Usage: "Enable the HTTP-RPC server",
	}

	HTTPListenAddrFlag = &cli.StringFlag{
		Name:  "http.addr",
		Usage: "HTTP-RPC server listening interface",
		Value: node.DefaultHTTPHost,
	}

	HTTPPortFlag = &cli.IntFlag{
		Name:  "http.port",
		Usage: "HTTP-RPC server listening port",
		Value: node.DefaultHTTPPort,
	}

	HTTPCORSDomainFlag = &cli.StringFlag{
		Name:  "http.corsdomain",
		Usage: "Comma separated list of domains from which to accept cross origin requests (browser enforced)",
		Value: "",
	}

	HTTPVirtualHostsFlag = &cli.StringFlag{
		Name:  "http.vhosts",
		Usage: "Comma separated list of virtual hostnames from which to accept requests (server enforced). Accepts '*' wildcard.",
		Value: strings.Join(node.DefaultConfig.HTTPVirtualHosts, ","),
	}

	WSEnabledFlag = &cli.BoolFlag{
		Name:  "ws",
		Usage: "Enable the WS-RPC server",
	}

	WSListenAddrFlag = &cli.StringFlag{
		Name:  "ws.addr",
		Usage: "WS-RPC server listening interface",
		Value: node.DefaultWSHost,
	}

	WSPortFlag = &cli.IntFlag{
		Name:  "ws.port",
		Usage: "WS-RPC server listening port",
		Value: node.DefaultWSPort,
	}

	// HTTPWeb3ProviderFlag provides an HTTP access endpoint to an ETH 1.0 RPC.
	VanguardRPCEndpoint = &cli.StringFlag{
		Name:  "vanguard-rpc-endpoint",
		Usage: "Vanguard node RPC provider endpoint(Default: 127.0.0.1:4000",
		Value: "127.0.0.1:4000",
	}

	PandoraRPCEndpoint = &cli.StringFlag{
		Name:  "pandora-rpc-endpoint",
		Usage: "Pandora node RP provider endpoint(Default: http://127.0.0.1:8545",
		Value: "http://127.0.0.1:8545",
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
