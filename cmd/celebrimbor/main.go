package main

import (
	"fmt"
	joonix "github.com/joonix/log"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/node"
	"github.com/lukso-network/lukso-orchestrator/shared/cmd"
	"github.com/lukso-network/lukso-orchestrator/shared/journald"
	"github.com/lukso-network/lukso-orchestrator/shared/logutil"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"os"
	"runtime"
	runtimeDebug "runtime/debug"
	"sync"
	"time"
)

// ANYBODY HAS THE BETTER NAME JUST GIVE PROPOSAL!

// This library is responsible to spin your lukso infrastructure (Pandora, Vanguard, Validator, Orchestrator)
// In Tolkien's stories, Celebrimbor is an elven-smith who was manipulated into forging the Rings of Power
// by the disguised villain Sauron. While Celebrimbor created a set of Three on his own,
// Sauron left for Mordor and forged the One Ring, a master ring to control all the others, in the fires of Mount Doom.
// https://en.wikipedia.org/wiki/Celebrimbor
// We want to spin also 3 libraries at once, and secretly rule them by orchestrator. It matches for me somehow

// This binary will also support only some of the possible networks.
// Make a pull request to attach your network.
// We are also very open to any improvements. Please make some issue or hackmd proposal to make it better.
// Join our lukso discord https://discord.gg/E2rJPP4 to ask some questions

var (
	appName         = "celebrimbor"
	operatingSystem string
	pandoraTag      string
	validatorTag    string
	vanguardTag     string
	orchestratorTag string
	log             = logrus.WithField("prefix", appName)

	pandoraRuntimeFlags   []string
	validatorRuntimeFlags []string
	vanguardRuntimeFlags  []string
)

func init() {
	allFlags := make([]cli.Flag, 0)
	allFlags = append(allFlags, pandoraFlags...)
	allFlags = append(allFlags, validatorFlags...)
	allFlags = append(allFlags, vanguardFlags...)
	allFlags = append(allFlags, appFlags...)

	appFlags = cmd.WrapFlags(allFlags)
}

func main() {
	app := cli.App{}
	app.Name = appName
	app.Usage = "Spins all lukso ecosystem components"
	app.Flags = appFlags
	app.Action = downloadAndRunBinaries

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

		// Pandora related parsing
		pandoraTag = ctx.String(pandoraTagFlag)
		pandoraRuntimeFlags = preparePandoraFlags(ctx)

		// Validator related parsing
		validatorTag = ctx.String(validatorTagFlag)
		validatorRuntimeFlags = prepareValidatorFlags(ctx)

		// Vanguard related parsing
		vanguardTag = ctx.String(vanguardTagFlag)
		vanguardRuntimeFlags = prepareVanguardFlags(ctx)

		return nil
	}

	defer func() {
		if x := recover(); x != nil {
			log.Errorf("Runtime panic: %v\n%v", x, string(runtimeDebug.Stack()))
			panic(x)
		}
	}()

	err := app.Run(os.Args)

	if nil != err {
		log.Error(err.Error())
	}
}

func downloadAndRunBinaries(ctx *cli.Context) (err error) {
	// Get os, then download all binaries into datadir matching desired system
	// After successful download run binary with desired arguments spin and connect them
	// Orchestrator can be run from-memory
	err = downloadGenesis(ctx)

	if nil != err {
		return
	}

	err = downloadPandora(ctx)

	if nil != err {
		return
	}

	err = downloadValidator(ctx)

	if nil != err {
		return
	}

	err = downloadVanguard(ctx)

	if nil != err {
		return
	}

	// TODO: Figure out what is the desired order
	err = startPandora(ctx)

	if nil != err {
		return
	}

	return startOrchestrator(ctx)
}

func downloadPandora(ctx *cli.Context) (err error) {
	log.WithField("dependencyTag", pandoraTag).Info("I am downloading pandora")
	pandoraDataDir := ctx.String(pandoraDatadirFlag)
	err = clientDependencies[pandoraDependencyName].Download(pandoraTag, pandoraDataDir)

	return
}

func downloadGenesis(ctx *cli.Context) (err error) {
	log.WithField("dependencyTag", pandoraTag).Info("I am downloading pandora genesis")
	pandoraDataDir := ctx.String(pandoraDatadirFlag)
	err = clientDependencies[pandoraGenesisDependencyName].Download(pandoraTag, pandoraDataDir)

	if nil != err {
		return
	}

	log.WithField("dependencyTag", vanguardTag).Info("I am downloading vanguard genesis")
	vanguardDataDir := ctx.String(vanguardDatadirFlag)
	err = clientDependencies[vanguardGenesisDependencyName].Download(vanguardTag, vanguardDataDir)

	return
}

func downloadVanguard(ctx *cli.Context) (err error) {
	log.WithField("dependencyTag", vanguardTag).Info("I am downloading vanguard")
	vanguardDataDir := ctx.String(vanguardDatadirFlag)
	err = clientDependencies[vanguardDependencyName].Download(vanguardTag, vanguardDataDir)

	return
}

func downloadValidator(ctx *cli.Context) (err error) {
	log.WithField("dependencyTag", validatorTag).Info("I am downloading validator")
	validatorDataDir := ctx.String(vanguardDatadirFlag)
	err = clientDependencies[validatorDependencyName].Download(vanguardTag, validatorDataDir)

	return
}

// startPandora will direct stdOut to log file
func startPandora(ctx *cli.Context) (err error) {
	log.WithField("dependencyTag", pandoraTag).Info("I am running genesis.json init")
	pandoraDataDir := ctx.String(pandoraDatadirFlag)
	pandoraGenesisArguments := []string{
		"init",
		clientDependencies[pandoraGenesisDependencyName].ResolveBinaryPath(pandoraTag, pandoraDataDir),
		"--datadir",
		pandoraDataDir,
	}

	err = clientDependencies[pandoraDependencyName].Run(pandoraTag, pandoraDataDir, pandoraGenesisArguments)

	if nil != err {
		return
	}

	log.WithField("dependencyTag", pandoraTag).Info("I am running execution engine")
	err = clientDependencies[pandoraDependencyName].Run(pandoraTag, pandoraDataDir, pandoraRuntimeFlags)

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)

	go func() {
		for {
			_, currentErr := os.Stat(cmd.DefaultPandoraRPCEndpoint)
			if nil != currentErr {
				log.Info("Pandora ipc is up")
				waitGroup.Done()

				return
			}

			if os.IsNotExist(currentErr) {
				time.Sleep(time.Millisecond * 50)

				return
			}

			panic(err)
		}
	}()

	waitGroup.Wait()

	return
}

func startVanguard(ctx *cli.Context) (err error) {
	//err = clientDependencies[pandoraDependencyName].Run(pandoraTag, pandoraDataDir, pandoraGenesisArguments)
	//
	//if nil != err {
	//	return
	//}

	return
}

func startOrchestrator(ctx *cli.Context) (err error) {
	verbosity := ctx.String(cmd.VerbosityFlag.Name)
	level, err := logrus.ParseLevel(verbosity)
	if err != nil {
		return err
	}
	logrus.SetLevel(level)

	log.WithField("pandoraFlags", pandoraRuntimeFlags).
		WithField("vanguardFlags", vanguardFlags).
		WithField("validatorFlags", validatorFlags).Info("\n I will try to run setup with this additional flags \n")

	orchestrator, err := node.New(ctx)
	if err != nil {
		return err
	}
	orchestrator.Start()
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
