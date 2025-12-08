package entity

import (
	"strings"
)

const (
	ClipStringType = 64
)

// ToASCII removes all non-ASCII runes from the string.
func ToASCII(s string) string {
	result := make([]rune, 0, len(s))

	for _, r := range s {
		if r <= 127 {
			result = append(result, r)
		}
	}

	return string(result)
}

// Clip trims leading/trailing whitespace and shortens the string to maxLen characters.
func Clip(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	l := len(s)

	if l <= maxLen {
		return s
	} else {
		return s[:maxLen]
	}
}

// SanitizeStringType normalizes identifier-like strings by stripping non-ASCII runes and clipping to 32 characters.
func SanitizeStringType(s string) string {
	return Clip(ToASCII(s), ClipStringType)
}

// SanitizeStringTypeLower lowercases the string before applying SanitizeStringType.
func SanitizeStringTypeLower(s string) string {
	return SanitizeStringType(strings.ToLower(s))
}

// TypeString returns an entity type string for logging, defaulting to "unknown".
func TypeString(entityType string) string {
	if entityType == "" {
		return "unknown"
	}

	return entityType
}
