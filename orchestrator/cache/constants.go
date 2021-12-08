package cache

import (
	"github.com/pkg/errors"
	"time"
)

var (
	cacheRemovalInterval = time.Second * 8

	errAlreadyInProgress = errors.New("requested slot number is already in progress")
)
