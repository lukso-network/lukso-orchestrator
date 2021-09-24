package kv

var (
	// 3 buckets for containing orchestrator data
	consensusInfosBucket    = []byte("consensus-info")
	verifiedSlotInfosBucket = []byte("verified-slots")
	invalidSlotInfosBucket  = []byte("invalid-slots")
)
