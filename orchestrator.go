package orchestrator

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"io"
	"os"
)

const (
	TekuCatalystImage          = "silesiacoin/ssc20-client:v002"
	CatalystClientName         = "catalyst"
	CatalystContainerName      = "luksoCatalyst"
	LuksoContainerNameSelector = "lukso_container_name"
	TekuClientName             = "teku"
	CatalystArguments          = "./geth --rpc --rpcapi net,eth,eth2,web3,personal,admin,db,debug,miner,shh,txpool --etherbase %s --datadir %s --rpccorsdomain \"*\" --rpcaddr \"localhost\" --verbosity 5 --unlock 0 --password \"/root/multinet/repo/data/common/node.pwds\" --targetgaslimit '9000000000000' --allow-insecure-unlock --txpool.processtxs  --txpool.accountslots 10000 --txpool.accountqueue 20000"
	TekuArguments              = `./teku/bin/teku --Xinterop-enabled=true \
--Xinterop-owned-validator-count=%d \
--network=minimal \
--p2p-enabled=true \
--p2p-discovery-enabled=true \
--initial-state %s \
--eth1-engine http://%s:8545 \
--rest-api-interface=%s \
--rest-api-host-allowlist="*" \
--rest-api-port=5051 \
--logging=all \
--log-file=%s \
--rest-api-enabled=true \
--metrics-enabled=true \
--p2p-discovery-bootnodes=%s \
--Xinterop-owned-validator-start-index "%d"`
)

type Orchestrator struct {
	params    Params
	dockerCli *client.Client
}

type Params struct{}

func New(params *Params) (orchestrator *Orchestrator, err error) {
	orchestrator = &Orchestrator{params: *params}
	dockerClient, err := orchestrator.newDockerClient()
	orchestrator.dockerCli = dockerClient

	return
}

func (orchestrator *Orchestrator) SpinEth1(ctx *context.Context, clientName string) (eth1Client *interface{}, err error) {
	if CatalystClientName != clientName {
		err = fmt.Errorf("client %s not supported, valid %s", clientName, CatalystClientName)

		return
	}

	imageList, err := orchestrator.findRunningContainer(CatalystContainerName)

	if nil != err {
		return
	}

	if len(imageList) > 0 {
		err = fmt.Errorf("%s container should not be running in docker, but it is", CatalystContainerName)

		return
	}

	return
}

func (orchestrator *Orchestrator) SpinEth2(ctx *context.Context, clientName string) (eth2Client *interface{}, err error) {
	orchestrator.assureDockerClient()

	return
}

// For now lets steer from ENV variables,
// TODO: provide documentation of possible env that you can use, and --help in cli
func (orchestrator *Orchestrator) newDockerClient() (dockerCli *client.Client, err error) {
	dockerCli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if nil != err {
		return
	}

	// This logic will be removed from here, but right now we have common docker image for teku and catalyst
	// so we pull it once
	ctx := context.Background()
	out, err := dockerCli.ImagePull(ctx, TekuCatalystImage, types.ImagePullOptions{})

	if nil != err {
		return
	}

	_, err = io.Copy(os.Stdout, out)

	return
}

func (orchestrator *Orchestrator) assureDockerClient() (dockerCli *client.Client) {
	dockerCli = orchestrator.dockerCli

	if nil == dockerCli {
		panic(fmt.Errorf("orchestrator only works in docker mode for now, please use New() func"))
	}

	return
}

func (orchestrator *Orchestrator) findRunningContainer(containerName string) (
	containerList []types.ImageSummary,
	err error,
) {
	ctx := context.Background()
	dockerCli := orchestrator.assureDockerClient()

	containerFilter := filters.NewArgs(filters.KeyValuePair{
		Key:   LuksoContainerNameSelector,
		Value: containerName,
	})

	containerList, err = dockerCli.ImageList(ctx, types.ImageListOptions{
		All:     true,
		Filters: containerFilter,
	})

	return
}
