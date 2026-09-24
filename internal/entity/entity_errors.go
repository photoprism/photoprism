package entity

import "fmt"

var (
	// ErrInvalidName is returned when a name fails validation.
	ErrInvalidName = fmt.Errorf("invalid name")

	// ErrInvalidValue marks a submitted value that failed validation, so a handler can answer with a
	// client error rather than reporting a server fault. Wrap it, keeping the message that names the
	// value: fmt.Errorf("%w: birthday must not be in the future", ErrInvalidValue).
	ErrInvalidValue = fmt.Errorf("invalid value")

	// ErrInUse is returned when a record cannot be deleted because other records still reference it.
	ErrInUse = fmt.Errorf("in use")

	// ErrSessionNotFound is returned when a session no longer exists in the database.
	ErrSessionNotFound = fmt.Errorf("session not found")
)
