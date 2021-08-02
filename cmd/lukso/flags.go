package main

import (
	"fmt"
	"github.com/lukso-network/lukso-orchestrator/shared/cmd"
	"github.com/urfave/cli/v2"
	"runtime"
	"strings"
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
	pandoraOutputFlag      = "pandora-output"
	pandoraWsOriginFlag    = "pandora-ws-origin"
	pandoraHttpOriginFlag  = "pandora-http-origin"
	pandoraNatFlag         = "pandora-nat"

	// Common for prysm client
	vanguardChainConfigFlag = "vanguard-chain-config"

	// Validator related flag names
	validatorTagFlag                 = "validator-tag"
	validatorVanguardRpcProviderFlag = "validator-vanguard-rpc"
	validatorVerbosityFlag           = "validator-verbosity"
	validatorTrustedPandoraFlag      = "validator-trusted-pandora"
	validatorWalletPasswordFileFlag  = "validator-wallet-password-file"
	validatorDatadirFlag             = "validator-datadir"
	validatorOutputFileFlag          = "validator-output-file"

	// Vanguard related flag names
	vanguardTagFlag                     = "vanguard-tag"
	vanguardGenesisStateFlag            = "vanguard-genesis-state"
	vanguardDatadirFlag                 = "vanguard-datadir"
	vanguardBootnodesFlag               = "vanguard-bootnode"
	vanguardPeerFlag                    = "vanguard-peer"
	vanguardOutputFlag                  = "vanguard-output"
	vanguardWeb3ProviderFlag            = "vanguard-web3provider"
	vanguardDepositContractFlag         = "vanguard-deposit-contract"
	vanguardContractDeploymentBlockFlag = "vanguard-deposit-deployment"
	vanguardVerbosityFlag               = "vanguard-verbosity"
	vanguardMinSyncPeersFlag            = "vanguard-min-sync-peers"
	vanguardMaxSyncPeersFlag            = "vanguard-max-sync-peers"
	vanguardP2pHostFlag                 = "vanguard-p2p-host"
	vanguardP2pLocalFlag                = "vanguard-p2p-local"
	vanguardP2pUdpPortFlag              = "vanguard-p2p-udp-port"
	vanguardP2pTcpPortFlag              = "vanguard-p2p-tcp-port"
	vanguardOrcProviderFlag             = "vanguard-orc-provider"
	vanguardDisableSyncFlag             = "vanguard-disable-sync"
	vanguardOutputFileFlag              = "vanguard-output-file"

	// Orchestrator related flag names are already present
)

var (
	appFlags                 = cmd.CommonFlagSet
	vanguardDefaultBootnodes = []string{
		// Bootnode library on A
		"enr:-Ku4QLjFvoOKPJNP6u4h5Lf3RzB-voVdpeg0ibv0qZN4ZbU8QXnCzEnTguQLAjJ3kpuZx07nDVcBLcK3U0Ukr1EGsXoBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpD1pf1CAAAAAP__________gmlkgnY0gmlwhCPGdJiJc2VjcDI1NmsxoQKWfbT1atCho149MGMvpgBWUymiOv9QyXYhgYEBZvPBW4N1ZHCCJw8",
		// A
		"enr:-LK4QHaSy__vW9HcjoyR2rRk7T08aVvAedYxOsFIBbL-MHu_Nhqd3SUaMuWMFH9Q1nmee_LCLbH2D7wHS-OFcbKIarsBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpDm-OYdAAAAAP__________gmlkgnY0gmlwhCPGdJiJc2VjcDI1NmsxoQPs4NO21PxfNDDb-SG-NnhmAADUPjtTC99O1Es5UgP8yIN0Y3CCMsiDdWRwgi7g",
		// B
		"enr:-LK4QKn9XaGJpOve4_ETG0jAwvBrF6vaGa2A-HYXer1dGkEQRrhWe9qjD02mhaXhTLIZ4asQOEzJbVO5RyvsMXw1BLUBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpDm-OYdAAAAAP__________gmlkgnY0gmlwhCKNbAqJc2VjcDI1NmsxoQL9nxlySi__9Ipf4bXcPjVwAf4KZqUYU8nJqQGV5PLsvoN0Y3CCMsiDdWRwgi7g",
		// D
		"enr:-LK4QEE9_HpA56uLowwVoAFpwxxiSGMSgjEHH89Mwz8Xd6kZbA5saleoK45Fm2rgygEDmCMrrlTH91l2Kaflr23lJDUBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpDm-OYdAAAAAP__________gmlkgnY0gmlwhCKNAdiJc2VjcDI1NmsxoQORvoKE1Og5TlKk0iVi0NQWX9PRlPcYsFOKaSab6kD034N0Y3CCMsiDdWRwgi7g",
		// E
		"enr:-LK4QJDu03pU3nj4qsmvCHJuf6_Kxu2OcV0aOqS0izGuHdr7acwsbzg_eW-pd-0weQLf9GdFaYgDnYXtg5RLRLO3kHcBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpDm-OYdAAAAAP__________gmlkgnY0gmlwhCKNIISJc2VjcDI1NmsxoQJAD14sMWd0_WBkdINr0SHXGK3qX8GXVSZckOQoKShS44N0Y3CCMsiDdWRwgi7g",
		// F
		"enr:-LK4QH91NzPKSYFOEFFPMU-CXiFAjedCLmOi8TBD2Pg_UzrEDSOwNJ76aWy7aC0cZONHFcUBZy8nwOS6nSYgABz1JeYBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpDm-OYdAAAAAP__________gmlkgnY0gmlwhCPGV3-Jc2VjcDI1NmsxoQM_oHLeIM30aLciVxxeGNqOYCZu70U7ZC-P3QvpxZGjKIN0Y3CCMsiDdWRwgi7g",
		// G
		"enr:-LK4QHGHsXNg7YDWCWKRSlKV1OtDSSQBRcmxeXpKc4c2dBWHUO5QEpkfxx8_zEQGM-yAYDZtJ6B3rYZnEyblCygIOBQBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpDm-OYdAAAAAP__________gmlkgnY0gmlwhCKNbE2Jc2VjcDI1NmsxoQNjoSakWCL5ZrJIloWZE-pPQiGiQwILcNDdvnRSiI0oeIN0Y3CCMsiDdWRwgi7g",
		// H
		"enr:-LK4QKY0K5-L8XRv8HN04-00ZkGQF02PJazCDCfPu9TwSqXKDDWkaMEZjfdey1Oml0BVnFsVQyFD-ltD049Ki3YOo1sBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpDm-OYdAAAAAP__________gmlkgnY0gmlwhCJrAOKJc2VjcDI1NmsxoQLBHrb00e59JlSUHEb2KjhzpmTe3osZE268gVrGV5oipIN0Y3CCMsiDdWRwgi7g",
		// Reto
		"enr:-LK4QD-eEheGEyUCTLAVTNeX2M81zDVM6GRlR3rNt3Owlb5PLLSThpDnLpD2Tvjo7KsF6RS-dHuvL7fh1npYxy5H1x8Ch2F0dG5ldHOIQOIggAAgAACEZXRoMpDm-OYdAAAAAP__________gmlkgnY0gmlwhC5_GlKJc2VjcDI1NmsxoQOeotdaC38C_YGSuL14IMD1bd2e6fbjgu-wV_RMg96e3YN0Y3CCMsiDdWRwgi7g",
	}
	pandoraFlags = []cli.Flag{
		&cli.StringFlag{
			Name:  pandoraTagFlag,
			Usage: "provide a tag of pandora you would like to run",
			Value: "v0.0.16-gamma",
		},
		&cli.StringFlag{
			Name:  pandoraDatadirFlag,
			Usage: "provide a path you would like to store your data",
			Value: "./pandora",
		},
		&cli.BoolFlag{
			Name:  pandoraOutputFlag,
			Usage: "do you want to have output attached to your combined output",
			Value: false,
		},
		&cli.StringFlag{
			Name:  pandoraEthstatsFlag,
			Usage: "nickname:STATS_LOGIN_SECRET@PANDORA_STATS_HOST",
			Value: "",
		},
		&cli.StringFlag{
			Name:  pandoraBootnodesFlag,
			Usage: "Default value should be ok for test network. Otherwise provide Comma separated enode urls, see at https://geth.ethereum.org/docs/getting-started/private-net.",
			Value: "enode://967db4f56ad0a1a35e3d30632fa600565329a23aff50c9762181810166f3c15b078cca522f930d1a2747778893232336bffd1ea5d2ca60543f1801d4360ea63a@35.204.255.172:0?discport=30301",
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
			Usage: "port for pandora api",
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
			Value: "3",
		},
		&cli.StringFlag{
			Name:  pandoraWsOriginFlag,
			Usage: "this flag sets up websocket accepted origins, default not set",
			Value: "",
		},
		&cli.StringFlag{
			Name:  pandoraHttpOriginFlag,
			Usage: "this flag sets up http accepted origins, default not set",
			Value: "",
		},
		&cli.StringFlag{
			Name:  pandoraNatFlag,
			Usage: "this flag sets up http nat to assign static ip for geth, default not set. Example `extip:172.16.254.4`",
			Value: "",
		},
	}
	validatorFlags = []cli.Flag{
		&cli.StringFlag{
			Name:  validatorTagFlag,
			Usage: "provide tag for validator binary. Release must be present in lukso namespace on github",
			Value: "v0.0.18-beta",
		},
		&cli.StringFlag{
			Name:  validatorVanguardRpcProviderFlag,
			Usage: "provide url without prefix, example: localhost:4000",
			Value: "localhost:4000",
		},
		&cli.StringFlag{
			Name:  vanguardChainConfigFlag,
			Usage: "path to chain config of vanguard and validator",
			// TODO: Parse it automatically
			Value: "./vanguard/v0.0.18-beta/config.yml",
		},
		&cli.BoolFlag{
			Name:  vanguardOutputFlag,
			Usage: "path to chain config of vanguard and validator",
			// TODO: Parse it automatically
			Value: false,
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
			Name:  validatorWalletPasswordFileFlag,
			Usage: "location of file password that you used for generation keys from deposit-cli",
			Value: "./password.txt",
		},
		&cli.StringFlag{
			Name:  validatorDatadirFlag,
			Usage: "location of keys from deposit-cli",
			Value: "",
		},
		&cli.StringFlag{
			Name:  validatorOutputFileFlag,
			Usage: "provide output destination of vanguard",
			Value: "./validator.log",
		},
	}
	vanguardFlags = []cli.Flag{
		&cli.StringFlag{
			Name:  vanguardTagFlag,
			Usage: "provide tag for vanguard",
			Value: "v0.0.18-beta",
		},
		&cli.StringFlag{
			Name: vanguardGenesisStateFlag,
			// TODO: see if it is possible to do this via url
			Usage: "provide genesis.ssz file",
			Value: "./vanguard/v0.0.18-beta/vanguard_private_testnet_genesis.ssz",
		},
		&cli.StringFlag{
			Name:  vanguardDatadirFlag,
			Usage: "provide vanguard datadir",
			Value: "./vanguard",
		},
		&cli.StringFlag{
			Name:  vanguardBootnodesFlag,
			Usage: fmt.Sprintf(`provide coma separated bootNode enr, default 8 with first record: "%s"`, strings.Join(vanguardDefaultBootnodes, ",")),
			Value: strings.Join(vanguardDefaultBootnodes, ","),
		},
		&cli.StringFlag{
			Name:  vanguardPeerFlag,
			Usage: `provide coma separated peer enr address, default: ""`,
			Value: "",
		},
		&cli.StringFlag{
			Name:  vanguardWeb3ProviderFlag,
			Usage: "provide web3 provider (network of deposit contract deployment), default: http://127.0.0.1:8565",
			Value: cmd.DefaultPandoraRPCEndpoint,
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
			Usage: "provide p2p host for vanguard, default empty",
			Value: "",
		}, &cli.StringFlag{
			Name:  vanguardP2pLocalFlag,
			Usage: "provide p2p local ip for vanguard, default empty",
			Value: "",
		},
		&cli.StringFlag{
			Name:  vanguardP2pUdpPortFlag,
			Usage: "provide p2p udp port for vanguard, default is 12000",
			Value: "12000",
		}, &cli.StringFlag{
			Name:  vanguardP2pTcpPortFlag,
			Usage: "provide p2p tcp udp port for vanguard, default is 13000",
			Value: "13000",
		},
		&cli.StringFlag{
			Name:  vanguardOrcProviderFlag,
			Usage: "provide orchestrator provider, default http://127.0.0.1:7878",
			Value: "http://127.0.0.1:7877",
		},
		&cli.BoolFlag{
			Name:  vanguardDisableSyncFlag,
			Usage: "disable initial sync phase",
			Value: false,
		},
		&cli.StringFlag{
			Name:  vanguardOutputFileFlag,
			Usage: "provide output destination of vanguard",
			Value: "./vanguard.log",
		},
	}
)

// setupOperatingSystem will parse flags and use it to deduce which system dependencies are required
func setupOperatingSystem() {
	systemOs = runtime.GOOS
}

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
	vanguardArguments = append(vanguardArguments, fmt.Sprintf("--datadir"))
	vanguardArguments = append(vanguardArguments, ctx.String(vanguardDatadirFlag))

	// This flag can be shared for sure. There is no possibility to use different specs for validator and vanguard.
	vanguardArguments = append(vanguardArguments, fmt.Sprintf(
		"--chain-config-file=%s",
		ctx.String(vanguardChainConfigFlag),
	))

	vanguardArguments = append(vanguardArguments, splitCommaSeparatedBootNodes(ctx)...)

	if "" != ctx.String(vanguardPeerFlag) {
		vanguardArguments = append(vanguardArguments, fmt.Sprintf(
			"--peer=%s",
			ctx.String(vanguardPeerFlag),
		))
	}

	if ctx.Bool(vanguardDisableSyncFlag) {
		vanguardArguments = append(vanguardArguments, "--disable-sync")
	}

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
	vanguardArguments = append(vanguardArguments, "--verbosity")
	vanguardArguments = append(vanguardArguments, ctx.String(vanguardVerbosityFlag))
	vanguardArguments = append(vanguardArguments, fmt.Sprintf(
		"--min-sync-peers=%s",
		ctx.String(vanguardMinSyncPeersFlag),
	))
	vanguardArguments = append(vanguardArguments, fmt.Sprintf(
		"--p2p-max-peers=%s",
		ctx.String(vanguardMaxSyncPeersFlag),
	))

	if "" != ctx.String(vanguardP2pHostFlag) {
		vanguardArguments = append(vanguardArguments, fmt.Sprintf(
			"--p2p-host-ip=%s",
			ctx.String(vanguardP2pHostFlag),
		))
	}

	if "" != ctx.String(vanguardP2pLocalFlag) {
		vanguardArguments = append(vanguardArguments, fmt.Sprintf(
			"--p2p-local-ip=%s",
			ctx.String(vanguardP2pLocalFlag),
		))
	}

	if "" != ctx.String(vanguardP2pUdpPortFlag) {
		vanguardArguments = append(vanguardArguments, fmt.Sprintf(
			"--p2p-udp-port=%s",
			ctx.String(vanguardP2pUdpPortFlag),
		))
	}

	if "" != ctx.String(vanguardP2pTcpPortFlag) {
		vanguardArguments = append(vanguardArguments, fmt.Sprintf(
			"--p2p-tcp-port=%s",
			ctx.String(vanguardP2pTcpPortFlag),
		))
	}

	vanguardArguments = append(vanguardArguments, fmt.Sprintf(
		"--log-file=%s",
		ctx.String(vanguardOutputFileFlag),
	))
	vanguardArguments = append(vanguardArguments, fmt.Sprintf(
		"--orc-http-provider=%s",
		ctx.String(vanguardOrcProviderFlag),
	))
	vanguardArguments = append(vanguardArguments, fmt.Sprintf(
		"--genesis-state=%s",
		ctx.String(vanguardGenesisStateFlag),
	))
	//vanguardArguments = append(vanguardArguments, "--lukso-network")

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
	validatorArguments = append(validatorArguments, "--verbosity")
	validatorArguments = append(validatorArguments, ctx.String(validatorVerbosityFlag))
	validatorArguments = append(validatorArguments, fmt.Sprintf(
		"--pandora-http-provider=%s",
		ctx.String(validatorTrustedPandoraFlag),
	))
	validatorArguments = append(validatorArguments, fmt.Sprintf(
		"--log-file=%s",
		ctx.String(validatorOutputFileFlag),
	))
	validatorArguments = append(validatorArguments, fmt.Sprintf(
		"--wallet-password-file=%s",
		ctx.String(validatorWalletPasswordFileFlag),
	))
	validatorArguments = append(validatorArguments, fmt.Sprintf(
		"--datadir=%s",
		ctx.String(vanguardDatadirFlag),
	))

	if "" != ctx.String(validatorDatadirFlag) {
		validatorArguments = append(validatorArguments, fmt.Sprintf(
			"--wallet-dir=%s",
			ctx.String(validatorDatadirFlag),
		))
	}

	validatorArguments = append(validatorArguments, "--lukso-network")

	// Added web interface for vanguard, default address is http://localhost:7500
	// Some explanation: https://docs.prylabs.network/docs/prysm-usage/web-interface/
	validatorArguments = append(validatorArguments, "--web")

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

	if "" != ctx.String(pandoraHttpOriginFlag) {
		pandoraArguments = append(pandoraArguments, "--http.corsdomain")
		pandoraArguments = append(pandoraArguments, ctx.String(pandoraHttpOriginFlag))
	}

	// Nat extIP
	if "" != ctx.String(pandoraNatFlag) {
		pandoraArguments = append(pandoraArguments, "--nat")
		pandoraArguments = append(pandoraArguments, ctx.String(pandoraNatFlag))
	}

	// Websocket
	pandoraArguments = append(pandoraArguments, "--ws")
	pandoraArguments = append(pandoraArguments, "--ws.addr")
	pandoraArguments = append(pandoraArguments, "0.0.0.0")
	pandoraArguments = append(pandoraArguments, "--ws.api")
	pandoraArguments = append(pandoraArguments, ctx.String(pandoraWSApiFlag))
	pandoraArguments = append(pandoraArguments, "--ws.port")
	pandoraArguments = append(pandoraArguments, ctx.String(pandoraWSPortFlag))

	if "" != ctx.String(pandoraWsOriginFlag) {
		pandoraArguments = append(pandoraArguments, "--ws.origins")
		pandoraArguments = append(pandoraArguments, ctx.String(pandoraWsOriginFlag))
	}

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

func splitCommaSeparatedBootNodes(ctx *cli.Context) (vanguardArguments []string) {
	bootNodesFlag := ctx.String(vanguardBootnodesFlag)

	if !strings.Contains(bootNodesFlag, ",") {
		vanguardArguments = append(vanguardArguments, fmt.Sprintf(
			"--bootstrap-node=%s",
			bootNodesFlag,
		))

		return
	}

	enrs := strings.Split(bootNodesFlag, ",")

	for _, enr := range enrs {
		vanguardArguments = append(vanguardArguments, fmt.Sprintf("--bootstrap-node=%s", enr))
	}

	return
}
