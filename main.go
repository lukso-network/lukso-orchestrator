package main

import "github.com/lukso-network/lukso-orchestrator/orchestrator"

func main() {
	orchestratorClient, err := orchestrator.New(nil)

	if nil != err {
		panic(err.Error())
	}

	err = orchestratorClient.Run()

	if nil != err {
		panic(err.Error())
	}
}
