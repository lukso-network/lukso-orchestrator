package cmd

import (
	"github.com/urfave/cli/v2"
	"time"
)

var (
	Now = time.Now().Unix()

	// HTTPWeb3ProviderFlag provides an HTTP access endpoint to an ETH 1.0 RPC.
	VanguardRPCEndpoint = &cli.StringFlag{
		Name:  "vanguard-rpc-endpoint",
		Usage: "Vanguard node RPC provider endpoint(Default: 127.0.0.1:4000",
		Value: "127.0.0.1:4000",
	}

	PandoraRPCEndpoint = &cli.StringFlag{
		Name: "pandora-rpc-endpoint",
		Usage: "Pandora node RP provider endpoint(Default: http://127.0.0.1:8545",
		Value: "http://127.0.0.1:8545",
	}

	GenesisTime = &cli.Uint64Flag {
		Name: "genesis-time",
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
