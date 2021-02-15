package main

import (
	"fmt"
	"github.com/lukso-network/lukso-orchestrator/tools/spin-up/spinner"
	"time"
)

func main() {
	orchestratorClient, err := spinner.New(nil)

	if nil != err {
		panic(err.Error())
	}

	timeout, _ := time.ParseDuration("3s")
	err, errList, runningContainers := orchestratorClient.Run(&timeout)

	// This will print all encountered errors
	for _, currentErr := range errList {
		fmt.Printf("\n Encountered err: %s", currentErr.Error())
	}

	if nil != err {
		panic(err.Error())
	}

	stopChan := make(chan bool)
	err = orchestratorClient.LogsFromContainers(runningContainers, stopChan)

	if nil != err {
		panic(err.Error())
	}

	select {
	case <-stopChan:
		fmt.Printf("\n Received stop signal")
	}
}
