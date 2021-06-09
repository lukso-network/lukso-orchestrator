package consensus

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	testDB "github.com/lukso-network/lukso-orchestrator/orchestrator/db/testing"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"
)

const (
	mockedPandoraFile  = "./fixtures/mockedPandora.json"
	mockedVanguardFile = "./fixtures/mockedVanguard.json"
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

	// TODO: if we maintain crawling in batches we should test with batchLimit 32 (epoch)
	t.Run("should invalidate based on static set", func(t *testing.T) {
		var (
			vanguardBlocks []*types.HeaderHash
			pandoraBlocks  []*types.HeaderHash
		)

		mockedPandora, err := os.Open(mockedPandoraFile)
		require.NoError(t, err)
		mockedPandoraJson, err := ioutil.ReadAll(mockedPandora)
		require.NoError(t, err)

		mockedVanguard, err := os.Open(mockedVanguardFile)
		require.NoError(t, err)
		mockedVanguardJson, err := ioutil.ReadAll(mockedVanguard)
		require.NoError(t, err)

		require.NoError(t, json.Unmarshal(mockedPandoraJson, &pandoraBlocks))
		require.NoError(t, json.Unmarshal(mockedVanguardJson, &vanguardBlocks))

		orchestratorDB := testDB.SetupDB(t)
		ctx := context.Background()
		vanguardHeadersChan, vanguardConsensusInfoChan, pandoraHeadersChan := prepareEnv()
		service := New(ctx, orchestratorDB, vanguardHeadersChan, vanguardConsensusInfoChan, pandoraHeadersChan)

		for index, block := range pandoraBlocks {
			if nil != block {
				require.NoError(t, service.PandoraHeaderHashDB.SavePandoraHeaderHash(uint64(index), block))
			}

		}

		for index, block := range vanguardBlocks {
			if nil != block {
				require.NoError(t, service.VanguardHeaderHashDB.SaveVanguardHeaderHash(uint64(index), block))
			}
		}

		vanguardErr, pandoraErr, realmErr := service.Canonicalize(0, 500)
		require.NoError(t, vanguardErr)
		require.NoError(t, pandoraErr)
		require.NoError(t, realmErr)

		confirmValidityOfSlot := func() {
			vanguardHeaderHash, err := service.VanguardHeaderHashDB.VanguardHeaderHash(180)
			require.NoError(t, err)

			pandoraHeaderHash, err := service.PandoraHeaderHashDB.PandoraHeaderHash(180)
			require.NoError(t, err)

			// This will check integrity of data
			// If mock wont change below should be true
			require.Equal(
				t,
				"0x6c4b454db445110b4587a485a1ca080255731d05138fbd61d19281f664fcab6a",
				vanguardHeaderHash.HeaderHash.String(),
			)
			require.Equal(
				t,
				"0x92bdf4ea28129191715eac13327c37c3c55bfb9cccaaa5d3d6591a217cf2188f",
				pandoraHeaderHash.HeaderHash.String(),
			)
		}

		confirmValidityOfSlot()

		expectedLatestVerifiedRealmSlot := uint64(231)
		expectedFirstVerifiedSlot := 180
		require.Equal(t, expectedLatestVerifiedRealmSlot, service.RealmDB.LatestVerifiedRealmSlot())
		pandoraBlocksLen := len(pandoraBlocks)
		vanguardBlocksLen := len(vanguardBlocks)
		require.Equal(t, pandoraBlocksLen, vanguardBlocksLen)

		confirmValidityOfFirstBatch := func(newBatchCame bool) {
			// first verified slot should be 180, below that slots should be skipped
			for index := 0; index < vanguardBlocksLen; index++ {
				currentVanguardHeaderHash, err := service.VanguardHeaderHashDB.VanguardHeaderHash(uint64(index))
				require.NoError(t, err)

				currentPandoraHeaderHash, err := service.PandoraHeaderHashDB.PandoraHeaderHash(uint64(index))
				require.NoError(t, err)

				pandoraRelative := pandoraBlocks[index]
				vanguardRelative := vanguardBlocks[index]

				// No pending block were present until expectedFirstVerifiedSlot
				if index < expectedFirstVerifiedSlot && nil != currentVanguardHeaderHash && nil != currentPandoraHeaderHash {
					require.Equal(t, types.Skipped, currentVanguardHeaderHash.Status, index)
					require.Equal(t, types.Skipped, currentPandoraHeaderHash.Status, index)
				}

				// Present on both sides
				if nil != pandoraRelative && nil != vanguardRelative {
					require.Equal(t, types.Verified, currentVanguardHeaderHash.Status, index)
					require.Equal(t, types.Verified, currentPandoraHeaderHash.Status, index)
				}

				if !newBatchCame && index > int(expectedLatestVerifiedRealmSlot) && nil != currentVanguardHeaderHash {
					require.Equal(t, types.Pending, currentVanguardHeaderHash.Status, index)
				}

				if !newBatchCame && index > int(expectedLatestVerifiedRealmSlot) && nil != currentPandoraHeaderHash {
					require.Equal(t, types.Pending, currentPandoraHeaderHash.Status, index)
				}

				if index < int(expectedLatestVerifiedRealmSlot) && nil == pandoraRelative && nil != vanguardRelative {
					require.Equal(t, types.Skipped, currentPandoraHeaderHash.Status, index)
				}

				if index < int(expectedLatestVerifiedRealmSlot) && nil == vanguardRelative && nil != pandoraRelative && nil != currentVanguardHeaderHash {
					require.Equal(t, types.Skipped, currentVanguardHeaderHash.Status, index)
				}
			}
		}

		confirmValidityOfFirstBatch(false)

		// Save next batch and see if crawler can go up
		for index, block := range pandoraBlocks {
			if nil != block {
				require.NoError(t, service.PandoraHeaderHashDB.SavePandoraHeaderHash(
					uint64(index+pandoraBlocksLen),
					block,
				))
			}

		}

		for index, block := range vanguardBlocks {
			if nil != block {
				require.NoError(t, service.VanguardHeaderHashDB.SaveVanguardHeaderHash(
					uint64(index+vanguardBlocksLen),
					block,
				))
			}
		}

		vanguardErr, pandoraErr, realmErr = service.Canonicalize(expectedLatestVerifiedRealmSlot, 500)
		require.NoError(t, vanguardErr)
		require.NoError(t, pandoraErr)
		require.NoError(t, realmErr)

		// Test only crucial side effects
		expectedHighestCheckedSlot := uint64(466)
		require.Equal(t, expectedHighestCheckedSlot, service.RealmDB.LatestVerifiedRealmSlot())

		// There should be no reorg
		confirmValidityOfSlot()
		confirmValidityOfFirstBatch(true)

		lastVanguardBlock, err := service.VanguardHeaderHashDB.VanguardHeaderHash(expectedHighestCheckedSlot)
		require.NoError(t, err)
		require.Equal(t, types.Verified, lastVanguardBlock.Status)
		require.Equal(t,
			"0x078ed0e94e50738b567764f8587b76a0c0a6bef2fd4ac8f6665b55cddba055db",
			lastVanguardBlock.HeaderHash.String(),
		)

		lastPandoraBlock, err := service.PandoraHeaderHashDB.PandoraHeaderHash(expectedHighestCheckedSlot)
		require.NoError(t, err)
		require.Equal(t, types.Verified, lastPandoraBlock.Status)
		require.Equal(t,
			"0x9cea5d4952be0bae951b717f4c8257a225cf837fb5720ce57293606219c990fc",
			lastPandoraBlock.HeaderHash.String(),
		)
	})

	// This test is not finished, it does not test the side effect
	// TODO: check how it can crawl up to 15000
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

}
