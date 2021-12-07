package kv

var (
	// 3 buckets for containing orchestrator data
	consensusInfosBucket    = []byte("consensus-info")
	verifiedSlotInfosBucket = []byte("verified-slots")
	invalidSlotInfosBucket  = []byte("invalid-slots")
	latestInfoMarkerBucket  = []byte("latest-info-marker") // Only use for storing the following keys
	multiShardsBucket       = []byte("multi-shards")
	slotStepIndicesBucket   = []byte("slot-step-indices")

	latestHeaderHashKey        = []byte("latest-header-hash")
	lastStoredEpochKey         = []byte("last-epoch")
	latestSavedVerifiedSlotKey = []byte("latest-verified-slot")
	latestFinalizedSlotKey     = []byte("latest-finalized-slot")
	latestFinalizedEpochKey    = []byte("latest-finalized-epoch")
	curStepId                  = []byte("latest-step-id")
)
