package entity

import (
	"strings"
)

const (
	ClipStringType = 64
)

// ToASCII removes all non-ascii characters from a string and returns it.
func ToASCII(s string) string {
	result := make([]rune, 0, len(s))

	for _, r := range s {
		if r <= 127 {
			result = append(result, r)
		}
	}

	return string(result)
}

// Clip shortens a string to the given number of characters, and removes all leading and trailing white space.
func Clip(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	l := len(s)

	if l <= maxLen {
		return s
	} else {
		return s[:maxLen]
	}
}

// SanitizeStringType omits invalid runes, ensures a maximum length of 32 characters, and returns the result.
func SanitizeStringType(s string) string {
	return Clip(ToASCII(s), ClipStringType)
}

// SanitizeStringTypeLower converts a type string to lowercase, omits invalid runes, and shortens it if needed.
func SanitizeStringTypeLower(s string) string {
	return SanitizeStringType(strings.ToLower(s))
}

// TypeString returns an entity type string for logging.
func TypeString(entityType string) string {
	if entityType == "" {
		return "unknown"
	}

	return entityType
}
