package orchestrator

import (
	"github.com/docker/docker/client"
)

type Orchestrator struct {
	params Params
}

type Params struct{}

func New(params *Params) (orchestrator *Orchestrator, err error) {
	orchestrator = &Orchestrator{params: *params}

	return
}

func (orchestrator *Orchestrator) SpinEth1(clientName string) (eth1Client *interface{}, err error) {
	return
}

func (orchestrator *Orchestrator) SpinEth2(clientName string) (eth2Client *interface{}, err error) {
	return
}

// For now lets steer from ENV variables,
// TODO: provide documentation of possible env that you can use, and --help in cli
func (orchestrator *Orchestrator) newDockerClient() (dockerCli *client.Client, err error) {
	dockerCli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	return
}
