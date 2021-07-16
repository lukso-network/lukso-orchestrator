package cache

import "github.com/pkg/errors"

var (
	// maxCacheSize with 1024 consensus infos will be 1024 * 1.5kb.
	maxCacheSize = 1 << 10

	// errInvalidSlot
	errInvalidSlot = errors.New("Invalid slot")

	// errAddingCache is error while put data into cache failed
	errAddingCache = errors.New("error adding data to cache")

	// errRemoveCache is error while removing invalid slot number from the cache
	errRemoveCache = errors.New("invalid slot removal")
)
