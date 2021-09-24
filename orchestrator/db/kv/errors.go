package kv

import "errors"

var (
	ErrValueNotFound = errors.New("value not found for the associated key")
)
