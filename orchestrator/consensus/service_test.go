package consensus

import (
	"context"
	"github.com/ethereum/go-ethereum/event"
	"github.com/lukso-network/lukso-orchestrator/orchestrator/cache"
	testDB "github.com/lukso-network/lukso-orchestrator/orchestrator/db/testing"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/types"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"testing"
	"time"
)

type mockFeedService struct {
	headerInfoFeed event.Feed
	shardInfoFeed  event.Feed
	scope          event.SubscriptionScope
}

func (mc *mockFeedService) SubscribeHeaderInfoEvent(ch chan<- *types.PandoraHeaderInfo) event.Subscription {
	return mc.scope.Track(mc.headerInfoFeed.Subscribe(ch))
}

func (mc *mockFeedService) SubscribeShardInfoEvent(ch chan<- *types.VanguardShardInfo) event.Subscription {
	return mc.scope.Track(mc.shardInfoFeed.Subscribe(ch))
}

func (mc *mockFeedService) sendHeaderInfo(headerInfo *types.PandoraHeaderInfo) {
	mc.headerInfoFeed.Send(headerInfo)
}

func (mc *mockFeedService) sendShardInfo(shardInfo *types.VanguardShardInfo) {
	mc.shardInfoFeed.Send(shardInfo)
}

func setup(ctx context.Context, t *testing.T) *Service {
	testDB := testDB.SetupDB(t)
	mfs := new(mockFeedService)

	cfg := &Config{
		VerifiedSlotInfoDB:           testDB,
		InvalidSlotInfoDB:            testDB,
		VanguardPendingShardingCache: cache.NewVanShardInfoCache(1024),
		PandoraPendingHeaderCache:    cache.NewPanHeaderCache(),
		VanguardShardFeed:            mfs,
		PandoraHeaderFeed:            mfs,
	}

	return New(ctx, cfg)
}

func TestService_Start(t *testing.T) {
	hook := logTest.NewGlobal()
	ctx := context.Background()
	svc := setup(ctx, t)
	defer svc.Stop()

	svc.Start()
	time.Sleep(1 * time.Second)
	assert.LogsContain(t, hook, "Starting consensus service")
	hook.Reset()
}

func TestSevice_Subscription(t *testing.T) {
	hook := logTest.NewGlobal()
	ctx := context.Background()
	svc := setup(ctx, t)
	defer svc.Stop()

	svc.Start()

	svc.vanguardShardFeed.SubscribeShardInfoEvent()

	time.Sleep(1 * time.Second)
	assert.LogsContain(t, hook, "Starting consensus service")
	hook.Reset()
}
