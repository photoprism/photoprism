package clean

import (
	"strings"
)

// Codec removes non-alphanumeric characters from a string and returns it.
func Codec(s string) string {
	if s == "" {
		return ""
	}

	// Remove unwanted characters.
	s = strings.Map(func(r rune) rune {
		if (r < '0' || r > '9') && (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && r != '_' {
			return -1
		}

		return r
	}, s)

	return s
}
