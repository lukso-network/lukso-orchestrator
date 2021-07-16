package main

import (
	"fmt"
	"github.com/lukso-network/lukso-orchestrator/shared/cmd"
	"github.com/urfave/cli/v2"
)

const (
	// Pandora related flag names
	pandoraTagFlag         = "pandora-tag"
	pandoraDatadirFlag     = "pandora-datadir"
	pandoraEthstatsFlag    = "pandora-ethstats"
	pandoraBootnodesFlag   = "pandora-bootnodes"
	pandoraNetworkIDFlag   = "pandora-networkid"
	pandoraPortFlag        = "pandora-port"
	pandoraChainIDFlag     = "pandora-chainid"
	pandoraHttpApiFlag     = "pandora-http-apis"
	pandoraWSApiFlag       = "pandora-ws-apis"
	pandoraWSPortFlag      = "pandora-websocket-port"
	pandoraEtherbaseFlag   = "pandora-etherbase"
	pandoraGenesisFileFlag = "pandora-genesis"
	pandoraNotifyFlag      = "pandora-notify"
	pandoraVerbosityFlag   = "pandora-verbosity"
	pandoraHttpPortFlag    = "pandora-http-port"

	// Common for prysm client
	vanguardChainConfigFlag = "validator-chain-config"

	// Validator related flag names
	validatorTagFlag                 = "validator-tag"
	validatorVanguardRpcProviderFlag = "validator-vanguard-rpc"
	validatorVerbosityFlag           = "validator-verbosity"
	validatorTrustedPandoraFlag      = "validator-trusted-pandora"

	// Vanguard related flag names
	vanguardTagFlag                     = "vanguard-tag"
	vanguardGenesisStateFlag            = "vanguard-genesis-state"
	vanguardDatadirFlag                 = "vanguard-datadir"
	vanguardBootnodesFlag               = "vanguard-bootnode"
	vanguardWeb3ProviderFlag            = "vanguard-web3provider"
	vanguardDepositContractFlag         = "vanguard-deposit-contract"
	vanguardContractDeploymentBlockFlag = "vanguard-deposit-deployment"
	vanguardVerbosityFlag               = "vanguard-verbosity"
	vanguardMinSyncPeersFlag            = "vanguard-min-sync-peers"
	vanguardMaxSyncPeersFlag            = "vanguard-max-sync-peers"
	vanguardP2pHostFlag                 = "vanguard-p2p-host"

	// Orchestrator related flag names are already present
)

var (
	appFlags     = cmd.CommonFlagSet
	pandoraFlags = []cli.Flag{
		&cli.StringFlag{
			Name:  pandoraTagFlag,
			Usage: "provide a tag of pandora you would like to run",
			Value: "v0.0.10-stable-without-mix-digest",
		},
		&cli.StringFlag{
			Name:  pandoraDatadirFlag,
			Usage: "provide a path you would like to store your data",
			Value: `"./pandora"`,
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
			Name:  pandoraChainIDFlag,
			Usage: "provide chain id if must be different than default",
			Value: "4004181",
		},
		&cli.StringFlag{
			Name:  pandoraPortFlag,
			Usage: "provide port for pandora",
			Value: "30405",
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
			Name:  pandoraGenesisFileFlag,
			Usage: "remote genesis file that will be downloaded to spin up the network",
			// yes, If you wont set it up, I'll get rewards ;]
			Value: "https://storage.googleapis.com/l16-common/pandora/pandora_private_testnet_genesis.json",
		},
		&cli.StringFlag{
			Name:  pandoraNotifyFlag,
			Usage: "this flag is used to pandora engine to notify validator and orchestrator",
			Value: "ws://127.0.0.1:7878,http://127.0.0.1:7877",
		},
		&cli.StringFlag{
			Name:  pandoraVerbosityFlag,
			Usage: "this flag sets up verobosity for pandora",
			Value: "5",
		},
	}
	validatorFlags = []cli.Flag{
		&cli.StringFlag{
			Name:  validatorTagFlag,
			Usage: "provide tag for validator binary. Release must be present in lukso namespace on github",
			Value: "v0.0.16-alpha",
		},
		&cli.StringFlag{
			Name:  validatorVanguardRpcProviderFlag,
			Usage: "provide url without prefix, example: localhost:4000",
			Value: "localhost:4000",
		},
		&cli.StringFlag{
			Name:  vanguardChainConfigFlag,
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
	}
	vanguardFlags = []cli.Flag{
		&cli.StringFlag{
			Name:  vanguardTagFlag,
			Usage: "provide tag for vanguard",
			Value: "v0.0.16-alpha",
		},
		&cli.StringFlag{
			Name: vanguardGenesisStateFlag,
			// TODO: see if it is possible to do this via url
			Usage: "provide genesis.ssz file",
			Value: "./genesis.ssz",
		},
		&cli.StringFlag{
			Name:  vanguardDatadirFlag,
			Usage: "provide vanguard datadir",
			Value: "./vanguard",
		},
		&cli.StringFlag{
			Name:  vanguardBootnodesFlag,
			Usage: `provide coma separated bootnode enr, default: "enr:-Ku4QANldTRLCRUrY9K4OAmk_ATOAyS_sxdTAaGeSh54AuDJXxOYij1fbgh4KOjD4tb2g3T-oJmMjuJyzonLYW9OmRQBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpD1pf1CAAAAAP__________gmlkgnY0gmlwhAoABweJc2VjcDI1NmsxoQKWfbT1atCho149MGMvpgBWUymiOv9QyXYhgYEBZvPBW4N1ZHCCD6A"`,
			Value: `"enr:-Ku4QANldTRLCRUrY9K4OAmk_ATOAyS_sxdTAaGeSh54AuDJXxOYij1fbgh4KOjD4tb2g3T-oJmMjuJyzonLYW9OmRQBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpD1pf1CAAAAAP__________gmlkgnY0gmlwhAoABweJc2VjcDI1NmsxoQKWfbT1atCho149MGMvpgBWUymiOv9QyXYhgYEBZvPBW4N1ZHCCD6A"`,
		},
		&cli.StringFlag{
			Name:  vanguardWeb3ProviderFlag,
			Usage: "provide web3 provider (network of deposit contract deployment), default: http://127.0.0.1:8565",
			Value: "http://127.0.0.1:8565",
		},
		&cli.StringFlag{
			Name:  vanguardDepositContractFlag,
			Usage: "provide deposit contract address",
			Value: "0x000000000000000000000000000000000000cafe",
		},
		&cli.StringFlag{
			Name:  vanguardContractDeploymentBlockFlag,
			Usage: "provide deployment height of deposit contract, default 0.",
			Value: "0",
		},
		&cli.StringFlag{
			Name:  vanguardVerbosityFlag,
			Usage: "provide verobosity for vanguard",
			Value: "info",
		},
		&cli.StringFlag{
			Name:  vanguardMinSyncPeersFlag,
			Usage: "provide min sync peers for vanguard, default 0",
			Value: "0",
		},
		&cli.StringFlag{
			Name:  vanguardMaxSyncPeersFlag,
			Usage: "provide max sync peers for vanguard, default 25",
			Value: "25",
		},
		&cli.StringFlag{
			Name:  vanguardP2pHostFlag,
			Usage: "provide p2p host for vanguard, default 127.0.0.1",
			Value: "127.0.0.1",
		},
	}
)

func prepareVanguardFlags(ctx *cli.Context) (vanguardArguments []string) {
	if !ctx.Bool(cmd.AcceptTOUFlag.Name) {
		log.Fatal("you must accept terms of use")
		ctx.Done()

		return
	}

	vanguardArguments = append(vanguardArguments, "--accept-terms-of-use")

	if ctx.IsSet(cmd.ForceClearDB.Name) {
		vanguardArguments = append(vanguardArguments, "--force-clear-db")
	}

	vanguardArguments = append(vanguardArguments, fmt.Sprintf("--chain-id=%s", ctx.String(pandoraChainIDFlag)))
	vanguardArguments = append(
		vanguardArguments,
		fmt.Sprintf("--network-id=%s", ctx.String(pandoraNetworkIDFlag)))
	vanguardArguments = append(vanguardArguments, fmt.Sprintf("--datadir %s", ctx.String(vanguardDatadirFlag)))

	// This flag can be shared for sure. There is no possibility to use different specs for validator and vanguard.
	vanguardArguments = append(vanguardArguments, fmt.Sprintf(
		"--chain-config-file=%s",
		ctx.String(vanguardChainConfigFlag),
	))
	vanguardArguments = append(vanguardArguments, fmt.Sprintf(
		"--bootstrap-node=%s",
		ctx.String(vanguardBootnodesFlag),
	))
	vanguardArguments = append(vanguardArguments, fmt.Sprintf(
		"--http-web3provider=%s",
		ctx.String(vanguardWeb3ProviderFlag),
	))
	vanguardArguments = append(vanguardArguments, fmt.Sprintf(
		"--deposit-contract=%s",
		ctx.String(vanguardDepositContractFlag),
	))
	vanguardArguments = append(vanguardArguments, fmt.Sprintf(
		"--contract-deployment-block=%s",
		ctx.String(vanguardContractDeploymentBlockFlag),
	))
	vanguardArguments = append(vanguardArguments, "--rpc-host=0.0.0.0")
	vanguardArguments = append(vanguardArguments, "--monitoring-host=0.0.0.0")
	vanguardArguments = append(vanguardArguments, fmt.Sprintf(
		"--verbosity=%s",
		ctx.String(vanguardVerbosityFlag),
	))
	vanguardArguments = append(vanguardArguments, fmt.Sprintf(
		"--min-sync-peers=%s",
		ctx.String(vanguardMinSyncPeersFlag),
	))
	vanguardArguments = append(vanguardArguments, fmt.Sprintf(
		"--p2p-max-peers=%s",
		ctx.String(vanguardMaxSyncPeersFlag),
	))
	vanguardArguments = append(vanguardArguments, fmt.Sprintf(
		"--p2p-host-ip=%s",
		ctx.String(vanguardP2pHostFlag),
	))

	return
}

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
		ctx.String(vanguardChainConfigFlag),
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
	pandoraArguments = append(pandoraArguments, "--datadir")
	pandoraArguments = append(pandoraArguments, ctx.String(pandoraDatadirFlag))

	if len(ctx.String(pandoraEthstatsFlag)) > 1 {
		pandoraArguments = append(pandoraArguments, "--ethstats")
		pandoraArguments = append(pandoraArguments, ctx.String(pandoraEthstatsFlag))
	}

	if len(ctx.String(pandoraBootnodesFlag)) > 1 {
		pandoraArguments = append(pandoraArguments, "--bootnodes")
		pandoraArguments = append(pandoraArguments, ctx.String(pandoraBootnodesFlag))
	}

	pandoraArguments = append(pandoraArguments, "--networkid")
	pandoraArguments = append(pandoraArguments, ctx.String(pandoraNetworkIDFlag))
	pandoraArguments = append(pandoraArguments, "--port")
	pandoraArguments = append(pandoraArguments, ctx.String(pandoraPortFlag))

	// Http api
	// TODO: change to new --http, because -rpc is deprecated in pandora
	pandoraArguments = append(pandoraArguments, "--rpc")
	pandoraArguments = append(pandoraArguments, "--rpcaddr")
	pandoraArguments = append(pandoraArguments, "0.0.0.0")
	pandoraArguments = append(pandoraArguments, "--rpcapi")
	pandoraArguments = append(pandoraArguments, ctx.String(pandoraHttpApiFlag))
	pandoraArguments = append(pandoraArguments, "--rpcport")
	pandoraArguments = append(pandoraArguments, ctx.String(pandoraHttpPortFlag))

	// Websocket
	pandoraArguments = append(pandoraArguments, "--ws")
	pandoraArguments = append(pandoraArguments, "--ws.addr")
	pandoraArguments = append(pandoraArguments, "0.0.0.0")
	pandoraArguments = append(pandoraArguments, "--ws.api")
	pandoraArguments = append(pandoraArguments, ctx.String(pandoraWSApiFlag))
	pandoraArguments = append(pandoraArguments, "--ws.port")
	pandoraArguments = append(pandoraArguments, ctx.String(pandoraWSPortFlag))
	pandoraArguments = append(pandoraArguments, "--ws.origins")
	pandoraArguments = append(pandoraArguments, "'*'")

	// Miner
	pandoraArguments = append(pandoraArguments, "--miner.etherbase")
	pandoraArguments = append(pandoraArguments, ctx.String(pandoraEtherbaseFlag))
	pandoraArguments = append(pandoraArguments, "--mine")
	pandoraArguments = append(pandoraArguments, "--miner.notify")
	pandoraArguments = append(pandoraArguments, ctx.String(pandoraNotifyFlag))

	// Verbosity
	pandoraArguments = append(pandoraArguments, "--verbosity")
	pandoraArguments = append(pandoraArguments, ctx.String(pandoraVerbosityFlag))

	return
}
