package api

import (
	"github.com/ethereum/go-ethereum/common"
	testDB "github.com/lukso-network/lukso-orchestrator/orchestrator/db/testing"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/rpc/api/events"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"github.com/stretchr/testify/require"
	"math/rand"
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

		require.NoError(t, orchestratorDB.SaveLatestPandoraSlot())
		require.NoError(t, orchestratorDB.SaveLatestPandoraHeaderHash())

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

	t.Run("should return invalid when hash does not match", func(t *testing.T) {
		orchestratorDB := testDB.SetupDB(t)
		backend := Backend{
			PandoraHeaderHashDB: orchestratorDB,
		}

		token := make([]byte, 4)
		rand.Read(token)
		properHash := common.BytesToHash(token)

		invalidToken := make([]byte, 8)
		rand.Read(invalidToken)
		invalidHash := common.BytesToHash(invalidToken)

		require.NoError(t, orchestratorDB.SavePandoraHeaderHash(1, &types.HeaderHash{
			HeaderHash: properHash,
			Status:     types.Pending,
		}))

		require.NoError(t, orchestratorDB.SaveLatestPandoraSlot())
		require.NoError(t, orchestratorDB.SaveLatestPandoraHeaderHash())

		status, err := backend.FetchPanBlockStatus(1, invalidHash)
		require.Error(t, err)
		require.Equal(t, events.Invalid, status)
	})

	t.Run("should return state when present in database", func(t *testing.T) {
		orchestratorDB := testDB.SetupDB(t)
		backend := Backend{
			PandoraHeaderHashDB: orchestratorDB,
		}

		token := make([]byte, 4)
		rand.Read(token)
		properHash := common.BytesToHash(token)
		properHeaderHash := &types.HeaderHash{
			HeaderHash: properHash,
			Status:     types.Verified,
		}

		nextToken := make([]byte, 8)
		rand.Read(nextToken)
		nextHash := common.BytesToHash(nextToken)
		nextProperHeaderHash := &types.HeaderHash{
			HeaderHash: nextHash,
			Status:     types.Pending,
		}

		require.NoError(t, orchestratorDB.SavePandoraHeaderHash(1, properHeaderHash))

		require.NoError(t, orchestratorDB.SavePandoraHeaderHash(2, nextProperHeaderHash))

		require.NoError(t, orchestratorDB.SaveLatestPandoraSlot())
		require.NoError(t, orchestratorDB.SaveLatestPandoraHeaderHash())

		status, err := backend.FetchPanBlockStatus(1, properHash)
		require.NoError(t, err)
		require.Equal(t, events.Verified, status)

		status, err = backend.FetchPanBlockStatus(2, nextHash)
		require.NoError(t, err)
		require.Equal(t, events.Pending, status)
	})
}

func TestBackend_FetchVanBlockStatus(t *testing.T) {
	t.Run("should return error when no database is present", func(t *testing.T) {
		backend := Backend{}
		status, err := backend.FetchVanBlockStatus(0, common.Hash{})
		require.Error(t, err)
		require.Equal(t, events.Invalid, status)
	})

	t.Run("should return pending, if slot is higher than known slot", func(t *testing.T) {
		orchestratorDB := testDB.SetupDB(t)
		backend := Backend{
			VanguardHeaderHashDB: orchestratorDB,
		}

		require.NoError(t, orchestratorDB.SaveVanguardHeaderHash(1, &types.HeaderHash{
			HeaderHash: common.Hash{},
			Status:     types.Pending,
		}))

		require.NoError(t, orchestratorDB.SaveLatestVanguardSlot())
		require.NoError(t, orchestratorDB.SaveLatestVanguardHeaderHash())

		status, err := backend.FetchVanBlockStatus(2, common.Hash{})
		require.NoError(t, err)

		require.Equal(t, events.Pending, status)
	})

	t.Run("should return error when hash is empty", func(t *testing.T) {
		orchestratorDB := testDB.SetupDB(t)
		backend := Backend{
			VanguardHeaderHashDB: orchestratorDB,
		}

		require.NoError(t, orchestratorDB.SaveVanguardHeaderHash(1, &types.HeaderHash{
			HeaderHash: common.Hash{},
			Status:     types.Pending,
		}))

		require.NoError(t, orchestratorDB.SaveLatestVanguardSlot())
		require.NoError(t, orchestratorDB.SaveLatestVanguardHeaderHash())

		status, err := backend.FetchVanBlockStatus(1, common.Hash{})
		require.Error(t, err)
		require.Equal(t, events.Invalid, status)
	})

	t.Run("should return invalid when hash does not match", func(t *testing.T) {
		orchestratorDB := testDB.SetupDB(t)
		backend := Backend{
			VanguardHeaderHashDB: orchestratorDB,
		}

		token := make([]byte, 4)
		rand.Read(token)
		properHash := common.BytesToHash(token)

		invalidToken := make([]byte, 8)
		rand.Read(invalidToken)
		invalidHash := common.BytesToHash(invalidToken)

		require.NoError(t, orchestratorDB.SaveVanguardHeaderHash(1, &types.HeaderHash{
			HeaderHash: properHash,
			Status:     types.Pending,
		}))

		require.NoError(t, orchestratorDB.SaveLatestVanguardSlot())
		require.NoError(t, orchestratorDB.SaveLatestVanguardHeaderHash())

		status, err := backend.FetchVanBlockStatus(1, invalidHash)
		require.Error(t, err)
		require.Equal(t, events.Invalid, status)
	})

	t.Run("should return state when present in database", func(t *testing.T) {
		orchestratorDB := testDB.SetupDB(t)
		backend := Backend{
			VanguardHeaderHashDB: orchestratorDB,
		}

		token := make([]byte, 4)
		rand.Read(token)
		properHash := common.BytesToHash(token)
		properHeaderHash := &types.HeaderHash{
			HeaderHash: properHash,
			Status:     types.Verified,
		}

		nextToken := make([]byte, 8)
		rand.Read(nextToken)
		nextHash := common.BytesToHash(nextToken)
		nextProperHeaderHash := &types.HeaderHash{
			HeaderHash: nextHash,
			Status:     types.Pending,
		}

		require.NoError(t, orchestratorDB.SaveVanguardHeaderHash(1, properHeaderHash))

		require.NoError(t, orchestratorDB.SaveVanguardHeaderHash(2, nextProperHeaderHash))

		require.NoError(t, orchestratorDB.SaveLatestVanguardSlot())
		require.NoError(t, orchestratorDB.SaveLatestVanguardHeaderHash())

		status, err := backend.FetchVanBlockStatus(1, properHash)
		require.NoError(t, err)
		require.Equal(t, events.Verified, status)

		status, err = backend.FetchVanBlockStatus(2, nextHash)
		require.NoError(t, err)
		require.Equal(t, events.Pending, status)
	})
}

func TestBackend_InvalidatePendingQueue(t *testing.T) {
	t.Run("should invalidate matched blocks on slot 1", func(t *testing.T) {
		orchestratorDB := testDB.SetupDB(t)
		backend := Backend{
			PandoraHeaderHashDB:  orchestratorDB,
			VanguardHeaderHashDB: orchestratorDB,
			RealmDB:              orchestratorDB,
		}

		pandoraToken := make([]byte, 4)
		rand.Read(pandoraToken)
		pandoraHash := common.BytesToHash(pandoraToken)

		vanguardToken := make([]byte, 8)
		rand.Read(vanguardToken)
		vanguardHash := common.BytesToHash(vanguardToken)

		require.NoError(t, orchestratorDB.SavePandoraHeaderHash(1, &types.HeaderHash{
			HeaderHash: pandoraHash,
			Status:     types.Verified,
		}))

		require.NoError(t, orchestratorDB.SaveLatestPandoraSlot())
		require.NoError(t, orchestratorDB.SaveLatestPandoraHeaderHash())

		require.NoError(t, orchestratorDB.SaveVanguardHeaderHash(1, &types.HeaderHash{
			HeaderHash: vanguardHash,
			Status:     types.Verified,
		}))

		require.NoError(t, orchestratorDB.SavePandoraHeaderHash(2, &types.HeaderHash{
			HeaderHash: pandoraHash,
			Status:     types.Pending,
		}))

		require.NoError(t, orchestratorDB.SaveVanguardHeaderHash(2, &types.HeaderHash{
			HeaderHash: vanguardHash,
			Status:     types.Pending,
		}))

		vanguardErr, pandoraErr, realmErr := backend.InvalidatePendingQueue()
		require.NoError(t, vanguardErr)
		require.NoError(t, pandoraErr)
		require.NoError(t, realmErr)

		status, err := backend.FetchVanBlockStatus(2, vanguardHash)
		require.NoError(t, err)
		require.Equal(t, events.Verified, status)

		status, err = backend.FetchPanBlockStatus(2, pandoraHash)
		require.NoError(t, err)
		require.Equal(t, events.Verified, status)
	})
}
