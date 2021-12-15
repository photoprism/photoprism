package sanitize

import (
	"strconv"
	"strings"
)

// IdString removes invalid character from an id string.
func IdString(s string) string {
	if s == "" || len(s) > 256 || strings.Contains(s, "${") {
		return ""
	}

	s = strings.ToLower(s)

	// Remove all invalid characters.
	s = strings.Map(func(r rune) rune {
		if (r < '0' || r > '9') && (r < 'a' || r > 'z') && r != '-' && r != '_' && r != ':' {
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
	if s == "" || len(s) > 10 || strings.Contains(s, "${") {
		return 0
	}

	result, err := strconv.ParseUint(s, 10, 32)

	if err != nil {
		return 0
	}

	return uint(result)
}
