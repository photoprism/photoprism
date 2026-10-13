package clean

import "errors"

var (
	// ErrTooLong is returned when a value had to be shortened to its maximum length.
	ErrTooLong = errors.New("too long")

	// ErrInvalidChars is returned when characters other than surrounding whitespace had to be removed or replaced.
	ErrInvalidChars = errors.New("invalid characters")
)
