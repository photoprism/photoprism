package txt

import (
	"strings"
)

// Empty checks whether a string represents an empty, unset, or undefined value.
func Empty(s string) bool {
	if s == "" {
		return true
	} else if s = strings.Trim(s, "%* "); s == "" || s == "0" || s == "-1" || EmptyTime(s) {
		return true
	} else if s = strings.ToLower(s); s == "nil" || s == "null" || s == "nan" {
		return true
	}

	return false
}

// NotEmpty tests if a string does not represent an empty/invalid value.
func NotEmpty(s string) bool {
	return !Empty(s)
}

// EmptyTime tests if the string is empty or matches an unknown time pattern.
func EmptyTime(s string) bool {
	switch s {
	case "":
		return true
	case "0000:00:00 00:00:00", "0000-00-00 00-00-00", "0000-00-00 00:00:00":
		return true
	case "0001-01-01 00:00:00", "0001-01-01 00:00:00 +0000 UTC":
		return true
	default:
		return false
	}
}
