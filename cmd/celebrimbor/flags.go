package main

import (
	"fmt"
	"github.com/lukso-network/lukso-orchestrator/shared/cmd"
	"github.com/urfave/cli/v2"
)

const (
	// Pandora related flag names
	pandoraTagFlag       = "pandora-tag"
	pandoraDatadirFlag   = "pandora-datadir"
	pandoraEthstatsFlag  = "pandora-ethstats"
	pandoraBootnodesFlag = "pandora-bootnodes"
	pandoraNetworkIDFlag = "pandora-networkid"
	pandoraHttpApiFlag   = "pandora-http-apis"
	pandoraWSApiFlag     = "pandora-ws-apis"
	pandoraWSPortFlag    = "pandora-websocket-port"
	pandoraEtherbaseFlag = "pandora-etherbase"
	pandoraNotifyFlag    = "pandora-notify"
	pandoraVerbosityFlag = "pandora-verbosity"
	pandoraHttpPortFlag  = "pandora-http-port"

	// Validator related flag names
	validatorTagFlag                 = "validator-tag"
	validatorVanguardRpcProviderFlag = "validator-vanguard-rpc"
	validatorChainConfigFileFlag     = "validator-chain-config"
	validatorVerbosityFlag           = "validator-verbosity"
	validatorTrustedPandoraFlag      = "validator-trusted-pandora"

	// Orchestrator related flag names

	// Vanguard related flag names
	//vanguardRpcProviderFlag = "vanguard-rpc"
)

var (
	appFlags     = cmd.CommonFlagSet
	pandoraFlags = []cli.Flag{
		&cli.StringFlag{
			Name:  pandoraTagFlag,
			Usage: "provide a tag of pandora you would like to run",
			Value: "p-v0.0.10-alpha-bloom-debug2",
		},
		&cli.StringFlag{
			Name:  pandoraDatadirFlag,
			Usage: "provide a path you would like to store your data",
			Value: "./pandora",
		},
		&cli.StringFlag{
			Name:  pandoraEthstatsFlag,
			Usage: "nickname:STATS_LOGIN_SECRET@PANDORA_STATS_HOST",
			Value: "",
		},
		&cli.StringFlag{
			Name:  pandoraBootnodesFlag,
			Usage: "Default value should be ok for test network. Otherwise provide Comma separated enode urls, see at https://geth.ethereum.org/docs/getting-started/private-net.",
			Value: "enode://967db4f56ad0a1a35e3d30632fa600565329a23aff50c9762181810166f3c15b078cca522f930d1a2747778893232336bffd1ea5d2ca60543f1801d4360ea63a@10.0.6.6:0?discport=30301",
		},
		&cli.StringFlag{
			Name:  pandoraNetworkIDFlag,
			Usage: "provide network id if must be different than default",
			Value: "4004181",
		},
		&cli.StringFlag{
			Name:  pandoraHttpApiFlag,
			Usage: "comma separated apis",
			Value: "eth,net",
		},
		&cli.StringFlag{
			Name:  pandoraHttpPortFlag,
			Usage: "port used in pandora http communication",
			Value: "8565",
		},
		&cli.StringFlag{
			Name:  pandoraWSApiFlag,
			Usage: "comma separated apis",
			Value: "eth,net",
		},
		&cli.StringFlag{
			Name:  pandoraWSPortFlag,
			Usage: "port for pandora ws api",
			Value: "8546",
		},
		&cli.StringFlag{
			Name:  pandoraEtherbaseFlag,
			Usage: "your ECDSA public key used to get rewards on pandora chain",
			// yes, If you wont set it up, I'll get rewards ;]
			Value: "0x59E3dADc83af3c127a2e29B12B0E86109Bb6d838",
		},
		&cli.StringFlag{
			Name:  pandoraNotifyFlag,
			Usage: "this flag is used to pandora engine to notify validator and orchestrator",
			Value: "ws://127.0.0.1:7878,http://127.0.0.1:7877",
		},
		&cli.StringFlag{
			Name:  pandoraVerbosityFlag,
			Usage: "this flag sets up verobosity for pandora",
			Value: "info",
		},
	}
	validatorFlags = []cli.Flag{
		&cli.StringFlag{
			Name:  validatorVanguardRpcProviderFlag,
			Usage: "provide url without prefix, example: localhost:4000",
			Value: "localhost:4000",
		},
		&cli.StringFlag{
			Name:  validatorChainConfigFileFlag,
			Usage: "path to chain config of vanguard and validator",
			// TODO: check if this can be done from url. As far as I understand it can.
			Value: "./chain-config.yml",
		},
		&cli.StringFlag{
			Name:  validatorVerbosityFlag,
			Usage: "provide verbosity of validator",
			Value: "info",
		},
		&cli.StringFlag{
			Name:  validatorTrustedPandoraFlag,
			Usage: "provide host:port for trusted pandora, default: http://127.0.0.1:8565",
			Value: "http://127.0.0.1:8565",
		},
		&cli.StringFlag{
			Name:  validatorTagFlag,
			Usage: "provide tag for validator binary. Release must be present in lukso namespace on github",
			Value: "v0.0.16-alpha",
		},
	}
)

func prepareValidatorFlags(ctx *cli.Context) (validatorArguments []string) {
	if !ctx.Bool(cmd.AcceptTOUFlag.Name) {
		log.Fatal("you must accept terms of use")
		ctx.Done()

		return
	}

	validatorArguments = append(validatorArguments, "--accept-terms-of-use")

	if ctx.IsSet(cmd.ForceClearDB.Name) {
		validatorArguments = append(validatorArguments, "--force-clear-db")
	}

	validatorArguments = append(validatorArguments, fmt.Sprintf(
		"--chain-config-file=%s",
		ctx.String(validatorChainConfigFileFlag),
	))
	validatorArguments = append(validatorArguments, fmt.Sprintf(
		"--verbosity %s",
		ctx.String(validatorVerbosityFlag),
	))
	validatorArguments = append(validatorArguments, fmt.Sprintf(
		"--pandora-http-provider=%s",
		ctx.String(validatorTrustedPandoraFlag),
	))

	return
}

func preparePandoraFlags(ctx *cli.Context) (pandoraArguments []string) {
	pandoraArguments = append(pandoraArguments, fmt.Sprintf("--datadir %s", ctx.String(pandoraDatadirFlag)))
	pandoraArguments = append(pandoraArguments, fmt.Sprintf("--ethstats %s", ctx.String(pandoraEthstatsFlag)))
	pandoraArguments = append(pandoraArguments, fmt.Sprintf("--bootnodes %s", ctx.String(pandoraBootnodesFlag)))
	pandoraArguments = append(pandoraArguments, fmt.Sprintf("--networkid %s", ctx.String(pandoraNetworkIDFlag)))

	// Http api
	// TODO: change to new --http, because -rpc is deprecated in pandora
	pandoraArguments = append(pandoraArguments, "--rpc")
	pandoraArguments = append(pandoraArguments, "--rpcaddr 0.0.0.0")
	pandoraArguments = append(pandoraArguments, fmt.Sprintf("--rpcapi %s", ctx.String(pandoraHttpApiFlag)))
	pandoraArguments = append(pandoraArguments, fmt.Sprintf("--rpcport %s", ctx.String(pandoraHttpPortFlag)))

	// Websocket
	pandoraArguments = append(pandoraArguments, "--ws")
	pandoraArguments = append(pandoraArguments, "--ws.addr 0.0.0.0")
	pandoraArguments = append(pandoraArguments, fmt.Sprintf("--ws.api %s", ctx.String(pandoraWSApiFlag)))
	pandoraArguments = append(pandoraArguments, fmt.Sprintf("--ws.port %s", ctx.String(pandoraWSPortFlag)))
	pandoraArguments = append(pandoraArguments, "--ws.origins '*'")

	// Miner
	pandoraArguments = append(pandoraArguments, fmt.Sprintf(
		"--miner.etherbase %s", ctx.String(pandoraEtherbaseFlag),
	))
	pandoraArguments = append(pandoraArguments, "--mine")
	pandoraArguments = append(pandoraArguments, fmt.Sprintf("--miner.notify %s", ctx.String(pandoraNotifyFlag)))

	// Verbosity
	pandoraArguments = append(pandoraArguments, fmt.Sprintf("--verbosity %s", ctx.String(pandoraVerbosityFlag)))

	return
}
