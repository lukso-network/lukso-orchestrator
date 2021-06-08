package utils

import (
	"context"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	"testing"
	"time"
)

func Test_Debounce(t *testing.T) {
	capacity := 20
	eventChannel := make(chan interface{}, capacity)
	ctx := context.Background()
	interval := time.Millisecond * 10
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

	time.Sleep(interval*time.Duration(capacity) + time.Millisecond*2)
	require.Equal(t, 1, triggeredTimes)
}
