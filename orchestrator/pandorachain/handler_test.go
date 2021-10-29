package pandorachain

import (
	"context"
	"encoding/json"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"testing"
)

var (
	pandoraForkedBlockStub = `{
  "difficulty": "0x1",
  "extraData": "0xf869c682149f81a41fb8608a7ce6ca9a11c04707fba03315b4a56fab0aa504168979b6e64c6dd64c8b1c7726eea9f5f1149cc20245ed98a4ca580404e55770a495dd031bf4fb96a8432a6e03edcf62b84799719cb956ea8a55fb5685de52e1c28ff6f21c8d5479558b74e6",
  "gasLimit": "0x7A1200",
  "gasUsed": "0x0",
  "hash": "0x0acf03ae123dc232e181d3273114b4fc1ae570f469c64655ccb7bc8c6b6aaa28",
  "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
  "miner": "0x616e6f6e796d6f75730000000000000000000000",
  "mixHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "nonce": "0x0000000000000000",
  "number": "0x127D",
  "parentHash": "0xcbd44e37125c599ff218b966658d720665e1dfa9fd3db0230ad8743e754495d5",
  "receiptsRoot": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
  "sha3Uncles": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
  "size": "0x26A",
  "stateRoot": "0x0d3a89ab3d4d10bbd424902b176ef48c998b3f30656b904b428a46bc40e20acb",
  "timestamp": "0x61785B63",
  "totalDifficulty": "0x8127D",
  "transactions": [],
  "transactionsRoot": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
  "uncles": []
}`
)

// Test_PandoraSvc_OnNewPendingHeader tests OnNewPendingHeader method
func Test_PandoraSvc_OnNewPendingHeader(t *testing.T) {
	ctx := context.Background()
	inProcServer, _ := SetupInProcServer(t)
	defer inProcServer.Stop()

	panSvc := SetupPandoraSvc(ctx, t, DialInProcClient(inProcServer))
	newPanHeader := testutil.NewEth1Header(123)
	require.NoError(t, panSvc.OnNewPendingHeader(ctx, newPanHeader))

	t.Run("should return error when unsupported fork happens", func(t *testing.T) {
		hook := logTest.NewGlobal()
		forkedPanHeader := &eth1Types.Header{}
		require.NoError(t, json.Unmarshal([]byte(pandoraForkedBlockStub), forkedPanHeader))
		require.NoError(t, panSvc.OnNewPendingHeader(ctx, forkedPanHeader))
		require.LogsContain(t, hook, "unsupported fork pair. Hash: 0x0acf03ae123dc232e181d3273114b4fc1ae570f469c64655ccb7bc8c6b6aaa28, slot: 5279")
	})
}
