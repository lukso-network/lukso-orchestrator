package cache

import "errors"

var (
	errParentHashMismatch = errors.New("parent hash mismatched")
	errInvalidElement     = errors.New("requested with invalid element")
	errAlreadyInProgress  = errors.New("requested slot number is already in progress")
)
