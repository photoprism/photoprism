package config

import (
	"errors"
)

var (
	// ErrReadOnly indicates an action is not permitted in read-only mode.
	ErrReadOnly = errors.New("not available in read-only mode")

	// ErrInvalidOptionValue indicates a value cannot be stored as the type its option is declared
	// with. Callers that accept option values from a request should report it as a bad request.
	ErrInvalidOptionValue = errors.New("invalid option value")
)

// LogErr logs a config-related error if it is non-nil.
func LogErr(err error) {
	if err != nil {
		log.Errorf("config: %s", err.Error())
	}
}
