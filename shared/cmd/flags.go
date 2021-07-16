package cmd

import (
	"github.com/urfave/cli/v2"
)

var (
	// DataDirFlag defines a path on disk.
	DataDirFlag = &cli.StringFlag{
		Name:  "datadir",
		Usage: "Data directory for storing consensus metadata and block headers",
		Value: DefaultConfigDir(),
	}

	// ForceClearDB removes any previously stored data at the data directory.
	ForceClearDB = &cli.BoolFlag{
		Name:  "force-clear-db",
		Usage: "Clear any previously stored data at the data directory",
	}
	// ClearDB prompts user to see if they want to remove any previously stored data at the data directory.
	ClearDB = &cli.BoolFlag{
		Name:  "clear-db",
		Usage: "Prompt for clearing any previously stored data at the data directory",
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

	VanguardGRPCEndpoint = &cli.StringFlag{
		Name:  "vanguard-grpc-endpoint",
		Usage: "Vanguard node gRPC provider endpoint",
		Value: DefaultVanguardGRPCEndpoint,
	}

	// PandoraRPCEndpoint provides an WSS/IPC access endpoint to an Pandora RPC.
	PandoraRPCEndpoint = &cli.StringFlag{
		Name:  "pandora-rpc-endpoint",
		Usage: "Pandora node RPC provider endpoint",
		Value: DefaultPandoraRPCEndpoint,
	}

	// VerbosityFlag defines the logrus configuration.
	VerbosityFlag = &cli.StringFlag{
		Name:  "verbosity",
		Usage: "Logging verbosity (trace, debug, info=default, warn, error, fatal, panic)",
		Value: "info",
	}

	// BoltMMapInitialSizeFlag specifies the initial size in bytes of boltdb's mmap syscall.
	BoltMMapInitialSizeFlag = &cli.IntFlag{
		Name:  "bolt-mmap-initial-size",
		Usage: "Specifies the size in bytes of bolt db's mmap syscall allocation",
		Value: 536870912, // 512 Mb as a default value.
	}

	// LogFormat specifies the log output format.
	LogFormat = &cli.StringFlag{
		Name:  "log-format",
		Usage: "Specify log formatting. Supports: text, json, fluentd, journald.",
		Value: "text",
	}

	// LogFileName specifies the log output file name.
	LogFileName = &cli.StringFlag{
		Name:  "log-file",
		Usage: "Specify log file name, relative or absolute",
	}

	CommonFlagSet = []cli.Flag{
		VanguardGRPCEndpoint,
		PandoraRPCEndpoint,
		VerbosityFlag,
		IPCPathFlag,
		HTTPEnabledFlag,
		HTTPListenAddrFlag,
		HTTPPortFlag,
		WSEnabledFlag,
		WSListenAddrFlag,
		WSPortFlag,
		DataDirFlag,
		ClearDB,
		ForceClearDB,
		LogFileName,
		LogFormat,
	}
)
