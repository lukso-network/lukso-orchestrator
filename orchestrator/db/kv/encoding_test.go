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

	var consensusInfoDecoded0 *types.MinimalEpochConsensusInfo
	require.NoError(t, decode(consensusInfoEncoded0, &consensusInfoDecoded0))
	assert.DeepEqual(t, consensusInfo0, consensusInfoDecoded0)
}

// Test_EncodingDecoding_PanHeader
func Test_EncodingDecoding_PanHeaderHash(t *testing.T) {
	encPandHeaderHash := testutil.NewPandoraHeaderHash(uint64(0), types.Status(0))
	enc0, err := encode(encPandHeaderHash)
	require.NoError(t, err)

	var decPanHeaderHash *types.HeaderHash
	require.NoError(t, decode(enc0, &decPanHeaderHash))
	assert.DeepEqual(t, encPandHeaderHash, decPanHeaderHash)
}
