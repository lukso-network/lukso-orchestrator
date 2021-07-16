package cache

import "github.com/pkg/errors"

var (
	// maxPanHeaderCacheSize with 1024 consensus infos will be 1024 * 1.5kb.
	maxPanHeaderCacheSize = 1 << 10

	// errInvalidSlot
	errInvalidSlot = errors.New("Invalid slot")

	// errAddingCache is error while put data into cache failed
	errAddingCache = errors.New("error adding data to cache")
)
