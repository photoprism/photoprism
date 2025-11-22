package clean

import (
	"strings"
)

// FieldName normalizes a struct field identifier so it can be compared safely.
// It strips all characters outside [A-Za-z0-9], rejects empty strings, and
// returns an empty string for inputs longer than 255 bytes to avoid abuse.
func FieldName(s string) string {
	if s == "" || len(s) > 255 {
		return ""
	}

	// Remove all invalid characters.
	s = strings.Map(func(r rune) rune {
		if (r < '0' || r > '9') && (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
			return -1
		}

		return r
	}, s)

	return s
}

// FieldNameLower normalizes a struct field identifier and lowercases it first.
// Useful when callers want case-insensitive comparisons against normalized data.
func FieldNameLower(s string) string {
	if s == "" || len(s) > 255 {
		return ""
	}

	return FieldName(strings.ToLower(s))
}
