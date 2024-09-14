package clean

import (
	"strings"
	"unicode"
)

// FileName removes invalid character from a filename string.
func FileName(s string) string {
	if s == "" || reject(s, 512) || strings.Contains(s, "/") || strings.Contains(s, "..") {
		return ""
	}

	// Trim whitespace.
	s = strings.TrimSpace(s)

	// Remove non-printable and other potentially problematic characters.
	// The following characters must never be used in a filename: / \ > < | : &
	s = strings.Map(func(r rune) rune {
		if !unicode.IsPrint(r) {
			return -1
		}

		switch r {
		case '~', '/', '\\', ':', '|', '"', '?', '*', '<', '>', '{', '}':
			return -1
		default:
			return r
		}
	}, s)

	if s == "." || s == ".." {
		return ""
	}

	return s
}
