package vanguardchain

import (
	"context"
	testDB "github.com/lukso-network/lukso-orchestrator/orchestrator/db/testing"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"testing"
)

func setupVanguardChain(t *testing.T) {

}

func TestVanguradChainStartStop_Initialized(t *testing.T) {
	hook := logTest.NewGlobal()
	ctx := context.Background()
	orchestratorDB := testDB.SetupDB(t)

}
