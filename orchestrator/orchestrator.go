package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	volumetypes "github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"io"
	"os"
	"time"
)

const (
	CatalystImage              = "silesiacoin/ssc20-client:v002"
	TekuImage                  = "consensys/teku:latest"
	CatalystClientName         = "catalyst"
	CatalystContainerName      = "luksoCatalyst"
	LuksoContainerNameSelector = "label"
	TekuClientName             = "teku"
	VolumeName                 = "orchestrator-volume"
	VolumePath                 = "/tmp/lukso-orchestrator"
	DarwinDefaultVolumePath    = "/var/lib/docker/volumes/orchestrator-volume/_data"
)

var (
	catalystArguments = []string{
		"./geth",
		"--rpc",
		"--rpcapi", "net,eth,eth2,web3,personal,admin,db,debug,miner,shh,txpool",
		"--rpccorsdomain", "*",
		"--rpcaddr", "0.0.0.0",
		"--verbosity", "4",
		"--txpool.processtxs", "--txpool.accountslots", "10000", "--txpool.accountqueue", "20000",
		"--datadir", DarwinDefaultVolumePath,
	}

	tekuArguments = []string{
		"./teku/bin/teku",
		"--network=minimal",
		"--p2p-enabled=true",
		`--rest-api-host-allowlist="*"`,
		"--rest-api-port=5051",
		"--rest-api-enabled=true",
		"--metrics-enabled=true",
		"--eth1-endpoint=http://34.78.227.45:8545",
		"--eth1-deposit-contract-address=0xEEBbf8e25dB001f4EC9b889978DC79B302DF9Efd",
		fmt.Sprintf("--data-base-path=%s", DarwinDefaultVolumePath),
	}

	orchestratorVolumePath = "/tmp/lukso-orchestrator"
)

type Orchestrator struct {
	params    *Params
	dockerCli *client.Client
}

type Params struct{}

func New(params *Params) (orchestratorClient *Orchestrator, err error) {
	orchestratorClient = &Orchestrator{params: params}
	dockerClient, err := orchestratorClient.newDockerClient()
	orchestratorClient.dockerCli = dockerClient

	return
}

func (orchestratorClient *Orchestrator) PrepareVolume(name, path string) (
	volume types.Volume, err error,
) {
	if 0 == len(path) || 0 == len(name) {
		return volume, errors.New("you need to provide proper name or path - not empty")
	}
	volumeOptions := volumetypes.VolumeCreateBody{
		Name: name,
	}
	volume, err = orchestratorClient.dockerCli.VolumeCreate(
		context.Background(),
		volumeOptions,
	)
	return volume, nil
}

func (orchestratorClient *Orchestrator) SpinEth1(clientName string) (
	containerBody container.ContainerCreateCreatedBody,
	err error,
) {
	if CatalystClientName != clientName {
		err = fmt.Errorf("client %s not supported, valid %s", clientName, CatalystClientName)

		return
	}

	err = orchestratorClient.guardContainerUniqueness(CatalystImage)

	if nil != err {
		return
	}

	containerBody, err = orchestratorClient.createCatalystContainer()

	if nil != err {
		return
	}

	err = orchestratorClient.dockerCli.ContainerStart(
		context.Background(),
		containerBody.ID,
		types.ContainerStartOptions{},
	)

	return
}

func (orchestratorClient *Orchestrator) SpinEth2(clientName string) (
	containerBody container.ContainerCreateCreatedBody,
	err error,
) {
	orchestratorClient.assureDockerClient()

	if TekuClientName != clientName {
		err = fmt.Errorf("client %s not supported, valid %s", clientName, TekuClientName)

		return
	}

	// TODO: This should be handled in other manner. Docker labeling seems most reasonable, or use two separate images
	// for each other. I leave it as it is (common image). Switch to Teku image, that is not used for now.
	err = orchestratorClient.guardContainerUniqueness(TekuImage)

	if nil != err {
		return
	}

	containerBody, err = orchestratorClient.createTekuContainer()

	if nil != err {
		return
	}

	err = orchestratorClient.dockerCli.ContainerStart(
		context.Background(),
		containerBody.ID,
		types.ContainerStartOptions{},
	)

	return
}

// Run will return err if any of errors happened, and errList with all of the errors that had happened.
// Best to assume that if errList is not empty run was faulty
func (orchestratorClient *Orchestrator) Run(timeout *time.Duration) (
	err error,
	errList []error,
	runningContainers []container.ContainerCreateCreatedBody,
) {
	errList = make([]error, 0)
	runningContainers = make([]container.ContainerCreateCreatedBody, 0)
	err, errList = orchestratorClient.stopTekuCatalystImages(timeout)

	if nil != err {
		return
	}

	_, err = orchestratorClient.PrepareVolume(VolumePath, VolumeName)
	if nil != err {
		return
	}

	eth1Container, err := orchestratorClient.SpinEth1(CatalystClientName)

	if nil != err {
		return
	}

	runningContainers = append(runningContainers, eth1Container)
	eth2Container, err := orchestratorClient.SpinEth2(TekuClientName)

	if nil != err {
		return
	}

	runningContainers = append(runningContainers, eth2Container)

	return
}

func (orchestratorClient *Orchestrator) LogsFromContainers(
	containerList []container.ContainerCreateCreatedBody,
	stopChan chan bool,
) (err error) {
	dockerClient := orchestratorClient.assureDockerClient()

	for _, runningContainer := range containerList {
		go func(stopChan chan bool, runningContainer container.ContainerCreateCreatedBody) {
			ctx := context.Background()
			statusCh, errChan := dockerClient.ContainerWait(ctx, runningContainer.ID, container.WaitConditionNotRunning)

			select {
			case <-stopChan:
				fmt.Printf("\n Received stop signal in container id: %s", runningContainer.ID)
				return
			case err := <-errChan:
				fmt.Printf(
					"\n Received error signal in container id: %s, err: %s",
					runningContainer.ID,
					err.Error(),
				)
				stopChan <- true

				return
			case status := <-statusCh:
				fmt.Printf(
					"\n Received not running signal in container id: %s, statusCode: %d, err: %s",
					runningContainer.ID,
					status.StatusCode,
					status.Error,
				)
				stopChan <- true

				return
			default:
				options := types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: true}
				output, err := dockerClient.ContainerLogs(ctx, runningContainer.ID, options)

				if nil != err {
					fmt.Printf("\n Error occured in container: %s, err: %s", runningContainer.ID, err.Error())
				}

				defer func() {
					_ = output.Close()
				}()

				_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, output)

				if nil != err {
					fmt.Printf("\n Error occured in container: %s, err: %s", runningContainer.ID, err.Error())
				}
			}

			stopChan <- true
		}(stopChan, runningContainer)
	}

	select {
	case <-stopChan:
		return
	}
}

func (orchestratorClient *Orchestrator) guardContainerUniqueness(imageName string) (err error) {
	imageList, err := orchestratorClient.findRunningContainerByImage(imageName)

	if nil != err {
		return
	}

	if len(imageList) > 0 {
		err = fmt.Errorf("container from image %s should not be running in docker", CatalystImage)

		return
	}

	return
}

// For now lets steer from ENV variables,
// TODO: provide documentation of possible env that you can use, and --help in cli
func (orchestratorClient *Orchestrator) newDockerClient() (dockerCli *client.Client, err error) {
	dockerCli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if nil != err {
		return
	}

	// This logic will be removed from here, but right now we have common docker image for teku and catalyst
	// so we pull it once
	ctx := context.Background()
	catalystPullOutput, err := dockerCli.ImagePull(ctx, CatalystImage, types.ImagePullOptions{})

	if nil != err {
		return
	}

	tekuPullOutput, err := dockerCli.ImagePull(ctx, TekuImage, types.ImagePullOptions{})

	if nil != err {
		return
	}

	defer func() {
		_ = catalystPullOutput.Close()
		_ = tekuPullOutput.Close()
	}()

	_, err = io.Copy(os.Stdout, catalystPullOutput)

	if nil != err {
		return
	}

	_, err = io.Copy(os.Stderr, tekuPullOutput)

	return
}

func (orchestratorClient *Orchestrator) assureDockerClient() (dockerCli *client.Client) {
	dockerCli = orchestratorClient.dockerCli

	if nil == dockerCli {
		panic(fmt.Errorf("orchestratorClient only works in docker mode for now, please use New() func"))
	}

	return
}

func (orchestratorClient *Orchestrator) findRunningContainerByImage(imageName string) (
	containerList []types.Container,
	err error,
) {
	ctx := context.Background()
	dockerCli := orchestratorClient.assureDockerClient()
	allContainers, err := dockerCli.ContainerList(ctx, types.ContainerListOptions{})

	if nil != err {
		return
	}

	containerList = make([]types.Container, 0)

	for _, runningContainer := range allContainers {
		if imageName == runningContainer.Image {
			containerList = append(containerList, runningContainer)
		}
	}

	return
}

func (orchestratorClient *Orchestrator) stopContainers(
	containerList []types.Container,
	timeout *time.Duration,
) (err error, errList []error) {
	dockerClient := orchestratorClient.assureDockerClient()
	ctx := context.Background()
	errList = make([]error, 0)

	for _, containerToStop := range containerList {
		latestError := dockerClient.ContainerStop(ctx, containerToStop.ID, timeout)

		if nil != latestError {
			err = latestError
			errList = append(errList, latestError)
		}
	}

	return
}

func (orchestratorClient *Orchestrator) getVolume(name string) (volume types.Volume, err error) {
	volume, err = orchestratorClient.dockerCli.VolumeInspect(context.Background(), name)
	if nil != err {
		return volume, err
	}
	return volume, nil
}

func (orchestratorClient *Orchestrator) createCatalystContainer() (
	containerBody container.ContainerCreateCreatedBody,
	err error,
) {
	dockerCli := orchestratorClient.assureDockerClient()
	ctx := context.Background()
	volume, err := orchestratorClient.getVolume(VolumeName)
	if nil != err {
		return
	}
	containerBody, err = dockerCli.ContainerCreate(
		ctx,
		&container.Config{
			// For now lets try with root
			User:         "root",
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          false,
			OpenStdin:    false,
			StdinOnce:    false,
			Env:          nil,
			Cmd:          catalystArguments,
			Image:        CatalystImage,
			Labels:       map[string]string{LuksoContainerNameSelector: CatalystContainerName},
		},
		&container.HostConfig{
			Binds: []string{
				volume.Name + ":" + volume.Mountpoint,
			},
		},
		&network.NetworkingConfig{},
		nil,
		"",
	)

	return
}

func (orchestratorClient *Orchestrator) createTekuContainer() (
	containerBody container.ContainerCreateCreatedBody,
	err error,
) {
	dockerCli := orchestratorClient.assureDockerClient()
	ctx := context.Background()
	volume, err := orchestratorClient.getVolume(VolumeName)
	if nil != err {
		return
	}
	containerBody, err = dockerCli.ContainerCreate(
		ctx,
		&container.Config{
			// For now lets try with root
			User:         "root",
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          false,
			OpenStdin:    false,
			StdinOnce:    false,
			Env:          nil,
			Cmd:          tekuArguments,
			Image:        CatalystImage,
		},
		&container.HostConfig{
			Binds: []string{
				volume.Name + ":" + volume.Mountpoint,
			},
		},
		&network.NetworkingConfig{},
		nil,
		"",
	)

	return
}

func (orchestratorClient *Orchestrator) stopTekuCatalystImages(timeout *time.Duration) (err error, errList []error) {
	containerList := make([]types.Container, 0)
	catalystContainerList, err := orchestratorClient.findRunningContainerByImage(CatalystImage)

	if nil != err {
		return
	}

	tekuContainerList, err := orchestratorClient.findRunningContainerByImage(TekuClientName)

	if nil != err {
		return
	}

	containerList = append(containerList, catalystContainerList...)
	containerList = append(containerList, tekuContainerList...)

	err, errList = orchestratorClient.stopContainers(catalystContainerList, timeout)

	return
}
