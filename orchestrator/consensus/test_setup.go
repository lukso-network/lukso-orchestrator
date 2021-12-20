package consensus

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/db"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/utils"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/event"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/cache"
	testDB "github.com/lukso-network/lukso-orchestrator/orchestrator/db/testing"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
)

type mockFeedService struct {
	headerInfoFeed           event.Feed
	shardInfoFeed            event.Feed
	subscriptionShutdownFeed event.Feed
	scope                    event.SubscriptionScope
}

func (mc *mockFeedService) SubscribeShutdownSignalEvent(signals chan<- *types.Reorg) event.Subscription {
	return mc.scope.Track(mc.subscriptionShutdownFeed.Subscribe(signals))
}

func (mc *mockFeedService) StopSubscription(reorgInfo *types.Reorg) {
	panic("implement me")
}

func (mc *mockFeedService) Resubscribe() {
	panic("implement Resubscribe")
}

func (mc *mockFeedService) SubscribeHeaderInfoEvent(ch chan<- *types.PandoraHeaderInfo) event.Subscription {
	return mc.scope.Track(mc.headerInfoFeed.Subscribe(ch))
}

func (mc *mockFeedService) SubscribeShardInfoEvent(ch chan<- *types.VanguardShardInfo) event.Subscription {
	return mc.scope.Track(mc.shardInfoFeed.Subscribe(ch))
}

func setup(ctx context.Context, t *testing.T) (*Service, *mockFeedService, db.Database, *utils.Stack, *utils.Stack) {
	testDB := testDB.SetupDB(t)
	mfs := new(mockFeedService)
	now := uint64(time.Now().Unix())
	panStack := utils.NewStack()
	vanStack := utils.NewStack()

	cfg := &Config{
		VerifiedShardInfoDB: testDB,
		PanHeaderCache:      cache.NewPandoraCache(1024, now, 6, panStack),
		VanShardCache:       cache.NewVanguardCache(1024, now, 6, vanStack),
		VanguardShardFeed:   mfs,
		PandoraHeaderFeed:   mfs,
	}

	return New(ctx, cfg), mfs, testDB, panStack, vanStack
}

func getHeaderInfosAndShardInfos(fromSlot uint64, num uint64) ([]*types.PandoraHeaderInfo, []*types.VanguardShardInfo) {
	headerInfos := make([]*types.PandoraHeaderInfo, 0)
	vanShardInfos := make([]*types.VanguardShardInfo, 0)

	for i := fromSlot; i <= num; i++ {
		headerInfo := new(types.PandoraHeaderInfo)
		headerInfo.Header = testutil.NewEth1Header(i)
		headerInfo.Slot = i
		if i > 1 {
			headerInfo.Header.ParentHash = headerInfos[i-2].Header.Hash()
		}

		vanShardInfo := testutil.NewVanguardShardInfo(i, headerInfo.Header, 0, 0)
		if i > 1 {
			vanShardInfo.ParentRoot = vanShardInfos[i-2].BlockRoot
		}
		vanShardInfos = append(vanShardInfos, vanShardInfo)
		headerInfos = append(headerInfos, headerInfo)
	}
	return headerInfos, vanShardInfos
}
