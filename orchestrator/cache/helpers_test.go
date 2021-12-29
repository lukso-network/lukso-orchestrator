package cache

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/utils"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPandoraCache_VerifyPandoraCache(t *testing.T) {
	t.Parallel()
	t.Run("fresh start without any previous information", func(t *testing.T) {
		panTestCache := NewPandoraCache(1024, 0, 6, utils.NewStack())
		err := panTestCache.VerifyPandoraCache(nil)
		require.Error(t, err)

		err = panTestCache.VerifyPandoraCache(&PanCacheInsertParams{CurrentVerifiedHeader: nil})
		require.Error(t, err)

		err = panTestCache.VerifyPandoraCache(&PanCacheInsertParams{CurrentVerifiedHeader: new(types.Header), LastVerifiedShardInfo: nil})
		require.NoError(t, err)
	})
}
