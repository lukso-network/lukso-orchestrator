package vanguardchain

//func TestService_SubscribeVanNewPendingBlockHash(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//	client := mock.NewMockBeaconChainClient(ctrl)
//
//	testDB := testDB.SetupDB(t)
//	ctx := context.Background()
//
//	bs := Service{
//		ctx:            ctx,
//		db: testDB,
//		beaconClient:   client,
//	}
//	_, cancel := context.WithCancel(ctx)
//	stream := mock.NewMockBeaconChain_StreamNewPendingBlocksClient(ctrl)
//	client.EXPECT().StreamNewPendingBlocks(ctx, &eth.StreamPendingBlocksRequest{BlockRoot: []byte{}, FromSlot: 0}).Return(stream, nil)
//	stream.EXPECT().Context().Return(ctx).AnyTimes()
//	stream.EXPECT().Recv().Return(
//		testutil.NewBeaconBlock(0),
//		nil,
//	).Do(func() {
//		cancel()
//	})
//	bs.subscribeVanNewPendingBlockHash(0)
//}

//// Test_VanguardChainStartStop_Initialized
//func Test_VanguardChainStartStop_Initialized(t *testing.T) {
//	hook := logTest.NewGlobal()
//	ctx := context.Background()
//	vanSvc, _ := SetupVanguardSvc(ctx, t, GRPCFunc)
//	vanSvc.Start()
//	defer func() {
//		_ = vanSvc.Stop()
//	}()
//
//	time.Sleep(1 * time.Second)
//	ConsensusInfoMocks = make([]*eth.MinimalConsensusInfo, 0)
//	minimalConsensusInfo := testutil.NewMinimalConsensusInfo(0)
//
//	ConsensusInfoMocks = append(ConsensusInfoMocks, &eth.MinimalConsensusInfo{
//		SlotTimeDuration: &duration.Duration{Seconds: 6},
//		ValidatorList:    minimalConsensusInfo.ValidatorList,
//	})
//	PendingBlockMocks = nil
//
//	defer func() {
//		ConsensusInfoMocks = nil
//		PendingBlockMocks = nil
//	}()
//
//	time.Sleep(1 * time.Second)
//	assert.LogsContain(t, hook, "Received new consensus info for next epoch")
//	hook.Reset()
//}
//
//func TestService_OnNewConsensusInfo(t *testing.T) {
//	ctx := context.Background()
//	vanSvc, newTestDB := SetupVanguardSvc(ctx, t, GRPCFunc)
//
//	nonReorgInfo := &types.MinimalEpochConsensusInfoV2{
//		Epoch:            5,
//		ValidatorList:    nil,
//		EpochStartTime:   0,
//		SlotTimeDuration: 0,
//		ReorgInfo:        nil,
//	}
//
//	require.NoError(t, vanSvc.onNewConsensusInfo(ctx, nonReorgInfo))
//	require.Equal(t, nonReorgInfo.Epoch, newTestDB.LatestSavedEpoch())
//	fetchedConsensusInfo, err := newTestDB.ConsensusInfo(ctx, nonReorgInfo.Epoch)
//	require.NoError(t, err)
//	require.Equal(t, nonReorgInfo.Epoch, fetchedConsensusInfo.Epoch)
//	require.Equal(t, nonReorgInfo.Epoch, newTestDB.LatestSavedEpoch())
//
//	t.Run("should revert to epoch 1", func(t *testing.T) {
//		nonReorgInfoEpoch1 := &types.MinimalEpochConsensusInfoV2{Epoch: 1}
//		require.NoError(t, vanSvc.onNewConsensusInfo(ctx, nonReorgInfoEpoch1))
//		require.Equal(t, nonReorgInfoEpoch1.Epoch, newTestDB.LatestSavedEpoch())
//
//		//	 PROBLEM NO 1 FOUND. CONSENSUS INFOS CAN BE INSERTED NONCONSECUTIVE
//		// IT MAY LEAD TO BREAK OF LOGIC OF THE APPLICATION
//		// THIS SHOULD BE PROTECTED
//	})
//
//	t.Run("should react to reorg info", func(t *testing.T) {
//		for epoch := uint64(0); epoch <= nonReorgInfo.Epoch; epoch++ {
//			epochInfoRecord := &types.MinimalEpochConsensusInfoV2{Epoch: epoch}
//			require.NoError(t, vanSvc.onNewConsensusInfo(ctx, epochInfoRecord))
//		}
//
//		// Missing scenario is what if ReorgInfo is not nil, but its empty
//		vanBlockHash := common.HexToHash("0xfcae73c029aa80d9bbc79cda6f23a02fb3bc3d543ca4793456f73125ed9bfecb")
//		panBlockHash := common.HexToHash("0x0b5d32ba8e74ab81d699a585c38bb6d5b62079089d8ff412729fe1fdd3c43497")
//
//		reorgInfoEpoch2 := &types.MinimalEpochConsensusInfoV2{
//			Epoch: 2,
//			ReorgInfo: &types.Reorg{
//				VanParentHash: vanBlockHash.Bytes(),
//				PanParentHash: panBlockHash.Bytes(),
//				NewSlot:       uint64(66),
//			},
//		}
//		require.NoError(t, newTestDB.SaveVerifiedSlotInfo(reorgInfoEpoch2.Epoch*32, &types.SlotInfo{
//			VanguardBlockHash: vanBlockHash,
//			PandoraHeaderHash: panBlockHash,
//		}))
//		require.NoError(t, newTestDB.SaveLatestVerifiedSlot(ctx))
//
//		consensusInfos, currentErr := newTestDB.ConsensusInfos(reorgInfoEpoch2.Epoch)
//		require.NoError(t, currentErr)
//		require.Equal(t, true, len(consensusInfos) == int(nonReorgInfo.Epoch-reorgInfoEpoch2.Epoch)+1)
//		require.NoError(t, vanSvc.onNewConsensusInfo(ctx, reorgInfoEpoch2))
//
//		consensusInfos, currentErr = newTestDB.ConsensusInfos(reorgInfoEpoch2.Epoch)
//		require.NoError(t, currentErr)
//		require.Equal(t, true, len(consensusInfos) == 1)
//		require.Equal(t, reorgInfoEpoch2.Epoch, newTestDB.LatestSavedEpoch())
//
//		fetchedConsensus, currentErr := newTestDB.ConsensusInfo(ctx, nonReorgInfo.Epoch)
//		require.NoError(t, currentErr)
//		require.Equal(t, false, nil == fetchedConsensus)
//	})
//}
