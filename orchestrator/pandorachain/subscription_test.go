package pandorachain

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"testing"
	"time"
)

// Test_PandoraSvc_PendingHeaderSub checks pending new header subscription method
func Test_PandoraSvc_PendingHeaderSub(t *testing.T) {
	hook := logTest.NewGlobal()
	ctx := context.Background()
	reConPeriod = 1 * time.Second

	inProcServer, panService := SetupInProcServer(t)
	defer inProcServer.Stop()

	panSvc := SetupPandoraSvc(ctx, t, DialInProcClient(inProcServer))
	client := rpc.DialInProc(inProcServer)
	filter := &types.PandoraPendingHeaderFilter{
		FromBlockHash: common.HexToHash("0000000000000000000000000000000000000000000000000000000000000034"),
	}
	panSvc.SubscribePendingHeaders(ctx, filter, "eth", client)
	panService.pendingHeaderCh <- testutil.NewEth1Header(1)

	time.Sleep(1 * time.Second)
	assert.LogsContain(t, hook, "Got new pending header from pandora")
}
