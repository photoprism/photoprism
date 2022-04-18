package clean

import (
	"strings"
)

// Hex removes invalid character from a hex string and makes it lowercase.
func Hex(s string) string {
	if s == "" || reject(s, 1024) {
		return ""
	}

	s = strings.ToLower(strings.TrimSpace(s))

	// Remove all invalid characters.
	s = strings.Map(func(r rune) rune {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return -1
		}

		return r
	}, s)

	return s
}
