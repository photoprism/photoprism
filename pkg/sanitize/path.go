package sanitize

import (
	"strings"
	"unicode"
)

// Path removes invalid character from a path string.
func Path(s string) string {
	if s == "" || reject(s, 512) || strings.Contains(s, "..") || strings.Contains(s, "//") {
		return ""
	}

	// Trim whitespace.
	s = strings.TrimSpace(s)

	// Remove non-printable and other potentially problematic characters.
	s = strings.Map(func(r rune) rune {
		if !unicode.IsPrint(r) {
			return -1
		}

		switch r {
		case '~', '\\', ':', '|', '"', '?', '*', '<', '>', '{', '}':
			return -1
		default:
			return r
		}
	}, s)

	return s
}
