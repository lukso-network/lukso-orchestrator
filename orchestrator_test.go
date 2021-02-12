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

	stopAllContainers := func() {
		containerList, _ := orchestratorClient.findRunningContainerByImage(TekuCatalystImage)
		timeout, _ := time.ParseDuration("2s")

		// Kill all leftovers
		if len(containerList) > 0 {
			_, _ = orchestratorClient.stopContainers(containerList, &timeout)
		}
	}

	stopAllContainers()
	defer stopAllContainers()

	containerBody, err := orchestratorClient.SpinEth1(CatalystClientName)
	assert.Nil(t, err)
	assert.Len(t, containerBody.ID, 64)

	t.Run("Should return error when container is already running", func(t *testing.T) {
		_, err := orchestratorClient.SpinEth1(CatalystClientName)
		assert.Error(t, err)
	})
}
