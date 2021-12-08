package kv

var (
	// 3 buckets for containing orchestrator data
	consensusInfosBucket   = []byte("consensus-info")
	invalidSlotInfosBucket = []byte("invalid-slots")
	latestInfoMarkerBucket = []byte("latest-info-marker") // Only use for storing the following keys
	multiShardsBucket      = []byte("multi-shards")
	slotStepIndicesBucket  = []byte("slot-step-indices")

	lastStoredEpochKey      = []byte("last-epoch")
	latestFinalizedSlotKey  = []byte("latest-finalized-slot")
	latestFinalizedEpochKey = []byte("latest-finalized-epoch")
	latestStepIdKey         = []byte("latest-step-id")
)
