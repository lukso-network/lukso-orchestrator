package consensus

import (
	"context"
	"encoding/json"
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

	// TODO: if we maintain crawling in batches we should test with batchLimit 32 (epoch)
	t.Run("should invalidate based on static set", func(t *testing.T) {
		var (
			vanguardBlocks []*types.HeaderHash
			pandoraBlocks  []*types.HeaderHash
		)

		require.NoError(t, json.Unmarshal([]byte(mockedPandoraJson), &pandoraBlocks))
		require.NoError(t, json.Unmarshal([]byte(mockedVanguardJson), &vanguardBlocks))

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
				if index < expectedFirstVerifiedSlot {
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

				if index < int(expectedLatestVerifiedRealmSlot) && nil == vanguardRelative && nil != pandoraRelative {
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

		for index := expectedLatestVerifiedRealmSlot; index < expectedHighestCheckedSlot; index++ {

		}
	})
}

const (
	mockedPandoraJson = `[
      null,
      {
        "headerHash": "0xde7a77f9563a51ef5879aeb2387659f07220c7dbffc75da8dada6ec4dbcdf3bc",
        "status": 0
      },
      {
        "headerHash": "0xc54e0c19c807b685944fd92308c311d1cbe106b80d62415d39ce68c6f2d9f689",
        "status": 0
      },
      {
        "headerHash": "0xdced7fe7f181d8c73ed1744335a3f6485452ad2903f20ff3e745371ffac3907f",
        "status": 0
      },
      {
        "headerHash": "0xbc6fd05b6a4cb0e9f445a25823955ffa4564573acb50ee70adcedb181e0a97c4",
        "status": 0
      },
      {
        "headerHash": "0x6ad2078ec8ae8335e1e33febfcd1be0308edb380cbf69be860826d131fc48c5d",
        "status": 0
      },
      {
        "headerHash": "0xf81e63f10eeba198d05523461da5391b3c099ffa745f325a4d4f4afde152d949",
        "status": 0
      },
      {
        "headerHash": "0xcb05d7142221400fa7a812cca5e126e51836feabe0f13f2f487e5b6894be1c08",
        "status": 0
      },
      {
        "headerHash": "0xd49a6acdc8e3d6bf681eb60dc182e3748e8be1c1376e9e141f4c1055b1d72634",
        "status": 0
      },
      {
        "headerHash": "0x3949bdba49164e9223706344923e2cad71ee7752f464c19b7c825af18473750f",
        "status": 0
      },
      {
        "headerHash": "0xef40e674cc7fa3e20f7aec3df5260e15d524ecc4806bd816225accbf91bb0cd8",
        "status": 0
      },
      {
        "headerHash": "0xea3c12b3ae75ce377e00033b709c1a64b128eb2ed045762d4d439b71d3212dcc",
        "status": 0
      },
      {
        "headerHash": "0x11f0e2f7e1333e82743c0ad6f620666a352dfdcec3b65fa2e542267f12baa742",
        "status": 0
      },
      {
        "headerHash": "0x6da19137743a7b8f03d8e36ddab9575b668470f8d26e71ec79bbc459fcd92bb8",
        "status": 0
      },
      {
        "headerHash": "0x21d47d9fda3b9797371ae2a34f1973dea145bbc8d292c1b5f4d10d9abe077127",
        "status": 0
      },
      {
        "headerHash": "0x4bad3bd9949ad7c68f48346123371ff7373ba1f8a66fa2111f530ff18e2aae19",
        "status": 0
      },
      {
        "headerHash": "0x856e46efbfac09da9f1cb0658872ddb5aca779c7ffd9b7c5b3b97307070781c8",
        "status": 0
      },
      null,
      {
        "headerHash": "0x59565f0be4d2a676c0ff543f8ab46939b11b56e86ff22c14e16e7bde18aa23e4",
        "status": 0
      },
      {
        "headerHash": "0x77af712c107947a831d2da1fd3dbdf34e327bf06ee019d9bc2da62d00c131a42",
        "status": 0
      },
      {
        "headerHash": "0x6746119f3c9fe35c14d5bee420a526fd4c159daf030759c3e4ac8c28c4569169",
        "status": 0
      },
      null,
      {
        "headerHash": "0xb440693baa21e55e51db9c79ff5da698bfbe96ad94dab839209d917d3bba04a3",
        "status": 0
      },
      {
        "headerHash": "0x5f59bb3926635241ccd73b17f6b7bea66655a89f0a6397ba82e2f562e51adefa",
        "status": 0
      },
      {
        "headerHash": "0x3ce4d6b57d0260171129c068a2ce9bbe57eb584cd2bcf851763964f7e9676646",
        "status": 0
      },
      {
        "headerHash": "0x59e80683e62e34945454fa4d060b29cfb01ae30abcc86d727d174a745f2806d5",
        "status": 0
      },
      {
        "headerHash": "0xc2b1c95fe157e7679aea2deb2ab2b510269e9f1d1f3e2a6e68b00c70ce183e65",
        "status": 0
      },
      {
        "headerHash": "0x0c9eb531a0ff8c85467f1a5e32eff00b77cfdd9b31e5a72211b814fc00d35523",
        "status": 0
      },
      {
        "headerHash": "0xaaea8ed2d8a16a00571fecd2633a8b93a545341832e6a177b8eb697c865ef3a4",
        "status": 0
      },
      {
        "headerHash": "0xf9d80dd0bec93f3c2ac8c5718ffd1a431c8152330959d652c19282c17e2e193a",
        "status": 0
      },
      {
        "headerHash": "0xd6208a8c714562edb195522ce66e5bb59854159cefaded98502811bd1086765f",
        "status": 0
      },
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      {
        "headerHash": "0x87830481e2e864cab3715abe28db909f0470687adb0af5ca19bbe2e44161e078",
        "status": 0
      },
      {
        "headerHash": "0x52ff95231316ebd1335e9cff7dc32c10d0539f6950c757de67ad7152c4dcfb37",
        "status": 0
      },
      {
        "headerHash": "0x4c048599c0889df1762fc97eded2c11659fc19c11ce371c9f7892cfe92e75c96",
        "status": 0
      },
      {
        "headerHash": "0x783181a36462cfbc98a6f2e7384dadf8c6bd51d1ab9fa2e91fc2a3b4983782a9",
        "status": 0
      },
      {
        "headerHash": "0x8288c1005b65d1ded14bd576672e10393b84b2b29eabb90e8d2989c7c1e94373",
        "status": 0
      },
      {
        "headerHash": "0x01c3fde62acd2ee94ccd9b20d74a67b69793259a0c2d5c1afd01ccbb9c7d18c9",
        "status": 0
      },
      {
        "headerHash": "0xf5c91a1c060ccee79ca2c031023b0f0cf0e9e6b52ae373a5c45e670ff8eb6d47",
        "status": 0
      },
      {
        "headerHash": "0x4aa9db3fe8bd2c0a3f6a6683d514d14e4a9fc526687b3ab5db329265467bad5d",
        "status": 0
      },
      null,
      {
        "headerHash": "0xf1ce9992b02793ec8873435e1d9182c6530b695502238971339efc8ed17fbcf9",
        "status": 0
      },
      {
        "headerHash": "0x5a7fd70b0d227d876812c67d3dd6c229eb8568eab6d66263c73273ba2c61083d",
        "status": 0
      },
      {
        "headerHash": "0xa21f5d962026b48440fcc6092f7f06066bbd534109dbb606c37a94e161df97a3",
        "status": 0
      },
      {
        "headerHash": "0x25b8f1a61ee1e38aca8e528e2aea7f022131c32925f504e7303c39be6fb7cda2",
        "status": 0
      },
      {
        "headerHash": "0xe3c7233ea8d31fe2046e6bfe3dd2a08e1ad05412df6b36c14e0a14ac7fb86aa6",
        "status": 0
      },
      {
        "headerHash": "0xac7fa567c747df0436c119b9bf68e9ac46b79af3a8b697acdac08c99fd9ef805",
        "status": 0
      },
      {
        "headerHash": "0x4425491daaa099d3a04a1ba0fc28da68c3d56c5b3cd5ddec22778d94f8bfb6b0",
        "status": 0
      },
      {
        "headerHash": "0x4473fd4a14a4a6dc30e3a8baf107ff7cee34b9270031228dd88c4041481468af",
        "status": 0
      },
      {
        "headerHash": "0x06a0de99334f45d8faaaabc14194bec09606b502da786d50c0f40fa58d4c182f",
        "status": 0
      },
      {
        "headerHash": "0x872f2e95dd245e993ba8a338e800c25ac341605399aeb01ab90ac0be26b6ad19",
        "status": 0
      },
      {
        "headerHash": "0x3f408a49bfdc3efb65b4343404b669ffb2d97aa9d9d7c59a4d2878726812100e",
        "status": 0
      },
      {
        "headerHash": "0xd387f5bae09875f3065a996ab609f00a334e9123c8d8cd8e5ee0c40dacc3698e",
        "status": 0
      },
      {
        "headerHash": "0xebb5fe0564a9b529c2ec80bf0051f5dc67d22bcf323cb904b1076fae93ad5e8a",
        "status": 0
      },
      null,
      null,
      {
        "headerHash": "0x53f89dfee3e13569c2332681883756d3bfafb8719521c6775255b08e025542b2",
        "status": 0
      },
      {
        "headerHash": "0xe4dccd3fde1f35aad1849447aab1f3e56e050c3acc36f98dbdcd35e83e3bdf8d",
        "status": 0
      },
      {
        "headerHash": "0x188365b5af52bc5cbc0081c2a4065386c4838d34449da2bd22e94a2be22f9451",
        "status": 0
      },
      null,
      null,
      {
        "headerHash": "0x98d3fd9dbb9463ac2db7fb9fdb013d122a5c16ceb914de0fbc304f3d9648a5d5",
        "status": 0
      },
      {
        "headerHash": "0x94b40adcec44c1567346ebace2658a8dc5be7888a17590148e1c6a9ee65779c7",
        "status": 0
      },
      {
        "headerHash": "0xd7c821274076745c26fa48d270ed284412af75999db734c915aad2118bef5d62",
        "status": 0
      },
      {
        "headerHash": "0x364bd8c93c740554b6e488347c29fbd96a8981fada53b9e7515a4e5494e28230",
        "status": 0
      },
      {
        "headerHash": "0x3276a179fb7b4e19e6d38f2bc7f2e1d7a77794be5cdff1d8944f301a2f5c93e8",
        "status": 0
      },
      {
        "headerHash": "0x0b3c2405475fa15ed7fe59d5f1e9e0e5fc6e71c02fc4f32d630663d39863f7f5",
        "status": 0
      },
      {
        "headerHash": "0xaf7a97b4f0f90069cb1d7d861e6ce87ae7b1458b920bd28fefe574912d801b11",
        "status": 0
      },
      {
        "headerHash": "0xeac340b5eda0013a3bb19dd5e82ed51d7b80fed8d3177e878d4ada5bad5936a0",
        "status": 0
      },
      {
        "headerHash": "0x790fb27cfa60cf8c36b985bd19b5643f469ee5dc6e22cbf3eca796d53874d870",
        "status": 0
      },
      {
        "headerHash": "0x9eb5f1fc3f96e3ce47702b512177ec07ac794e187cfe18d6d0236b7ea3688038",
        "status": 0
      },
      {
        "headerHash": "0xf171d2039bf92239418b113e14b9946abc39f260c0874588020df567090484b4",
        "status": 0
      },
      {
        "headerHash": "0x547792b36e5939407f34f66a1d33aef3d2b80d135fd8a6adc46178aaafe75019",
        "status": 0
      },
      {
        "headerHash": "0x4b1c3e00410ca145436beed9de1b0d4e67b985e831786f6a631996d1737fe071",
        "status": 0
      },
      null,
      {
        "headerHash": "0x3c1faf487c7af0f78572e9c9b41e64594e7295a1fc0a0ecc78aea97709c72ec9",
        "status": 0
      },
      {
        "headerHash": "0xc64c2732c88df57c51974137334e5e7815462a7c1c9d584417197ff00287b924",
        "status": 0
      },
      {
        "headerHash": "0xedb49110a4fc7d95f3fa3f53413b00a2c47c077a55ae45a9a99e4c4678a45792",
        "status": 0
      },
      {
        "headerHash": "0xcb8ed373779155d64ff40f64011dcc40c6f36b06c4922e2efcf3c2c29da1bd5e",
        "status": 0
      },
      {
        "headerHash": "0xf7388fc5e485dca1833457d14ef87eb9cd77515fc2427bcc07df13f73c44dd72",
        "status": 0
      },
      null,
      {
        "headerHash": "0xc63c7a36ba510097ce1d815ef1c26d15c95001aed1afae5baf265d12b752d962",
        "status": 0
      },
      {
        "headerHash": "0xa72ea48aa30a9ed447e9a815ff07429735b1f778c690ed1485b4b6b9c4409cbc",
        "status": 0
      },
      {
        "headerHash": "0x882ef3ff35df4bddc966c8ecfba39618d8739036116f729506e86db99c5837e6",
        "status": 0
      },
      null,
      {
        "headerHash": "0x5f9ab3521a521a65524bb13c824ef40a182380821574550a2409e4face03f47e",
        "status": 0
      },
      null,
      null,
      {
        "headerHash": "0xd82dc076d8aead88d89c2dd45d934be6006f8b918371f80a2893d27c4b110d17",
        "status": 0
      },
      {
        "headerHash": "0xcd198e45b60cd3285267112ef035f22b2b1cf2951556da20bd377df9fa7f51dd",
        "status": 0
      },
      {
        "headerHash": "0x239050796be40f19290d2b777cc2ad31b4153316f133dc6f45f8217367ca0d90",
        "status": 0
      },
      {
        "headerHash": "0xad1af73ba42874274ff5e981f84e76bef692e67850432e006bc98e99f12c04bb",
        "status": 0
      },
      {
        "headerHash": "0xfa32061a09a763a02e4c39f562f1ea9fa501b1e73c5125c138e671c1cdc3ae6b",
        "status": 0
      },
      {
        "headerHash": "0x545b9aa05443097463c2077104a2e4e102021bd6eeebda51de16357cf66fc3c9",
        "status": 0
      },
      {
        "headerHash": "0xf8fecf67230dda92b026c5f5b42f8ff7a63ab70708de6351d4d59913162ec61c",
        "status": 0
      },
      null,
      {
        "headerHash": "0x69122f03ad1cbdfbe4104a2eb8d7aa9c3f552fe2cba049804c01559f8385a109",
        "status": 0
      },
      {
        "headerHash": "0x862875a197d5b3e2af167b46a2af4068d602fe9bd566db5a01bd63ad470b4ede",
        "status": 0
      },
      {
        "headerHash": "0x9f75ba479bcffe1b4659f48e6a123f9c1543098fe7bb0191d798c3817a0cd667",
        "status": 0
      },
      {
        "headerHash": "0x3db397658e9b08d3d8e8f1ff98f6b2d074756b8f770dbe8b6d8f471d7620adc8",
        "status": 0
      },
      null,
      {
        "headerHash": "0xbcbab1c6a259d07cd2a8e9bd262105d11572a4287f9df71ed259b9c1cc5683a5",
        "status": 0
      },
      {
        "headerHash": "0xf8915e9a6c5116ccecb5df04b634ffe843af94f638f80fe9823c27955260f9d1",
        "status": 0
      },
      {
        "headerHash": "0x83564ce3216ba89d571bef42e5eb71546912aa2b079ef184795635136ff02e19",
        "status": 0
      },
      null,
      {
        "headerHash": "0x1f83de5708798733343ca1817960ebcc08fabf732d325e1a32603b10d6ac09f1",
        "status": 0
      },
      {
        "headerHash": "0x8e1d31ebe2b05f4061d84cb6a50a0a6e71d790d4227b429175210f98f237ecff",
        "status": 0
      },
      {
        "headerHash": "0xce9bf2cba5a2f0ec00bf1d413cbde33a89ecdec3d027b40322a1a902a932022a",
        "status": 0
      },
      {
        "headerHash": "0x8838e3a7d29316461a5943c862b6a45092a1faf53675b5f5b431e2431dd4d4e9",
        "status": 0
      },
      null,
      {
        "headerHash": "0x7d0029a2b7af24d6d78f4355e9ce85a5c76f9213e20677907a9802018ed01a9a",
        "status": 0
      },
      {
        "headerHash": "0x8427de14d893f4459c063afa2c95bf40e9b4abb343da9b9bf50b15d43ef76e3d",
        "status": 0
      },
      {
        "headerHash": "0x6f3cdfc8143fb3f01cbd0a1c1c944c2b96dc3c28c992080a045acad8c53f99dd",
        "status": 0
      },
      {
        "headerHash": "0x3f03287d5e34a04f02e970263e1079470a43b6d2bbde9731bb37e50059f6c04d",
        "status": 0
      },
      {
        "headerHash": "0xa9b15109f97d96caf60e81c65bae5d5390dfb1065b5fd570a563266a8ecb0230",
        "status": 0
      },
      {
        "headerHash": "0x7782ecefd711aacbe3edd79adec75ac50e7c16368a5156ed01dbfa0b9d730caf",
        "status": 0
      },
      {
        "headerHash": "0x6644729b0a67c01a2ccf24640923291d63d2e4a11a898348f61b6858731c36af",
        "status": 0
      },
      {
        "headerHash": "0x0e163dc19aa182ad988422a1fe518e6bb76010bf1388699b6a2103ffcb360f08",
        "status": 0
      },
      null,
      null,
      {
        "headerHash": "0x31c52b7b68cb3c7a9c002f2ebc74df2eafffe60ded2c326853b407957b121811",
        "status": 0
      },
      {
        "headerHash": "0xcd3cd9bbc2dee226eea86559c7f9632dedc9ac26497afd070991417be3300e61",
        "status": 0
      },
      {
        "headerHash": "0x48155d47fff1f8e419d27a706652913062cdb6e0dff1c5920d1d7fe7a2df5237",
        "status": 0
      },
      {
        "headerHash": "0x06dba531d9ff4a18d717763b40c91190c76bcf849ac42789130313ad15cd9654",
        "status": 0
      },
      {
        "headerHash": "0xa508fdbbdc49484198ac9361e265b7842ef6380e9fe80d16e8897cc96a915565",
        "status": 0
      },
      null,
      {
        "headerHash": "0x3c9ad9b5d57a15a1903c11f49e77aa396422576cebf19c9e6fef3e130cf4b89e",
        "status": 0
      },
      {
        "headerHash": "0xa0f68bbf00ca18a5221c42c1e1b20553cfa390d417317608529fb072b6d3bb96",
        "status": 0
      },
      {
        "headerHash": "0xf232b2cbd46a420171fdcfcd1ce78bd7ac6471248ffcf2d4dbe3b5c8cfed01b1",
        "status": 0
      },
      {
        "headerHash": "0xca32f189d2b13ebcb73f7b29d0912ba7e7324c6bb92d1f0969768c7068b2272c",
        "status": 0
      },
      null,
      {
        "headerHash": "0x42bbfa7adf0c19f4faa59003dcaac54591b81799146a0d1a8260f6989708bb6e",
        "status": 0
      },
      {
        "headerHash": "0x1992fd73b1bd0501ed073462ef6ee04c6aa0c4e1d00e2ce005a882d2be2b5924",
        "status": 0
      },
      {
        "headerHash": "0xf9b561423ffb9c396028a55b8f6b08f4c7b54417105646f445e078890f01db64",
        "status": 0
      },
      {
        "headerHash": "0x406a3f796087489eadb78aa8561a1d036d481e00f05b70e2fbe06daec32a08b4",
        "status": 0
      },
      {
        "headerHash": "0x60b6abb398f40252aa7d3e64ea3de6c234c7fb41e441e38adc1b8290d26707a5",
        "status": 0
      },
      {
        "headerHash": "0x8d2069049223769ced4c4ab9abd37ef3577a6f7a52908463355371d4bd54831e",
        "status": 0
      },
      {
        "headerHash": "0xadd3420d7d21f5b1fd2b0089fb40037c597834bc2668a3ab0f71b0bd6d7cd777",
        "status": 0
      },
      {
        "headerHash": "0xfcf8354e68d0111d38a5f20edaa869705fbdc8416c14c0be15b57cc5f6221541",
        "status": 0
      },
      null,
      {
        "headerHash": "0x06c633a0b8612eeff249b1bb112be5a593036dd2ec02dff0af0d6f7a67b76a1e",
        "status": 0
      },
      {
        "headerHash": "0xb636561fc6cc7b98eb88d33519f21181faf1281195d872ff6b8bf000d2eacdca",
        "status": 0
      },
      {
        "headerHash": "0xc5ca96421fab6e4f665a14e9acafc8dcdd646af316595badb020761c59ea8ebf",
        "status": 0
      },
      {
        "headerHash": "0x1407f5869aac841713b46f86647b47220c9b33c4023bf798f93de9405d31dc9a",
        "status": 0
      },
      {
        "headerHash": "0x92944328cfb8d12351b0b2551f578634725ba635963adb27ce5cc017fb603f2e",
        "status": 0
      },
      {
        "headerHash": "0x0d60e02b160b71d416fe64bf6dd743eb4e571181fdedeb22985a2b0532d3e3ed",
        "status": 0
      },
      {
        "headerHash": "0xc74706b16b45d3fbf0fd5764c459998d4a23fb7af08257ae6d3ba7fe9cf0c5c0",
        "status": 0
      },
      {
        "headerHash": "0x95823059caa449258d77902943fb0b2fce8a9551c18b9094ff0735988f5e4c6b",
        "status": 0
      },
      null,
      {
        "headerHash": "0xc1275c6fb0169d442967d871a03296ea122e1f62d0cf13b08801022a53cd611a",
        "status": 0
      },
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      {
        "headerHash": "0x0a297624494e342958703412c586f204457d8adfaef6b7610505282515434878",
        "status": 0
      },
      null,
      {
        "headerHash": "0x11a7c55ad28676f0db64026f5a72539f96dd9d717701274885bacd80acd85daf",
        "status": 0
      },
      {
        "headerHash": "0xea66e5236e6531f857dbcc6172186087e004845abb8159474f9eeaf7c7cad84e",
        "status": 0
      },
      {
        "headerHash": "0x1b90a68fcd1aceaa6d464957b207e3aa235c08f80344e7ac1e607cf840dcd9a7",
        "status": 0
      },
      {
        "headerHash": "0xc124039babb3b0bb970e116a27573bfb3dd96a035d3d990ff43196497e449d36",
        "status": 0
      },
      {
        "headerHash": "0xec1b6aeee86247e6053f3aef520c6d2fddfe2f4ad62c427afdfcd8b5facb9d74",
        "status": 0
      },
      {
        "headerHash": "0xcf73340ddf35adcb41e57d1ef53bd358b52723c6a1e452f577176ed323e9f3bd",
        "status": 0
      },
      {
        "headerHash": "0xc79620fa38f4d4dd0b2d7878df6d0d9b3c7f49f75be2ecaee3f4748e8cef1ac6",
        "status": 0
      },
      {
        "headerHash": "0xf37847f384683cc1d82ca43376d93b3066178cf84acc7e97c2c638de09b529f6",
        "status": 0
      },
      null,
      {
        "headerHash": "0x92bdf4ea28129191715eac13327c37c3c55bfb9cccaaa5d3d6591a217cf2188f",
        "status": 0
      },
      {
        "headerHash": "0xb087b73fa196e3412640b78dbbe178792716be3f23653dbe753120cd939136ce",
        "status": 0
      },
      {
        "headerHash": "0xce6a31904973be61dd94a9780899c3346b544785953e13e95a7d95c043b00cf8",
        "status": 0
      },
      {
        "headerHash": "0x1e8c9157d6e48f1ecd8c210268a607d602190c52fd228f2b6c15ee25fb2e0a76",
        "status": 0
      },
      {
        "headerHash": "0x42e62ec91a7ad517553302b0524c434f678f3b9a707146404d61d97194a391cb",
        "status": 0
      },
      {
        "headerHash": "0xf07cbf99faf2121a8cb8e94e999d94317e72f53c7571bce2a8fa89344fa94860",
        "status": 0
      },
      {
        "headerHash": "0x21002916e86ff2b17508e6ff0ada0133e231b72c876f5fad06e27545e2c2b75e",
        "status": 0
      },
      {
        "headerHash": "0xff9ed2b2cf2c5e901d96955b7a18ce3cc1d82f16077877c5f5f8d0726d62084b",
        "status": 0
      },
      {
        "headerHash": "0xec2835ec4d95663ebf844a6fec69d16841335e9ac9464abdbaa7983674e8ced0",
        "status": 0
      },
      {
        "headerHash": "0xd383e9cbfdd37ff2107cab0caf0d1d99fd9001da56fbf2edd8468080067e01ea",
        "status": 0
      },
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      {
        "headerHash": "0x974613392f2c17295b495c1f72626814493b14bf1bab007c179784dcd610f83b",
        "status": 0
      },
      {
        "headerHash": "0x838663e93603ee9f7df03449a221a3726fda2e71e0dd9eb54cd50e9935ef756a",
        "status": 0
      },
      {
        "headerHash": "0x559d2b7c7bc99bd9e0ba62685524aca68cc6e4420aeeeaad7f589b7c4c9632f8",
        "status": 0
      },
      {
        "headerHash": "0x1425f11f47b90d2195af8c7b13fb74dbb0dcfef10cf7b898d3319a3c3b32d377",
        "status": 0
      },
      {
        "headerHash": "0x332a6bc32b64c9b4e4cb41b63be585c69fa9c24ee93d5c5447855f3c3118a401",
        "status": 0
      },
      {
        "headerHash": "0xb38bcbba4a9ebfefe36a3e54610a7eecae5480428232f88eaf3be903f3e33f4e",
        "status": 0
      },
      {
        "headerHash": "0x16f4f71699499ae6e6bdd0cc3d870aa92afc2638b7326ce2a1c8edfca8f169b1",
        "status": 0
      },
      null,
      {
        "headerHash": "0x4bf6ac6874f220beee6db0b02e7e68ac46b5cc5805b8aaf8318bf62081c36b47",
        "status": 0
      },
      {
        "headerHash": "0xd3c0c761e2b8b3760441a4b52fa22c6df243c56a7aaf609eb39994eead2d0268",
        "status": 0
      },
      {
        "headerHash": "0xadb3aeeb803e835850331e9f56c71dcbb6f6849d28d7a966f84a360d5ceae7bc",
        "status": 0
      },
      {
        "headerHash": "0xef5f766648c63e7420a85a4e4f2a9fbb502a112ad49b6474890976d5f49859c9",
        "status": 0
      },
      {
        "headerHash": "0x1b4aae895e92141e5f70806733b2a38ad41c8e3f54dead95d3ec4631f1ad4d3c",
        "status": 0
      },
      {
        "headerHash": "0xfcce7c881258465283f696809bea97698f809d0165ed4acca63ef1f237f3d2af",
        "status": 0
      },
      {
        "headerHash": "0xdc20e033a6371ce3178fea10067813cc74847d8cb76b5dbcf001e5d02b36888a",
        "status": 0
      },
      {
        "headerHash": "0x920606ef5fea8c2523da27e88f9faebaffffcb6434667270aa9601e3a7e8f624",
        "status": 0
      },
      null,
      {
        "headerHash": "0x122ad4243299550a2db4a4bec5a528c7b54bab610568231c5b2aaa490518b285",
        "status": 0
      },
      {
        "headerHash": "0x75dc81b04b567869b995c97c403ee39d0fc1259b6c6d10b5dd7955fb47ae9c71",
        "status": 0
      },
      {
        "headerHash": "0x0c07ffa4b28f150f07a82fb19485a8541f736103ea4e69ce8295036ba828fb72",
        "status": 0
      },
      {
        "headerHash": "0x6fb0c553ca292ad3293657e193ac29d0ec9e8266ea73835b480a8e0fe3e06639",
        "status": 0
      },
      {
        "headerHash": "0x4c14ff548d72f297b750e156b7448159a88dbe420e860fddd97c53f9e9a8afa7",
        "status": 0
      },
      {
        "headerHash": "0xfb75d99e33379420de35840285940589ec2da23dddbe2d4f4b3c59f1629c935d",
        "status": 0
      },
      {
        "headerHash": "0x90b95e21202267f70fb946520315e1cf89d249f8c4227672469cb44ea387ffcb",
        "status": 0
      },
      {
        "headerHash": "0xf06a26cf674773fbd60cb3a2e3711e7603112638df001584709f830c72365d4b",
        "status": 0
      },
      null,
      {
        "headerHash": "0x92cc84f7bb68a867a82aea5d2db808229812d3004a87f3705a99258503f670eb",
        "status": 0
      },
      null,
      null,
      {
        "headerHash": "0xfa3aea444126d76d30444ced2b1b3cfd18f97d8f540b6c42e7adee925b735955",
        "status": 0
      },
      {
        "headerHash": "0xa1dd5a854d642068a4a497601f7bf44242ea5e01025e5164cd1549d95e01efac",
        "status": 0
      },
      {
        "headerHash": "0x7639fe3a544fb2cc4e1fa3746d8739ee51b8ff03ec98cc778595c85b6d26266a",
        "status": 0
      },
      null,
      {
        "headerHash": "0x1a120b71c104b5984f93005573864b2f529527120f0f8c29d9111d6974026b4f",
        "status": 0
      },
      {
        "headerHash": "0x9cea5d4952be0bae951b717f4c8257a225cf837fb5720ce57293606219c990fc",
        "status": 0
      },
      null,
      {
        "headerHash": "0x9cea5d4952be0bae951b717f4c8257a225cf837fb5720ce57293606219c990fc",
        "status": 0
      },
      null
    ]
`
	mockedVanguardJson = `[
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      {
        "headerHash": "0x6c4b454db445110b4587a485a1ca080255731d05138fbd61d19281f664fcab6a",
        "status": 0
      },
      {
        "headerHash": "0x3a2ca006dfe1aa24ee85dc5c74ff89955bf81a84020bbae5493b35c21f234e47",
        "status": 0
      },
      {
        "headerHash": "0x9cf63d51b7a5f82652401afbda0e22a8b0a0056125e235047d3f671154ebf7b8",
        "status": 0
      },
      {
        "headerHash": "0x8ec594eb28960e779fd4c502c581e53d07f9045b656e390503086707950ebca5",
        "status": 0
      },
      {
        "headerHash": "0x2b68a910c5ae251c8733cbcbac76fb1714b7e4f5076685d5e22e1cc8cf43634d",
        "status": 0
      },
      {
        "headerHash": "0x48726637195a9770b39f4ec5796a92f926da23b99e33cdf420ac58a1e383206c",
        "status": 0
      },
      {
        "headerHash": "0x23de832ad03b1e170836efc75703363e6382100b4c1e64dd544225c3b79a0459",
        "status": 0
      },
      {
        "headerHash": "0xbe9554bc46ca3cd1bb05e9512d89fb7680fae9cfcdcd14e265f570482b65d341",
        "status": 0
      },
      {
        "headerHash": "0x72eb04446f1000d4868283014dae9b01153aa80b7e383ce4ccb9dc977934cab4",
        "status": 0
      },
      {
        "headerHash": "0xf65494119976f3d44359fdf09b8bea928f361d2bc2dee83e49fc39f777209e67",
        "status": 0
      },
      null,
      null,
      null,
      null,
      null,
      null,
      null,
      {
        "headerHash": "0x0603dd50e47f30c171c7741581ec2f24c4649ff90fdb9bcfdc35bbcb5908c62f",
        "status": 0
      },
      {
        "headerHash": "0x2daeca22db012299b1ddc64540c5c119efdfbb3b24bd860caf386528af66423c",
        "status": 0
      },
      {
        "headerHash": "0xda04180d79e4378061752d241afbae137f2040db0c610765070f5a0c9639f5e2",
        "status": 0
      },
      {
        "headerHash": "0xd1b1dd685419c552309e6301fb44cc00a248a775626faa530981ef9f7b809288",
        "status": 0
      },
      {
        "headerHash": "0x70313a0a7262f1584943617a9d118a5c2a900990c08b7af2f415d5e236dc97b0",
        "status": 0
      },
      {
        "headerHash": "0xca4287472a85d04411d1bde2965d93de1886bbb4bc7f4cf8dc48ada373b68b69",
        "status": 0
      },
      {
        "headerHash": "0xc003e1c75facd632897ebfc2542069d926a06798d4ba859b667ebb39d804e900",
        "status": 0
      },
      null,
      {
        "headerHash": "0x19c01e5201b2e00fb56dbfd2f23b04552020cd055a066264ef10e4a3dfa2a47e",
        "status": 0
      },
      {
        "headerHash": "0x75cef272d89f408f97604369d5c2907afeabda60055d6c215d022e4fabfec390",
        "status": 0
      },
      {
        "headerHash": "0xb833be03ee17dd651fd3e05cfac093f8c2099001e81db0abd6246b04d86ecd81",
        "status": 0
      },
      {
        "headerHash": "0x09bafe0bdd20e2bc4803426895221fcce644ce027991b0522946b23ccde9cb19",
        "status": 0
      },
      {
        "headerHash": "0x90c95f371c1d37b4279e27864cec3e30c57de3a57591098ed21068bc88639bf3",
        "status": 0
      },
      {
        "headerHash": "0xeca535cc9a65dc535f2c336f77cc7d0d4196d22cd6eb3d30f3f9358ed58bd4ae",
        "status": 0
      },
      {
        "headerHash": "0x7cde50abb91f2c91067629e4cc0e15cf69a77bb49402d5110039bb9613993de6",
        "status": 0
      },
      {
        "headerHash": "0x0bd7243fb0d94b312d6dab2ada6366509b32c8cf29730787673418a662cac2e4",
        "status": 0
      },
      null,
      {
        "headerHash": "0x60e646d7b5e616aee3d16544e05e274c402d796e6c782a9f0081c9038d051a7a",
        "status": 0
      },
      {
        "headerHash": "0xd4831f5d85ca210c81c76d4be91b2271c1da1a35625e2630b803f9e323449418",
        "status": 0
      },
      {
        "headerHash": "0x796e401badf03952f9b534b8ec6a349bd5b09668a3cd8511224f8168a9868006",
        "status": 0
      },
      {
        "headerHash": "0x6c556f5ff3ced4556b2625cfe2ad1b040ff3a1501e9ad30a1f9661705eb3daeb",
        "status": 0
      },
      {
        "headerHash": "0xdc3797ff8a5cc6f46e21aa243894f1301380edbc0cf017144f5a532f410d7214",
        "status": 0
      },
      {
        "headerHash": "0x59540c3e95a1cf3ae8845dd49067dff5f39265ecd2d4d915bedebb2adcf7bc44",
        "status": 0
      },
      {
        "headerHash": "0xdba8235408121c119f73e3e8febfe05d4aff49ba5046aecce231a83aff2d36c9",
        "status": 0
      },
      {
        "headerHash": "0xf10691515b20b9d3c143acdd19bae152048bcf9d7032c06a98507f5fbf61a144",
        "status": 0
      },
      null,
      {
        "headerHash": "0x454fccdcdbcc854a72b563f2084bc8f8d141ce9d7538bbfcbc8c2637336f6700",
        "status": 0
      },
      null,
      null,
      {
        "headerHash": "0x97185e02534f89b0a04e848c9ed07d433fc9906fe46f3e2e2d5660302ff7f6c3",
        "status": 0
      },
      {
        "headerHash": "0xeb371b4f5c9665a215868322a2fecc815d33bef9d60185d63467dbd65dc65e69",
        "status": 0
      },
      {
        "headerHash": "0xdd881a462f22c485637b6a3c2168c32291d69f25c71f76075ed3981d691a68f5",
        "status": 0
      },
      null,
      {
        "headerHash": "0xf7b20f121dc818efe9d30df1f27bbaa060e12b11cc9ae2d7c52cb92355820422",
        "status": 0
      },
      {
        "headerHash": "0x078ed0e94e50738b567764f8587b76a0c0a6bef2fd4ac8f6665b55cddba055db",
        "status": 0
      },
      null,
      null,
      {
        "headerHash": "0x9cea5d4952be0bae951b717f4c8257a225cf837fb5720ce57293606219c990fc",
        "status": 0
      }
    ]
`
)
