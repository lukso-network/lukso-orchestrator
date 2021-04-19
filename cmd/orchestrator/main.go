package main

import (
	"github.com/lukso-network/lukso-orchestrator/orchestrator/node"
	"github.com/lukso-network/lukso-orchestrator/shared/cmd"
	"github.com/lukso-network/lukso-orchestrator/shared/version"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	runtimeDebug "runtime/debug"
)

var appFlags = []cli.Flag{
	cmd.VanguardRPCEndpoint,
	cmd.PandoraRPCEndpoint,
	cmd.GenesisTime,
	cmd.VerbosityFlag,
	cmd.IPCPathFlag,
	cmd.HTTPEnabledFlag,
	cmd.HTTPListenAddrFlag,
	cmd.HTTPPortFlag,
	cmd.WSEnabledFlag,
	cmd.WSListenAddrFlag,
	cmd.WSPortFlag,
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
