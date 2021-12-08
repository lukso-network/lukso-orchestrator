package cache

import (
	"github.com/pkg/errors"
	"math"
	"time"
)

var (
	// maxCacheSize with 1024 consensus infos will be 1024 * 1.5kb.
	maxCacheSize = 1 << 10

	// need to define maximum size. It will take maximum latest 100 epochs
	maxInt = math.MaxInt32 - 1

	cacheRemovalInterval = time.Second * 8

	// errInvalidSlot
	errInvalidSlot = errors.New("Invalid slot")

	// errAddingCache is error while put data into cache failed
	errAddingCache = errors.New("error adding data to cache")

	// errRemoveCache is error while removing invalid slot number from the cache
	errRemoveCache = errors.New("invalid slot removal")

	errAlreadyInProgress = errors.New("requested slot number is already in progress")
)
