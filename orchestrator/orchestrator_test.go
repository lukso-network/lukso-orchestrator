package orchestrator

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestOrchestrator_SpinEth1(t *testing.T) {
	orchestratorClient, err := New(nil)
	assert.Nil(t, err)
	assert.IsType(t, &Orchestrator{}, orchestratorClient)

	stopAllContainers(orchestratorClient)
	defer stopAllContainers(orchestratorClient)

	containerBody, err := orchestratorClient.SpinEth1(CatalystClientName)
	assert.Nil(t, err)
	assert.Len(t, containerBody.ID, 64)

	t.Run("Should return error when client is not supported", func(t *testing.T) {
		_, err := orchestratorClient.SpinEth1("Dummy")
		assert.Error(t, err)
	})

	t.Run("Should return error when container is already running", func(t *testing.T) {
		_, err := orchestratorClient.SpinEth1(CatalystClientName)
		assert.Error(t, err)
	})
}

func TestOrchestrator_SpinEth2(t *testing.T) {
	orchestratorClient, err := New(nil)
	assert.Nil(t, err)
	assert.IsType(t, &Orchestrator{}, orchestratorClient)

	stopAllContainers(orchestratorClient)
	defer stopAllContainers(orchestratorClient)

	containerBody, err := orchestratorClient.SpinEth2(TekuClientName)
	assert.Nil(t, err)
	assert.Len(t, containerBody.ID, 64)

	t.Run("Should return error when client is not supported", func(t *testing.T) {
		_, err := orchestratorClient.SpinEth2("Dummy")
		assert.Error(t, err)
	})

	t.Run("Should return error when container is already running", func(t *testing.T) {
		_, err := orchestratorClient.SpinEth2(TekuClientName)
		assert.Error(t, err)
	})
}

// TODO: fill this up
func TestOrchestrator_Run(t *testing.T) {

}

func stopAllContainers(orchestratorClient *Orchestrator) {
	containerList, _ := orchestratorClient.findRunningContainerByImage(TekuCatalystImage)
	timeout, _ := time.ParseDuration("2s")

	// Kill all leftovers
	if len(containerList) > 0 {
		_, _ = orchestratorClient.stopContainers(containerList, &timeout)
	}
}
