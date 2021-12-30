package pandorachain

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	IsPandoraConnected = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "panchain_sync_pandora_connected",
		Help: "Boolean indicating whether an pandora node's endpoint is currently connected: 0=false, 1=true.",
	})

	RequestedFromBlockNumber = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "panchain_requested_from_block",
		Help: "Orchestrator subscribes for pending headers to pandora chain from this block number",
	})

	RequestedFromSlot = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "panchain_requested_from_slot",
		Help: "Orchestrator subscribes for pending headers to pandora chain from this slot",
	})
)
