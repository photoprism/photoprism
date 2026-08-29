package entity

import "fmt"

var (
	// ErrInvalidName is returned when a name fails validation.
	ErrInvalidName = fmt.Errorf("invalid name")
)
