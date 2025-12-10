package crop

import (
	"errors"
)

var (
	// ErrNotFound indicates the requested crop size or option was not found.
	ErrNotFound = errors.New("not found")
)
