package cmd

import (
	"github.com/urfave/cli/v2"
)

// TODO(Atif)- Need to have more options for http and ws for security purpose
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

	HTTPVirtualHosts = &cli.StringSliceFlag{
		Name:  "http.vhosts",
		Usage: "Comma separated list of virtual hostnames from which to accept requests (server enforced). Accepts '*' wildcard.",
		Value: cli.NewStringSlice("localhost"),
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

	// VanguardGenesisTime specifies time when vanguard node has started
	VanguardGenesisTime = &cli.Uint64Flag{
		Name:  "genesis-time",
		Usage: "Genesis timestamp of vanguard node",
	}

	SecondsPerSlot = &cli.Uint64Flag{
		Name:  "seconds-per-slot",
		Usage: "Seconds per slot",
	}

	// DisableMonitoringFlag defines a flag to disable the metrics collection.
	DisableMonitoringFlag = &cli.BoolFlag{
		Name:  "disable-monitoring",
		Usage: "Disable monitoring service.",
	}

	// MonitoringHostFlag defines the host used to serve prometheus metrics.
	MonitoringHostFlag = &cli.StringFlag{
		Name:  "monitoring-host",
		Usage: "Host used for listening and responding metrics for prometheus.",
		Value: "127.0.0.1",
	}

	// MonitoringPortFlag defines the http port used to serve prometheus metrics.
	MonitoringPortFlag = &cli.IntFlag{
		Name:  "monitoring-port",
		Usage: "Port used to listening and respond metrics for prometheus.",
		Value: 9080,
	}

	// EnableBackupWebhookFlag for users to trigger db backups via an HTTP webhook.
	EnableBackupWebhookFlag = &cli.BoolFlag{
		Name:  "enable-db-backup-webhook",
		Usage: "Serve HTTP handler to initiate database backups. The handler is served on the monitoring port at path /db/backup.",
	}
	// BackupWebhookOutputDir to customize the output directory for db backups.
	BackupWebhookOutputDir = &cli.StringFlag{
		Name:  "db-backup-output-dir",
		Usage: "Output directory for db backups",
	}

	// EnableTracingFlag defines a flag to enable p2p message tracing.
	EnableTracingFlag = &cli.BoolFlag{
		Name:  "enable-tracing",
		Usage: "Enable request tracing.",
	}
	// TracingProcessNameFlag defines a flag to specify a process name.
	TracingProcessNameFlag = &cli.StringFlag{
		Name:  "tracing-process-name",
		Usage: "The name to apply to tracing tag \"process_name\"",
	}
	// TracingEndpointFlag flag defines the http endpoint for serving traces to Jaeger.
	TracingEndpointFlag = &cli.StringFlag{
		Name:  "tracing-endpoint",
		Usage: "Tracing endpoint defines where beacon chain traces are exposed to Jaeger.",
		Value: "http://127.0.0.1:14268/api/traces",
	}
	// TraceSampleFractionFlag defines a flag to indicate what fraction of p2p
	// messages are sampled for tracing.
	TraceSampleFractionFlag = &cli.Float64Flag{
		Name:  "trace-sample-fraction",
		Usage: "Indicate what fraction of p2p messages are sampled for tracing.",
		Value: 0.20,
	}
)
