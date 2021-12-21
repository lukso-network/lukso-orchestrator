package consensus

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"testing"
	"time"
)

func TestService_Start(t *testing.T) {
	hook := logTest.NewGlobal()
	ctx := context.Background()
	svc, _, _, _, _ := setup(ctx, t)
	defer svc.Stop()

	svc.Start()
	time.Sleep(1 * time.Second)
	assert.LogsContain(t, hook, "Starting consensus service")
	hook.Reset()
}
