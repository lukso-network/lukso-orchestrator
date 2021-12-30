package events

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	TotalVanInvalidStatusCnt = promauto.NewCounter(prometheus.CounterOpts{
		Name: "total_invalid_van_status_count",
		Help: "Total invalid status count for vanguard",
	})

	TotalPanInvalidStatusCnt = promauto.NewCounter(prometheus.CounterOpts{
		Name: "total_invalid_pan_status_count",
		Help: "Total invalid status count for pandora",
	})

	RequestedFromEpoch = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "requested_from_epoch",
		Help: "Requested from epoch in which orchestrator subscribe for minimal consensus infos",
	})
)
