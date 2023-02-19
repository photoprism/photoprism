package clean

import (
	"strconv"
	"strings"
)

// ID sanitizes identifier tokens, for example, a session ID, a UUID, or some other string ID.
func ID(s string) string {
	if s == "" || len(s) > 4096 {
		return ""
	}

	var prev rune

	// Remove all invalid characters.
	s = strings.Map(func(r rune) rune {
		if r == ' ' && (prev == 0 || prev == ' ') {
			return -1
		}

		prev = r

		switch r {
		case ' ', '"', '-', '+', '/', '=', '#', '$', '@', ':', ';', '_':
			return r
		}

		if (r < '0' || r > '9') && (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
			return -1
		}

		return r
	}, s)

	return s
}

// UID sanitizes unique identifier strings and returns them in lowercase.
func UID(s string) string {
	if l := len(s); l < 16 || l > 64 {
		return ""
	}

	s = strings.ToLower(s)

	// Remove all invalid characters.
	s = strings.Map(func(r rune) rune {
		switch r {
		case '-', '_', ':':
			return r
		}

		if (r < '0' || r > '9') && (r < 'a' || r > 'z') {
			return -1
		}

		return r
	}, s)

	return s
}

// IdUint converts the string converted to an unsigned integer and 0 if the string is invalid.
func IdUint(s string) uint {
	// Largest possible values:
	// UInt64: 18446744073709551615 (20 digits)
	// UInt32: 4294967295 (10 digits)
	if s == "" || reject(s, 10) {
		return 0
	}

	s = strings.TrimSpace(s)

	result, err := strconv.ParseUint(s, 10, 32)

	if err != nil {
		return 0
	}

	return uint(result)
}
