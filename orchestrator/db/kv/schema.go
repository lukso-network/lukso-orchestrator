package kv

var (
	consensusInfosBucket       = []byte("consensus-info")
	pandoraHeaderHashesBucket  = []byte("headers-pandora")
	vanguardHeaderHashesBucket = []byte("headers-vanguard")
	realmBucket                = []byte("realm")

	lastStoredEpochKey          = []byte("last-epoch")
	latestSavedPanSlotKey       = []byte("latest-pandora-slot")
	latestSavedPanHeaderHashKey = []byte("latest-pandora-header-hash")

	latestSavedVanSlotKey = []byte("latest-vanguard-slot")
	latestSavedVanHashKey = []byte("latest-vanguard-hash")

	//	 Realm
	latestVerifiedRealmSlot = []byte("latest-verified-realm-slot")
)
