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

		realmSlot := backend.RealmDB.LatestVerifiedRealmSlot()
		require.Equal(t, uint64(2), realmSlot)
	})

	t.Run("should handle skipped blocks", func(t *testing.T) {
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

		// 5 slots will be skipped
		require.NoError(t, orchestratorDB.SavePandoraHeaderHash(5, &types.HeaderHash{
			HeaderHash: pandoraHash,
			Status:     types.Pending,
		}))

		require.NoError(t, orchestratorDB.SaveVanguardHeaderHash(5, &types.HeaderHash{
			HeaderHash: vanguardHash,
			Status:     types.Pending,
		}))

		vanguardErr, pandoraErr, realmErr := backend.InvalidatePendingQueue()
		require.NoError(t, vanguardErr)
		require.NoError(t, pandoraErr)
		require.NoError(t, realmErr)

		status, err := backend.FetchVanBlockStatus(5, vanguardHash)
		require.NoError(t, err)
		require.Equal(t, events.Verified, status)

		status, err = backend.FetchPanBlockStatus(5, pandoraHash)
		require.NoError(t, err)
		require.Equal(t, events.Verified, status)

		realmSlot := backend.RealmDB.LatestVerifiedRealmSlot()
		require.Equal(t, uint64(5), realmSlot)
	})

	t.Run("should handle skipped blocks with different states", func(t *testing.T) {
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

		require.NoError(t, orchestratorDB.SaveVanguardHeaderHash(1, &types.HeaderHash{
			HeaderHash: vanguardHash,
			Status:     types.Verified,
		}))

		// Save state for slot 1 that is verified on each side and start iteration from slot 1
		require.NoError(t, orchestratorDB.SaveLatestVanguardSlot())
		require.NoError(t, orchestratorDB.SaveLatestPandoraSlot())
		require.NoError(t, orchestratorDB.SaveLatestVerifiedRealmSlot(1))

		// on vanguard side slot 2 will be skipped and on vanguard side slot 2 will be skipped
		// on vanguard side slot 3 will be skipped and on vanguard side slot 3 will be skipped
		// on vanguard side slot 4 will be missing, but on pandora side will be present
		// on vanguard side slot 5 will be present, but on pandora slot 5 will be missing
		// on vanguard side slot 6 will be present, and on pandora side slot 6 will be present
		type realmPairSuite []struct {
			slot                   uint64
			pandoraHash            *types.HeaderHash
			vanguardHash           *types.HeaderHash
			expectedPandoraStatus  types.Status
			expectedVanguardStatus types.Status
		}

		testSuite := realmPairSuite{
			{
				slot:                   2,
				pandoraHash:            nil,
				vanguardHash:           nil,
				expectedPandoraStatus:  types.Pending,
				expectedVanguardStatus: types.Pending,
			},
			{
				slot:                   3,
				pandoraHash:            nil,
				vanguardHash:           nil,
				expectedPandoraStatus:  types.Pending,
				expectedVanguardStatus: types.Pending,
			},
			{
				slot: 4,
				pandoraHash: &types.HeaderHash{
					HeaderHash: pandoraHash,
					Status:     types.Pending,
				},
				vanguardHash:           nil,
				expectedPandoraStatus:  types.Pending,
				expectedVanguardStatus: types.Pending,
			},
			{
				slot:        5,
				pandoraHash: nil,
				vanguardHash: &types.HeaderHash{
					HeaderHash: vanguardHash,
					Status:     types.Pending,
				},
				expectedPandoraStatus:  types.Pending,
				expectedVanguardStatus: types.Pending,
			},
			{
				slot: 6,
				pandoraHash: &types.HeaderHash{
					HeaderHash: pandoraHash,
					Status:     types.Pending,
				},
				vanguardHash: &types.HeaderHash{
					HeaderHash: vanguardHash,
					Status:     types.Pending,
				},
				expectedPandoraStatus:  types.Pending,
				expectedVanguardStatus: types.Pending,
			},
		}

		for index, suite := range testSuite {
			if nil != suite.pandoraHash {
				require.NoError(t, orchestratorDB.SavePandoraHeaderHash(suite.slot, suite.pandoraHash), index)
			}

			if nil != suite.vanguardHash {
				require.NoError(t, orchestratorDB.SaveVanguardHeaderHash(suite.slot, suite.vanguardHash), index)
			}
		}

		vanguardErr, pandoraErr, realmErr := backend.InvalidatePendingQueue()
		require.NoError(t, vanguardErr)
		require.NoError(t, pandoraErr)
		require.NoError(t, realmErr)

		for index, suite := range testSuite {
			if nil != suite.vanguardHash {
				status, err := backend.FetchVanBlockStatus(suite.slot, suite.vanguardHash.HeaderHash)
				require.NoError(t, err, index)
				require.Equal(t, events.FromDBStatus(suite.expectedVanguardStatus), status, index)
			}

			if nil != suite.pandoraHash {
				status, err := backend.FetchPanBlockStatus(suite.slot, suite.pandoraHash.HeaderHash)
				require.NoError(t, err, index)
				require.Equal(t, events.FromDBStatus(suite.expectedPandoraStatus), status, index)
			}

		}

		realmSlot := backend.RealmDB.LatestVerifiedRealmSlot()
		require.Equal(t, uint64(6), realmSlot)
	})

	// TODO: consider if needed to test it here
	t.Run("should notify about errors", func(t *testing.T) {

	})
}
