package clean

import (
	"strings"
)

// Token removes invalid character from a token string.
func Token(s string) string {
	if s == "" || reject(s, 200) {
		return ""
	}

	// Remove all invalid characters.
	s = strings.Map(func(r rune) rune {
		if (r < '0' || r > '9') && (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && r != '-' && r != '_' && r != ':' {
			return -1
		}

		return r
	}, s)

	return s
}
