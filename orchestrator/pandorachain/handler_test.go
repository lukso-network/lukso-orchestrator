package pandorachain

import (
	"context"
	"github.com/ethereum/go-ethereum/rlp"
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

	// Prepare new pandora header with extraData with BLS signature
	extraDataWithSig, err := testutil.GenerateExtraDataWithBLSSig(newPanHeader)
	require.NoError(t, err)
	newPanHeader.Extra, err = rlp.EncodeToBytes(extraDataWithSig)
	require.NoError(t, err)
	require.NoError(t, panSvc.OnNewPendingHeader(ctx, newPanHeader))
}
