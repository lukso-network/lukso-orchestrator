package utils

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	"testing"
	"time"
)

func Test_Debounce(t *testing.T) {
	debounceFunc := func(
		capacity int,
		interval time.Duration,
		checkAfter time.Duration,
		expectedHandlerCalls int,
	) {
		eventChannel := make(chan interface{}, capacity)
		ctx := context.Background()
		triggeredChannel := make(chan bool)

		handler := func(triggeredTimes interface{}) {
			triggeredChannel <- true
		}

		dummyEvent := struct {
			Something int
		}{}

		for index := 0; index < capacity; index++ {
			eventChannel <- dummyEvent
		}

		triggeredTimes := 0

		go Debounce(ctx, interval, eventChannel, handler)
		go func() {
			for {
				<-triggeredChannel
				triggeredTimes++
			}
		}()

		time.Sleep(checkAfter)
		require.Equal(t, expectedHandlerCalls, triggeredTimes)
	}

	t.Run("should be invoked 1 time", func(t *testing.T) {
		capacity := 20
		interval := time.Millisecond * 10
		checkAfter := interval*time.Duration(capacity) + time.Millisecond*2
		debounceFunc(capacity, interval, checkAfter, 1)
	})

	t.Run("should not be invoked", func(t *testing.T) {
		capacity := 500000
		interval := time.Millisecond * 10
		checkAfter := time.Millisecond * 15
		debounceFunc(capacity, interval, checkAfter, 0)
	})
}
