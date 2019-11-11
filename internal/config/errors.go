package config

import (
	"errors"
)

var (
	ErrReadOnly = errors.New("not available in read-only mode")
	ErrUnauthorized = errors.New("please log in and try again")
)
