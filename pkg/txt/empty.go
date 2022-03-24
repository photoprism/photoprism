package txt

import (
	"strings"
)

// IsEmpty tests if a string represents an empty/invalid value.
func IsEmpty(s string) bool {
	s = strings.Trim(strings.TrimSpace(s), "%*")

	if s == "" || s == "0" || s == "-1" {
		return true
	}

	s = strings.ToLower(s)

	return s == "nil" || s == "null" || s == "nan"
}

// NotEmpty tests if a string does not represent an empty/invalid value.
func NotEmpty(s string) bool {
	return !IsEmpty(s)
}
