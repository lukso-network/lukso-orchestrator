package consensus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ReorgDetectedCnt = promauto.NewCounter(prometheus.CounterOpts{
		Name: "reorg_count",
		Help: "Reorg detected count after inserting any vanguard slot",
	})

	ReorgResolvedCnt = promauto.NewCounter(prometheus.CounterOpts{
		Name: "reorg_resolve_count",
		Help: "Successfully resolved reorg count",
	})

	LiveSyncCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "orc_live_sync_transition_count",
		Help: "Orchestrator live sync transition count from initial sync or from re-sync mode",
	})

	InitialOrResyncCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "orc_initial_or_resync_transition_count",
		Help: "Orchestrator initial or re-sync transition count from live sync",
	})

	CurSyncingStatus = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "orc_current_syncing_status",
		Help: "Orchestrator's current syncing status: 0=false, 1=true",
	})

	LatestVerifiedSlot = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "latest_verified_slot",
		Help: "Latest inserted verified slot",
	})

	LatestVerifiedBlockNumber = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "latest_verified_block_number",
		Help: "Latest inserted verified pandora sharding block number",
	})

	LatestFinalizedSlot = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "latest_finalized_slot",
		Help: "Latest inserted finalized slot",
	})

	LatestFinalizedEpoch = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "latest_finalized_epoch",
		Help: "Latest inserted finalized epoch",
	})

	TotalMisMatchedShardCnt = promauto.NewCounter(prometheus.CounterOpts{
		Name: "total_mismatched_shard_count",
		Help: "Total mismatched shard count",
	})
)
