package clean

import (
	"strings"
)

// Token returns the sanitized token string with a length of up to 4096 characters.
func Token(s string) string {
	if s == "" || reject(s, LengthLimit) {
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

// UrlToken returns the sanitized URL token with a length of up to 42 characters.
func UrlToken(s string) string {
	if s == "" || len(s) > 64 {
		return ""
	}

	return Token(s)
}

// ShareToken returns the sanitized link share token with a length of up to 160 characters.
func ShareToken(s string) string {
	if s == "" || len(s) > 160 {
		return ""
	}

	return Token(s)
}
