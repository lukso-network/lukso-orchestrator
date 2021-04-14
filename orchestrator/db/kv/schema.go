package kv

var (
	consensusInfosBucket       = []byte("consensus-info")
	pandoraHeaderHashesBucket  = []byte("headers-pandora")
	vanguardHeaderHashesBucket = []byte("headers-vanguard")

	lastStoredEpochKey    = []byte("last-epoch")
	latestSavedPanSlotKey = []byte("latest-pandora-slot")
)
