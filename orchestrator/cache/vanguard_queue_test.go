package cache

import (
	"github.com/lukso-network/lukso-orchestrator/orchestrator/utils"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestVanguardCache_Put(t *testing.T) {
	pc := NewVanguardCache(1024, 0, 6, utils.NewStack())
	_, vanShardInfos := testutil.GetHeaderInfosAndShardInfos(1, 25)
	for _, vanShard := range vanShardInfos {
		pc.Put(vanShard.Slot, &VanCacheInsertParams{
			CurrentShardInfo:      vanShard,
			LastVerifiedShardInfo: nil,
			DisableDelete:         true,
		})
	}
	assert.Equal(t, 25, len(pc.cache.Keys()))

	t.Run("test async vanguard push", func(t *testing.T) {
		vanCache := NewVanguardCache(1<<10, 0, 6, utils.NewStack())
		var wg sync.WaitGroup
		putTimeMap := make(map[uint64]time.Time)
		//getTimeMap := make(map[uint64]time.Time)

		wg.Add(3)
		_, vanInfoGroup1 := testutil.GetHeaderInfosAndShardInfos(1, 1<<10)
		runtime.GOMAXPROCS(3)
		go func() {
			defer wg.Done()
			for _, vanInfo := range vanInfoGroup1 {
				putTimeMap[vanInfo.Slot] = time.Now()
				err := vanCache.Put(vanInfo.Slot, &VanCacheInsertParams{
					CurrentShardInfo: vanInfo,
					DisableDelete:    true,
				})
				assert.NoError(t, err)
			}
		}()
		go func() {
			defer wg.Done()
			ticker := time.NewTicker(1 * time.Millisecond)
			counter := 0
			for counter < 100 {
				select {
				case <-ticker.C:
					vanCache.MarkInProgress(uint64(1))
					vanCache.Get(uint64(1))
					time.Sleep(2 * time.Millisecond)
					vanCache.MarkNotInProgress(uint64(1))
					counter++
				}
			}
		}()

		go func() {
			defer wg.Done()
			ticker := time.NewTicker(1 * time.Millisecond)
			counter := 0
			for counter < 100 {
				select {
				case <-ticker.C:
					vanCache.MarkInProgress(uint64(1))
					vanCache.Get(uint64(1))
					vanCache.MarkNotInProgress(uint64(1))
					counter++
				}
			}
		}()
		wg.Wait()
	})
}
