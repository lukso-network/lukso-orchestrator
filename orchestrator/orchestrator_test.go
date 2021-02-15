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
	orchestratorClient, err := New(nil)
	assert.Nil(t, err)
	assert.IsType(t, &Orchestrator{}, orchestratorClient)

	timeout, err := time.ParseDuration("10s")
	assert.Nil(t, err)

	err, errList, runningContainers := orchestratorClient.Run(&timeout)
	assert.Nil(t, err)
	assert.Empty(t, errList)
	assert.Len(t, runningContainers, 2)

	// This test is very naive, but you should have output in your console from docker images.
	// It tests by logic if stopSignal works, in other way this would be infinite
	// TODO: copy stdout and crawl or assume if received informations are received from container
	t.Run("Should attach logs from containers", func(t *testing.T) {
		stopChan := make(chan bool)
		go time.AfterFunc(timeout, func() {
			stopChan <- true
		})

		err := orchestratorClient.LogsFromContainers(runningContainers, stopChan)
		assert.Nil(t, err)
	})

	defer func() {
		err, errList = orchestratorClient.stopTekuCatalystImages(&timeout)
		assert.Nil(t, err)
		assert.Empty(t, errList)
	}()
}

func stopAllContainers(orchestratorClient *Orchestrator) {
	containerList, _ := orchestratorClient.findRunningContainerByImage(CatalystImage)
	timeout, _ := time.ParseDuration("2s")

	// Kill all leftovers
	if len(containerList) > 0 {
		_, _ = orchestratorClient.stopContainers(containerList, &timeout)
	}
}
