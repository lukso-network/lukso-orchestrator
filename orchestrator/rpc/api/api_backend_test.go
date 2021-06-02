package api

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	testDB "github.com/lukso-network/lukso-orchestrator/orchestrator/db/testing"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/rpc/api/events"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"github.com/stretchr/testify/require"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"
)

type realmPairSuite []struct {
	slot                   uint64
	pandoraHash            *types.HeaderHash
	vanguardHash           *types.HeaderHash
	expectedPandoraStatus  types.Status
	expectedVanguardStatus types.Status
}

func TestBackend_FetchPanBlockStatus(t *testing.T) {
	t.Run("should return error when no database is present", func(t *testing.T) {
		backend := Backend{}
		status, err := backend.FetchPanBlockStatus(0, common.Hash{})
		require.Error(t, err)
		require.Equal(t, events.Invalid, status)
	})

	t.Run("should return skipped with any hash", func(t *testing.T) {
		orchestratorDB := testDB.SetupDB(t)
		backend := Backend{
			PandoraHeaderHashDB: orchestratorDB,
		}

		require.NoError(t, orchestratorDB.SavePandoraHeaderHash(1, &types.HeaderHash{
			HeaderHash: common.Hash{},
			Status:     types.Skipped,
		}))

		require.NoError(t, orchestratorDB.SaveLatestPandoraSlot())
		require.NoError(t, orchestratorDB.SaveLatestPandoraHeaderHash())

		status, err := backend.FetchPanBlockStatus(1, common.Hash{})
		require.NoError(t, err)

		require.Equal(t, events.Skipped, status)

		status, err = backend.FetchPanBlockStatus(
			1,
			common.HexToHash("0x8040da14e5c49a5ca64802844c5aa7248e78eecf104ac8f4c3176226ced06116"),
		)
		require.NoError(t, err)
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

	t.Run("should return skipped with any hash", func(t *testing.T) {
		orchestratorDB := testDB.SetupDB(t)
		backend := Backend{
			VanguardHeaderHashDB: orchestratorDB,
		}

		require.NoError(t, orchestratorDB.SaveVanguardHeaderHash(1, &types.HeaderHash{
			HeaderHash: common.Hash{},
			Status:     types.Skipped,
		}))

		require.NoError(t, orchestratorDB.SaveLatestVanguardSlot())
		require.NoError(t, orchestratorDB.SaveLatestVanguardHeaderHash())

		status, err := backend.FetchVanBlockStatus(1, common.Hash{})
		require.NoError(t, err)

		require.Equal(t, events.Skipped, status)

		status, err = backend.FetchVanBlockStatus(
			1,
			common.HexToHash("0x8040da14e5c49a5ca64802844c5aa7248e78eecf104ac8f4c3176226ced06116"),
		)
		require.NoError(t, err)
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
		testSuite := realmPairSuite{
			{
				slot:                   2,
				pandoraHash:            nil,
				vanguardHash:           nil,
				expectedPandoraStatus:  types.Skipped,
				expectedVanguardStatus: types.Skipped,
			},
			{
				slot:                   3,
				pandoraHash:            nil,
				vanguardHash:           nil,
				expectedPandoraStatus:  types.Skipped,
				expectedVanguardStatus: types.Skipped,
			},
			{
				slot: 4,
				pandoraHash: &types.HeaderHash{
					HeaderHash: pandoraHash,
					Status:     types.Pending,
				},
				vanguardHash:           nil,
				expectedPandoraStatus:  types.Skipped,
				expectedVanguardStatus: types.Skipped,
			},
			{
				slot:        5,
				pandoraHash: nil,
				vanguardHash: &types.HeaderHash{
					HeaderHash: vanguardHash,
					Status:     types.Pending,
				},
				expectedPandoraStatus:  types.Skipped,
				expectedVanguardStatus: types.Skipped,
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
				expectedPandoraStatus:  types.Verified,
				expectedVanguardStatus: types.Verified,
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
				require.Equal(t, events.FromDBStatus(suite.expectedVanguardStatus), status, suite.slot)
			}

			if nil != suite.pandoraHash {
				status, err := backend.FetchPanBlockStatus(suite.slot, suite.pandoraHash.HeaderHash)
				require.NoError(t, err, index)
				require.Equal(t, events.FromDBStatus(suite.expectedPandoraStatus), status, suite.slot)
			}

		}

		realmSlot := backend.RealmDB.LatestVerifiedRealmSlot()
		require.Equal(t, uint64(6), realmSlot)
	})

	t.Run("should invalidate lots of pending blocks", func(t *testing.T) {
		//orchestratorDB := testDB.SetupDB(t)
		//backend := Backend{
		//	PandoraHeaderHashDB:  orchestratorDB,
		//	VanguardHeaderHashDB: orchestratorDB,
		//	RealmDB:              orchestratorDB,
		//}

		type realmPair struct {
			pandoraSlot  uint64
			pandoraHash  common.Hash
			vanguardSlot uint64
			vanguardHash common.Hash
		}

		batchSize := 5000
		pendingBatch := make([]*realmPair, batchSize)

		// It will fill batch with random order
		// This will simulate network traffic from both parties for existing network
		for index := range pendingBatch {
			rand.Seed(time.Now().UnixNano())
			min := 1
			max := batchSize
			pandoraSlot := rand.Intn(max-min+1) + min

			pandoraToken := make([]byte, 4)
			rand.Read(pandoraToken)
			pandoraHash := common.BytesToHash(pandoraToken)

			vanguardToken := make([]byte, 8)
			rand.Read(vanguardToken)
			vanguardHash := common.BytesToHash(vanguardToken)

			rand.Seed(time.Now().UnixNano())
			vanguardSlot := rand.Intn(max-min+1) + min

			rand.Seed(time.Now().UnixNano())
			pandoraNotPresentSlot := rand.Intn(max-min+1) + min

			rand.Seed(time.Now().UnixNano())
			vanguardNotPresentSlot := rand.Intn(max-min+1) + min

			// Do not fill the queue with 33,(3) % chance
			if pandoraSlot < len(pendingBatch)/3 {
				continue
			}

			pendingBatch[index] = &realmPair{
				pandoraSlot:  uint64(pandoraSlot),
				pandoraHash:  pandoraHash,
				vanguardSlot: uint64(vanguardSlot),
				vanguardHash: vanguardHash,
			}

			pendingBatch[pandoraNotPresentSlot] = &realmPair{
				pandoraSlot: uint64(vanguardNotPresentSlot),
				pandoraHash: pandoraHash,
			}
		}

		t.Logf("pendingBatch: %v", pendingBatch)

	})
}

func TestBackend_StressTestForInvalidatePendingBlocks(t *testing.T) {
	iterations := 0

	if "true" == os.Getenv("STRESS_TEST") {
		iterations = 4
	}

	for index := 0; index < iterations; index++ {
		t.Logf("running stress test suite: %d", index)
		time.Sleep(time.Millisecond * 25)
		t.Run("should work with multiple batch requests", func(t *testing.T) {
			orchestratorDB := testDB.SetupDBWithoutClose(t)
			t.Cleanup(func() {
				require.NoError(t, orchestratorDB.Close())
			})
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

			firstBatch := realmPairSuite{
				{
					slot:                   2,
					pandoraHash:            nil,
					vanguardHash:           nil,
					expectedPandoraStatus:  types.Skipped,
					expectedVanguardStatus: types.Skipped,
				},
				{
					slot:                   3,
					pandoraHash:            nil,
					vanguardHash:           nil,
					expectedPandoraStatus:  types.Skipped,
					expectedVanguardStatus: types.Skipped,
				},
			}

			secondBatch := realmPairSuite{
				{
					slot: 4,
					pandoraHash: &types.HeaderHash{
						HeaderHash: pandoraHash,
						Status:     types.Pending,
					},
					vanguardHash:           nil,
					expectedPandoraStatus:  types.Skipped,
					expectedVanguardStatus: types.Skipped,
				},
				{
					slot:        5,
					pandoraHash: nil,
					vanguardHash: &types.HeaderHash{
						HeaderHash: vanguardHash,
						Status:     types.Pending,
					},
					expectedPandoraStatus:  types.Skipped,
					expectedVanguardStatus: types.Skipped,
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
					expectedPandoraStatus:  types.Verified,
					expectedVanguardStatus: types.Verified,
				},
			}

			forkedUnorderedBatch := realmPairSuite{
				{
					slot: 5,
					pandoraHash: &types.HeaderHash{
						HeaderHash: pandoraHash,
						Status:     types.Pending,
					},
					vanguardHash: &types.HeaderHash{
						HeaderHash: vanguardHash,
						Status:     types.Pending,
					},
					expectedPandoraStatus:  types.Verified,
					expectedVanguardStatus: types.Verified,
				},
				{
					slot:                   3,
					pandoraHash:            nil,
					vanguardHash:           nil,
					expectedPandoraStatus:  types.Skipped,
					expectedVanguardStatus: types.Skipped,
				},
				{
					slot: 2,
					pandoraHash: &types.HeaderHash{
						HeaderHash: pandoraHash,
						Status:     types.Pending,
					},
					vanguardHash:           nil,
					expectedPandoraStatus:  types.Skipped,
					expectedVanguardStatus: types.Skipped,
				},
				{
					slot:        6,
					pandoraHash: nil,
					vanguardHash: &types.HeaderHash{
						HeaderHash: vanguardHash,
						Status:     types.Pending,
					},
					expectedPandoraStatus:  types.Skipped,
					expectedVanguardStatus: types.Skipped,
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
					expectedPandoraStatus:  types.Verified,
					expectedVanguardStatus: types.Verified,
				},
				{
					slot: 7,
					pandoraHash: &types.HeaderHash{
						HeaderHash: pandoraHash,
						Status:     types.Pending,
					},
					vanguardHash: &types.HeaderHash{
						HeaderHash: vanguardHash,
						Status:     types.Pending,
					},
					expectedPandoraStatus:  types.Verified,
					expectedVanguardStatus: types.Verified,
				},
			}

			failChan := make(chan error)

			testSuiteFunc := func(
				t *testing.T,
				testSuite realmPairSuite,
				waitGroup *sync.WaitGroup,
				suiteNumber int,
			) {
				waitGroup.Add(len(testSuite))

				for index, suite := range testSuite {
					if nil != suite.pandoraHash {
						require.NoError(
							t,
							orchestratorDB.SavePandoraHeaderHash(suite.slot, suite.pandoraHash),
							index,
						)
					}

					if nil != suite.vanguardHash {
						require.NoError(
							t,
							orchestratorDB.SaveVanguardHeaderHash(suite.slot, suite.vanguardHash),
							index,
						)
					}

					waitGroup.Done()
				}

				vanguardErr, pandoraErr, realmErr := backend.InvalidatePendingQueue()
				require.NoError(t, vanguardErr, suiteNumber)
				require.NoError(t, pandoraErr, suiteNumber)
				require.NoError(t, realmErr, suiteNumber)

				waitGroup.Done()
			}

			parallelTestSuites := make([]realmPairSuite, 3)
			parallelTestSuites[0] = firstBatch
			parallelTestSuites[1] = secondBatch
			parallelTestSuites[2] = forkedUnorderedBatch
			waitGroup := &sync.WaitGroup{}
			waitGroup.Add(len(parallelTestSuites))

			for testSuiteIndex, parallelTestSuite := range parallelTestSuites {
				time.Sleep(time.Millisecond * 25)
				go testSuiteFunc(t, parallelTestSuite, waitGroup, testSuiteIndex)
			}

			time.AfterFunc(time.Second*2, func() {
				failChan <- fmt.Errorf("timeout of test suite")
			})

			select {
			case err := <-failChan:
				require.NoError(t, err)
			default:
				waitGroup.Wait()
			}
		})
	}
}
