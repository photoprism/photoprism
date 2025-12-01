package config

import (
	"errors"
)

var (
	// ErrReadOnly indicates an action is not permitted in read-only mode.
	ErrReadOnly = errors.New("not available in read-only mode")
)

// LogErr logs a config-related error if it is non-nil.
func LogErr(err error) {
	if err != nil {
		log.Errorf("config: %s", err.Error())
	}
}
