package pandorachain

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	"testing"
)

// Test_PandoraSvc_OnNewPendingHeader tests OnNewPendingHeader method
func Test_PandoraSvc_OnNewPendingHeader(t *testing.T) {
	ctx := context.Background()
	inProcServer, _ := SetupInProcServer(t)
	defer inProcServer.Stop()

	panSvc := SetupPandoraSvc(ctx, t, DialInProcClient(inProcServer))
	newPanHeader := testutil.NewEth1Header(123)
	require.NoError(t, panSvc.OnNewPendingHeader(ctx, newPanHeader))
}
