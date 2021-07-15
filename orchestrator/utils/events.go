package utils

import (
	"context"
	log "github.com/sirupsen/logrus"
	"time"
)

// This will trigger handler only after certain timeout
func Debounce(
	ctx context.Context,
	interval time.Duration,
	eventsChan <-chan interface{},
	handler func(interface{}),
) {
	for event := range eventsChan {
	loop:
		for {
			timer := time.NewTimer(interval)

			select {
			case event = <-eventsChan:
				log.WithField("event", event).Warn("I am discarding the flood")
			case <-timer.C:
				handler(event)
				timer.Stop()
				break loop
			case <-ctx.Done():
				timer.Stop()
				return
			}

			timer.Stop()
		}
	}
}
