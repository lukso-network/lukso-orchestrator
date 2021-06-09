package consensus

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	testDB "github.com/lukso-network/lukso-orchestrator/orchestrator/db/testing"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"math/rand"
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

func prepareEnv() (
	vanguardHeadersChan chan *types.HeaderHash,
	vanguardConsensusInfoChan chan *types.MinimalEpochConsensusInfo,
	pandoraHeadersChan chan *types.HeaderHash,
) {
	capacity := 50000
	vanguardHeadersChan = make(chan *types.HeaderHash, capacity)
	vanguardConsensusInfoChan = make(chan *types.MinimalEpochConsensusInfo, capacity)
	pandoraHeadersChan = make(chan *types.HeaderHash, capacity)

	return
}

func TestNew(t *testing.T) {
	orchestratorDB := testDB.SetupDB(t)
	ctx := context.Background()
	vanguardHeadersChan, vanguardConsensusInfoChan, pandoraHeadersChan := prepareEnv()
	service := New(ctx, orchestratorDB, vanguardHeadersChan, vanguardConsensusInfoChan, pandoraHeadersChan)
	require.Equal(t, orchestratorDB, service.VanguardHeaderHashDB)
	require.Equal(t, orchestratorDB, service.PandoraHeaderHashDB)
	require.Equal(t, orchestratorDB, service.RealmDB)
}

func TestService_Canonicalize(t *testing.T) {
	t.Run("should invalidate matched blocks on slot 1", func(t *testing.T) {
		orchestratorDB := testDB.SetupDB(t)
		ctx := context.Background()
		vanguardHeadersChan, vanguardConsensusInfoChan, pandoraHeadersChan := prepareEnv()
		service := New(ctx, orchestratorDB, vanguardHeadersChan, vanguardConsensusInfoChan, pandoraHeadersChan)

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

		vanguardErr, pandoraErr, realmErr := service.Canonicalize(0, 50)
		require.NoError(t, vanguardErr)
		require.NoError(t, pandoraErr)
		require.NoError(t, realmErr)

		headerHash, err := orchestratorDB.VanguardHeaderHash(2)
		require.NoError(t, err)
		require.Equal(t, types.Verified, headerHash.Status)
		require.Equal(t, vanguardHash, headerHash.HeaderHash)

		headerHash, err = orchestratorDB.PandoraHeaderHash(2)
		require.NoError(t, err)
		require.Equal(t, types.Verified, headerHash.Status)
		require.Equal(t, pandoraHash, headerHash.HeaderHash)

		realmSlot := orchestratorDB.LatestVerifiedRealmSlot()
		require.Equal(t, uint64(2), realmSlot)
	})

	t.Run("should handle skipped blocks", func(t *testing.T) {
		orchestratorDB := testDB.SetupDB(t)
		ctx := context.Background()
		vanguardHeadersChan, vanguardConsensusInfoChan, pandoraHeadersChan := prepareEnv()
		service := New(ctx, orchestratorDB, vanguardHeadersChan, vanguardConsensusInfoChan, pandoraHeadersChan)

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

		vanguardErr, pandoraErr, realmErr := service.Canonicalize(0, 50)
		require.NoError(t, vanguardErr)
		require.NoError(t, pandoraErr)
		require.NoError(t, realmErr)

		headerHash, err := orchestratorDB.VanguardHeaderHash(5)
		require.NoError(t, err)
		require.Equal(t, types.Verified, headerHash.Status)
		require.Equal(t, vanguardHash, headerHash.HeaderHash)

		headerHash, err = orchestratorDB.PandoraHeaderHash(5)
		require.NoError(t, err)
		require.Equal(t, types.Verified, headerHash.Status)
		require.Equal(t, pandoraHash, headerHash.HeaderHash)

		realmSlot := orchestratorDB.LatestVerifiedRealmSlot()
		require.Equal(t, uint64(5), realmSlot)
	})

	t.Run("should handle skipped blocks with different states", func(t *testing.T) {
		orchestratorDB := testDB.SetupDB(t)
		ctx := context.Background()
		vanguardHeadersChan, vanguardConsensusInfoChan, pandoraHeadersChan := prepareEnv()
		service := New(ctx, orchestratorDB, vanguardHeadersChan, vanguardConsensusInfoChan, pandoraHeadersChan)

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

		vanguardErr, pandoraErr, realmErr := service.Canonicalize(0, 600)
		require.NoError(t, vanguardErr)
		require.NoError(t, pandoraErr)
		require.NoError(t, realmErr)

		for index, suite := range testSuite {
			indexMsg := fmt.Sprintf("Failed on slot: %d, index: %d", suite.slot, index)

			if nil != suite.vanguardHash {
				headerHash, err := orchestratorDB.VanguardHeaderHash(suite.slot)
				require.NoError(t, err, indexMsg)
				require.Equal(t, suite.expectedVanguardStatus, headerHash.Status, indexMsg)
			}

			if nil != suite.pandoraHash {
				headerHash, err := orchestratorDB.PandoraHeaderHash(suite.slot)
				require.NoError(t, err, indexMsg)
				require.Equal(t, suite.expectedPandoraStatus, headerHash.Status, indexMsg)
			}
		}

		realmSlot := service.RealmDB.LatestVerifiedRealmSlot()
		require.Equal(t, uint64(6), realmSlot)
	})

	t.Run("should invalidate lots of pending blocks", func(t *testing.T) {
		orchestratorDB := testDB.SetupDB(t)
		ctx := context.Background()
		vanguardHeadersChan, vanguardConsensusInfoChan, pandoraHeadersChan := prepareEnv()
		service := New(ctx, orchestratorDB, vanguardHeadersChan, vanguardConsensusInfoChan, pandoraHeadersChan)

		type realmPair struct {
			pandoraSlot  uint64
			pandoraHash  common.Hash
			vanguardSlot uint64
			vanguardHash common.Hash
		}

		batchSize := 1550
		pendingBatch := make([]*realmPair, batchSize)

		// It will fill batch with random order
		// This will simulate network traffic from both parties for existing network
		t.Log("Started process of hash and slot creation")
		rand.Seed(time.Now().UnixNano())
		for index := 0; index < len(pendingBatch); index++ {
			min := 1
			max := batchSize - 1
			pandoraSlot := rand.Intn(max-min+1) + min

			pandoraToken := make([]byte, 4)
			rand.Read(pandoraToken)
			pandoraHash := common.BytesToHash(pandoraToken)

			vanguardToken := make([]byte, 8)
			rand.Read(vanguardToken)
			vanguardHash := common.BytesToHash(vanguardToken)

			vanguardSlot := rand.Intn(max-min+1) + min
			pandoraNotPresentSlot := rand.Intn(max-min+1) + min
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
				vanguardSlot: uint64(pandoraNotPresentSlot),
				vanguardHash: vanguardHash,
			}

			pendingBatch[vanguardNotPresentSlot] = &realmPair{
				pandoraSlot: uint64(vanguardNotPresentSlot),
				pandoraHash: vanguardHash,
			}
		}

		t.Log("Starting process of database save")

		for index, pair := range pendingBatch {
			if nil == pair {
				t.Logf("I am skipping nil pendingBatch: %d", index)
				continue
			}

			pandoraHash := &types.HeaderHash{
				HeaderHash: pair.pandoraHash,
				Status:     types.Pending,
			}

			vanguardHash := &types.HeaderHash{
				HeaderHash: pair.vanguardHash,
				Status:     types.Pending,
			}

			if pair.pandoraSlot > 0 {
				require.NoError(
					t,
					orchestratorDB.SavePandoraHeaderHash(
						pair.pandoraSlot,
						pandoraHash,
					),
					index,
				)
			}

			if pair.vanguardSlot > 0 {
				require.NoError(
					t,
					orchestratorDB.SaveVanguardHeaderHash(
						pair.vanguardSlot,
						vanguardHash,
					),
					index,
				)
			}
		}

		t.Log("Starting process of invalidation")

		invalidationTicker := time.NewTicker(time.Millisecond * 50)
		// This timeout can be random within a range
		invalidationTimeOut := time.NewTicker(time.Second * 20)

		for {
			select {
			case <-invalidationTicker.C:
				vanguardErr, pandoraErr, realmErr := service.Canonicalize(
					orchestratorDB.LatestVerifiedRealmSlot(),
					500,
				)
				require.NoError(t, vanguardErr)
				require.NoError(t, pandoraErr)
				require.NoError(t, realmErr)
			case <-invalidationTimeOut.C:
				t.Log("I have reached the test end")
				// Here I should test the side effect of invalidation

				return
			}
		}
	})

	// What if pandora is broken and vanguard is ok, should we miss the blocks and discard them?
	// Maybe something like isSyncing?

	t.Run("should invalidate again when there was a hole of missing blocks", func(t *testing.T) {

	})
}
