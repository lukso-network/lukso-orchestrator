package kv

import (
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"testing"
)

// Test_EncodingDecoding_Success
func Test_EncodingDecoding_Success(t *testing.T) {
	consensusInfo0 := testutil.NewMinimalConsensusInfo(1)
	consensusInfoEncoded0, err := encode(consensusInfo0)
	require.NoError(t, err)

	var consensusInfoDecoded0 *types.MinimalEpochConsensusInfoV2
	require.NoError(t, decode(consensusInfoEncoded0, &consensusInfoDecoded0))
	assert.DeepEqual(t, consensusInfo0, consensusInfoDecoded0)
}
