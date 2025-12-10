package places

import (
	"errors"
)

// ErrMissingQuery indicates that a place search query was empty.
var ErrMissingQuery = errors.New("missing query")

// ErrMissingCoordinates indicates that both latitude and longitude were missing.
var ErrMissingCoordinates = errors.New("missing coordinates")
