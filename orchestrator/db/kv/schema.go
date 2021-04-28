package kv

var (
	consensusInfosBucket  = []byte("consensus-info")
	pandoraHeadersBucket  = []byte("headers-pandora")
	vanguardHeadersBucket = []byte("headers-vanguard")

	lastStoredEpochKey        = []byte("last-epoch")
	latestSavedPanBlockNumKey = []byte("latest-pandora-block")
	latestSavedPanSlot        = []byte("latest-pandora-slot")
)
