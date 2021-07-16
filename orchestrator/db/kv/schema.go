package kv

var (
	// 3 buckets for containing orchestrator data
	consensusInfosBucket    = []byte("consensus-info")
	verifiedSlotInfosBucket = []byte("verified-slots")
	invalidSlotInfosBucket  = []byte("invalid-slots")

	latestHeaderHashKey        = []byte("latest-header-hash")
	lastStoredEpochKey         = []byte("last-epoch")
	latestSavedVerifiedSlotKey = []byte("latest-verified-slot")
)
