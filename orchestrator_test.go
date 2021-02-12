package orchestrator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOrchestrator_SpinEth1(t *testing.T) {
	orchestratorClient, err := New(nil)
	assert.Nil(t, err)
	assert.IsType(t, &Orchestrator{}, orchestratorClient)

	containerList, err := orchestratorClient.findRunningContainerByImage(TekuCatalystImage)
	assert.Nil(t, err)

	assert.Len(t, containerList, 0)

	containerBody, err := orchestratorClient.SpinEth1(CatalystClientName)
	assert.Nil(t, err)
	assert.Nil(t, containerBody)
}
