package vanguardchain

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	IsVanConnected = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "vanchain_sync_pandora_connected",
		Help: "Boolean indicating whether an vanguard node's endpoint is currently connected: 0=false, 1=true.",
	})

	RequestedFromSlot = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "vanchain_requested_from_slot",
		Help: "Orchestrator subscribes for pending blocks to vanguard chain from this slot",
	})

	RequestedFromEpoch = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "vanchain_requested_from_epoch",
		Help: "Orchestrator subscribes for minimal consensus infos to vanguard chain from this epoch",
	})
)
