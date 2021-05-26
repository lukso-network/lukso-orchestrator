package api

import (
	"github.com/ethereum/go-ethereum/common"
	testDB "github.com/lukso-network/lukso-orchestrator/orchestrator/db/testing"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/rpc/api/events"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBackend_FetchPanBlockStatus(t *testing.T) {
	t.Run("should return error when no database is present", func(t *testing.T) {
		backend := Backend{}
		status, err := backend.FetchPanBlockStatus(0, common.Hash{})
		require.Error(t, err)
		require.Equal(t, events.Invalid, status)
	})

	t.Run("should return pending, if slot is higher than known slot", func(t *testing.T) {
		orchestratorDB := testDB.SetupDB(t)
		backend := Backend{
			PandoraHeaderHashDB: orchestratorDB,
		}

		require.NoError(t, orchestratorDB.SavePandoraHeaderHash(1, &types.HeaderHash{
			HeaderHash: common.Hash{},
			Status:     types.Pending,
		}))

		status, err := backend.FetchPanBlockStatus(2, common.Hash{})
		require.NoError(t, err)

		require.Equal(t, events.Pending, status)
	})

	t.Run("should return error when hash is empty", func(t *testing.T) {
		orchestratorDB := testDB.SetupDB(t)
		backend := Backend{
			PandoraHeaderHashDB: orchestratorDB,
		}

		require.NoError(t, orchestratorDB.SavePandoraHeaderHash(1, &types.HeaderHash{
			HeaderHash: common.Hash{},
			Status:     types.Pending,
		}))

		require.NoError(t, orchestratorDB.SaveLatestPandoraSlot())
		require.NoError(t, orchestratorDB.SaveLatestPandoraHeaderHash())

		status, err := backend.FetchPanBlockStatus(1, common.Hash{})
		require.Error(t, err)
		require.Equal(t, events.Invalid, status)
	})

	t.Run("should return invalid when slot does not match", func(t *testing.T) {

	})

	t.Run("should return valid when present in database", func(t *testing.T) {

	})

	t.Run("should return pending, when pending in database", func(t *testing.T) {

	})
}
