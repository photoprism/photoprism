package thumb

import (
	"errors"
)

var (
	// ErrNotCached indicates a requested thumbnail is not present in cache.
	ErrNotCached = errors.New("not cached")
)
