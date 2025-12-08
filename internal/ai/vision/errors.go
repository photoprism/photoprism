package vision

import (
	"fmt"
)

var (
	// ErrInvalidModel indicates an unknown or unsupported vision model name.
	ErrInvalidModel = fmt.Errorf("vision: invalid model")
)
