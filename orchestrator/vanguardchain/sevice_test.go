package vanguardchain

// Integration test
/* 1. DB setup (100 verified infos)
	 - latestVerifiedSlot = 100
     - finalizedSlot = 96
     - finalizedEpoch = 3
     - epoch info db
     - verified info db
  2. Service init
	- db
    - cache
    - set mock beaconClient, nodeClient
    - feed: vanguardShardingInfoFeed, subscriptionShutdownFeed
    - endpoint
  3. Service run

Reorg test scenario:
  1. For syncing flow testing:
	- Prepare head container(future any finalized slot and epoch(128 and 4) for `GetChainHead` mock api
    - If we trigger re-org, then test following:
		a. pending block streaming subscription shutdown (check log: "Received re-org event, exiting vanguard pending block streaming subscription!")
        b. check verified slot info db.
 		c. check re-subscription and check log ("Successfully subscribed to vanguard blocks") and fromSlot number

  2. For event based testing:
	- Prepare consensus info with reorg
	- Send consensus info with reorg in mocked `stream.Recv()` function
    - As we pass reorg info so it will trigger re-org, then test following:
		a. pending block streaming subscription shutdown (check log: "Received re-org event, exiting vanguard pending block streaming subscription!")
        b. check verified slot info db.
 		c. check re-subscription and check log ("Successfully subscribed to vanguard blocks") and fromSlot number
**/

// Test_VanguardSvc_StartStop checks start and stop process. When the vanguard service starts, it also subscribes
// van_subscribe to get new consensus info
//func serviceInit(t *testing.T) (*Service, *logTest.Hook) {
//	ctx:= context.Background()
//	hook := logTest.NewGlobal()
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	mockedBeaconClient := mock.NewMockBeaconChainClient(ctrl)
//	mockedNodeClient := mock.NewMockNodeClient(ctrl)
//
//	testDB := testDB.SetupDB(t)
//	cache := cache.NewVanShardInfoCache(1024)
//	s, err := NewService(ctx, "127.0.0.1:4000", testDB, cache)
//	require.NoError(t, err)
//
//	s.beaconClient = mockedBeaconClient
//	s.nodeClient = mockedNodeClient

//blockPendingStream := mock.NewMockBeaconChain_StreamNewPendingBlocksClient(ctrl)
//epochInfoStream := mock.NewMockBeaconChain_StreamMinimalConsensusInfoClient(ctrl)

//go s.run()

//mockedBeaconClient.EXPECT().GetChainHead(
//	gomock.Any(),
//	gomock.Any(),
//	).Return(nil, errors.New("Fuck me")).Times(10)
//mockedBeaconClient.EXPECT().GetChainHead(
//	gomock.Any(),
//	gomock.Any(),
//).Return(&ethpb.ChainHead{}, nil)
//
//mockedBeaconClient.EXPECT().StreamNewPendingBlocks(
//	gomock.Any(),
//	gomock.Any(),
//	).Return(blockPendingStream, nil)
//mockedBeaconClient.EXPECT().StreamMinimalConsensusInfo(
//	gomock.Any(),
//	gomock.Any(),
//	).Return(epochInfoStream, nil)
//blockPendingStream.EXPECT().Recv().Return(
//	&ethpb.StreamPendingBlockInfo{},
//	nil,
//).Do(func() {
//	s.Stop()
//	assert.LogsContain(t, hook, "Connected vanguard chain")
//	hook.Reset()
//})
//epochInfoStream.EXPECT().Recv().Return(
//	&ethpb.MinimalConsensusInfoRequest{},
//	nil,
//).Do(func() {
//	assert.LogsContain(t, hook, "Connected vanguard chain")
//	hook.Reset()
//})

//time.Sleep(30 * time.Second)
//assert.LogsContain(t, hook, "Connected vanguard chain")
//hook.Reset()
//	return s, hook
//}
//
//// Test_VanguardSvc_StartStop checks start and stop process. When the vanguard service starts, it also subscribes
//// van_subscribe to get new consensus info
//func Test_VanguardSvc_NoServerConn(t *testing.T) {
//	hook := logTest.NewGlobal()
//	reConPeriod = 1 * time.Second
//
//	ConsensusInfoMocks = make([]*eth.MinimalConsensusInfo, 0)
//	ConsensusInfoMocks = append(ConsensusInfoMocks, &eth.MinimalConsensusInfo{
//		SlotTimeDuration: &duration.Duration{Seconds: 6}})
//
//	PendingBlockMocks = nil
//
//	defer func() {
//		CleanConsensusMocks()
//		CleanPendingBlocksMocks()
//	}()
//
//	ctx := context.Background()
//	vanSvc, _ := SetupVanguardSvc(ctx, t, GRPCFunc)
//	vanSvc.vanGRPCEndpoint = "wsad://invalid.not.reachable!@:BrOKkeeeeeennnnnnnnnn"
//	vanSvc.dialGRPCFn = DIALGRPCFn(func(endpoint string) (client.VanguardClient, error) {
//		return nil, fmt.Errorf("dummy error")
//	})
//
//	vanSvc.Start()
//	defer func() {
//		_ = vanSvc.Stop()
//	}()
//
//	time.Sleep(1 * time.Second)
//	assert.LogsContain(t, hook, "Could not connect to vanguard endpoint")
//	hook.Reset()
//}
//
//func Test_VanguardSvc_RetryToConnServer(t *testing.T) {
//	hook := logTest.NewGlobal()
//	reConPeriod = 1 * time.Second
//
//	ConsensusInfoMocks = make([]*eth.MinimalConsensusInfo, 0)
//	ConsensusInfoMocks = append(ConsensusInfoMocks, &eth.MinimalConsensusInfo{
//		SlotTimeDuration: &duration.Duration{Seconds: 6}})
//
//	PendingBlockMocks = nil
//
//	defer func() {
//		CleanConsensusMocks()
//		CleanPendingBlocksMocks()
//	}()
//
//	ctx := context.Background()
//	vanSvc, _ := SetupVanguardSvc(ctx, t, GRPCFunc)
//	shouldPass := false
//
//	vanSvc.dialGRPCFn = DIALGRPCFn(func(endpoint string) (client.VanguardClient, error) {
//		if shouldPass {
//			return GRPCFunc(endpoint)
//		}
//
//		return nil, fmt.Errorf("dummy error")
//	})
//
//	oldReconPeriod := reConPeriod
//	reConPeriod = time.Millisecond * 50
//
//	defer func() {
//		reConPeriod = oldReconPeriod
//	}()
//
//	vanSvc.Start()
//	defer func() {
//		_ = vanSvc.Stop()
//	}()
//
//	time.Sleep(10 * reConPeriod)
//	assert.LogsContain(t, hook, "Could not connect to vanguard endpoint")
//	shouldPass = true
//
//	time.Sleep(time.Second)
//	assert.LogsContain(t, hook, "Connected vanguard chain")
//	hook.Reset()
//}
//
//func CleanConsensusMocks() {
//	ConsensusInfoMocks = nil
//}
//
//func CleanPendingBlocksMocks() {
//	PendingBlockMocks = nil
//}
