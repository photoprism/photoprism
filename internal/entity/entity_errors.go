package entity

import "fmt"

var (
	// ErrInvalidName is returned when a name fails validation.
	ErrInvalidName = fmt.Errorf("invalid name")

	// ErrInvalidValue marks a submitted value that failed validation, so a handler can answer with a
	// client error rather than reporting a server fault. Wrap it, keeping the message that names the
	// value: fmt.Errorf("%w: birthday must not be in the future", ErrInvalidValue).
	ErrInvalidValue = fmt.Errorf("invalid value")
)
