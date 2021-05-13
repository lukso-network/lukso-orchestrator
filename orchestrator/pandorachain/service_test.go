package pandorachain

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/pkg/errors"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"testing"
	"time"
)

// Test_PandoraSvc_StartStop checks start and stop process. When the pandora service starts, it also subscribes
// pan_subscribe to get new pending headers
func Test_PandoraSvc_StartStop(t *testing.T) {
	hook := logTest.NewGlobal()
	ctx := context.Background()
	reConPeriod = 1 * time.Second

	inProcServer, _ := SetupInProcServer(t)
	defer inProcServer.Stop()

	panSvc := SetupPandoraSvc(ctx, t, DialInProcClient(inProcServer))
	panSvc.Start()

	time.Sleep(1 * time.Second)
	assert.LogsContain(t, hook, "Connected and subscribed to pandora chain")

	hook.Reset()
	assert.NoError(t, panSvc.Stop())
}

// Test_PandoraSvc_RetrySub checks retry option of pandora service.
func Test_PandoraSvc_RetrySub(t *testing.T) {
	hook := logTest.NewGlobal()
	ctx := context.Background()
	reConPeriod = 1 * time.Second

	inProcServer, _ := SetupInProcServer(t)
	defer inProcServer.Stop()

	panSvc := SetupPandoraSvc(ctx, t, DialInProcClient(inProcServer))
	panSvc.Start()

	time.Sleep(1 * time.Second)
	assert.LogsContain(t, hook, "Connected and subscribed to pandora chain")

	panSvc.conInfoSubErrCh <- errors.New("Error in subscription")
	time.Sleep(2 * time.Second)
	assert.LogsContain(t, hook, "Connected and subscribed to pandora chain")

	hook.Reset()
	assert.NoError(t, panSvc.Stop())
}
