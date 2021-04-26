package main

import (
	"fmt"
	joonix "github.com/joonix/log"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/node"
	"github.com/lukso-network/lukso-orchestrator/shared/cmd"
	"github.com/lukso-network/lukso-orchestrator/shared/journald"
	"github.com/lukso-network/lukso-orchestrator/shared/logutil"
	"github.com/lukso-network/lukso-orchestrator/shared/version"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"os"
	"runtime"
	runtimeDebug "runtime/debug"
)

var appFlags = []cli.Flag{
	cmd.VanguardRPCEndpoint,
	cmd.VerbosityFlag,
	cmd.IPCPathFlag,
	cmd.HTTPEnabledFlag,
	cmd.HTTPListenAddrFlag,
	cmd.HTTPPortFlag,
	cmd.WSEnabledFlag,
	cmd.WSListenAddrFlag,
	cmd.WSPortFlag,
	cmd.DataDirFlag,
	cmd.ClearDB,
	cmd.ForceClearDB,
	cmd.LogFileName,
	cmd.LogFormat,
}

func init() {
	appFlags = cmd.WrapFlags(append(appFlags))
}

func main() {
	app := cli.App{}
	app.Name = "orchestrator"
	app.Usage = "Orchestrator client orchestrates pandora and vanguard client"
	app.Action = startNode
	app.Version = version.Version()

	app.Flags = appFlags
	app.Before = func(ctx *cli.Context) error {
		format := ctx.String(cmd.LogFormat.Name)
		switch format {
		case "text":
			formatter := new(prefixed.TextFormatter)
			formatter.TimestampFormat = "2006-01-02 15:04:05"
			formatter.FullTimestamp = true
			// If persistent log files are written - we disable the log messages coloring because
			// the colors are ANSI codes and seen as gibberish in the log files.
			formatter.DisableColors = ctx.String(cmd.LogFileName.Name) != ""
			logrus.SetFormatter(formatter)
		case "fluentd":
			f := joonix.NewFormatter()
			if err := joonix.DisableTimestampFormat(f); err != nil {
				panic(err)
			}
			logrus.SetFormatter(f)
		case "json":
			logrus.SetFormatter(&logrus.JSONFormatter{})
		case "journald":
			if err := journald.Enable(); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown log format %s", format)
		}

		logFileName := ctx.String(cmd.LogFileName.Name)
		if logFileName != "" {
			if err := logutil.ConfigurePersistentLogging(logFileName); err != nil {
				log.WithError(err).Error("Failed to configuring logging to disk.")
			}
		}

		runtime.GOMAXPROCS(runtime.NumCPU())
		return nil
	}

	defer func() {
		if x := recover(); x != nil {
			log.Errorf("Runtime panic: %v\n%v", x, string(runtimeDebug.Stack()))
			panic(x)
		}
	}()

	if err := app.Run(os.Args); err != nil {
		log.Error(err.Error())
	}
}

// startNode
func startNode(ctx *cli.Context) error {
	verbosity := ctx.String(cmd.VerbosityFlag.Name)
	level, err := logrus.ParseLevel(verbosity)
	if err != nil {
		return err
	}
	logrus.SetLevel(level)

	orchestrator, err := node.New(ctx)
	if err != nil {
		return err
	}
	orchestrator.Start()
	return nil
}
