package config

import (
	"errors"
)

var (
	ErrReadOnly = errors.New("not available in read-only mode")
)
