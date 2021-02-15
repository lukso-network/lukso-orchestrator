package spinner

import (
	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestOrchestrator_PrepareVolume(t *testing.T) {
	orchestratorClient, err := New(nil)
	assert.Nil(t, err)
	assert.IsType(t, &Orchestrator{}, orchestratorClient)
	stopAllContainers(orchestratorClient)
	defer stopAllContainers(orchestratorClient)
	t.Run("Should be volume create failure", func(t *testing.T) {
		exceptedVolumeOption := types.Volume{}
		volume, err := orchestratorClient.PrepareVolume("", "")
		assert.Error(t, err)
		assert.Equal(t, exceptedVolumeOption, volume)
	})
	t.Run("Should be volume create success", func(t *testing.T) {
		exceptedVolumeOption := types.Volume{
			Mountpoint: DarwinDefaultVolumePath,
			Name:       VolumeName,
		}
		volume, err := orchestratorClient.PrepareVolume(VolumeName, VolumePath)
		assert.Nil(t, err)
		assert.Equal(t, exceptedVolumeOption.Mountpoint, volume.Mountpoint)
		assert.Equal(t, exceptedVolumeOption.Name, volume.Name)
	})
}

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

	// TODO: Maintain or discard this design choice.
	//t.Run("Should return error when container is already running", func(t *testing.T) {
	//	_, err := orchestratorClient.SpinEth2(TekuClientName)
	//	assert.Error(t, err)
	//})
}

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
	containerList := make([]types.Container, 0)
	catalystList, err := orchestratorClient.findRunningContainerByImage(CatalystImage)

	if nil != err {
		panic(err.Error())
	}

	containerList = append(containerList, catalystList...)
	tekuList, err := orchestratorClient.findRunningContainerByImage(TekuImage)

	if nil != err {
		panic(err.Error())
	}

	containerList = append(containerList, tekuList...)

	timeout, _ := time.ParseDuration("2s")

	// Kill all leftovers
	if len(containerList) > 0 {
		_, _ = orchestratorClient.stopContainers(catalystList, &timeout)
	}
}
