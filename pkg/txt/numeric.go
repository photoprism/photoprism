package txt

import (
	"strings"
)

// Numeric removes non-numeric characters from a string and returns the result.
func Numeric(s string) string {
	if s == "" {
		return ""
	}

	sep := '.'

	if c := strings.Count(s, "."); c == 0 || c > 1 {
		sep = ','
	}

	// Remove invalid characters.
	s = strings.Map(func(r rune) rune {
		if r == sep {
			return '.'
		} else if r == '-' {
			return '-'
		} else if r < '0' || r > '9' {
			return -1
		}

		return r
	}, s)

	return s
}
