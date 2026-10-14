package clean

import "errors"

// ModelId sanitizes a model identifier like Type, preserving case, and returns ErrTooLong if it had to be
// shortened to LengthType characters, or ErrInvalidChars if other characters than surrounding whitespace had
// to be removed or replaced. The sanitized identifier is returned in either case.
func ModelId(s string) (string, error) {
	id, changed, clipped := typeString(s, LengthType)

	switch {
	case changed && clipped:
		return id, errors.Join(ErrInvalidChars, ErrTooLong)
	case changed:
		return id, ErrInvalidChars
	case clipped:
		return id, ErrTooLong
	default:
		return id, nil
	}
}
