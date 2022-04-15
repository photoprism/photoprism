package clean

import (
	"strings"
)

// Type omits invalid runes, ensures a maximum length of 32 characters, and returns the result.
func Type(s string) string {
	return Clip(ASCII(s), ClipType)
}

// TypeLower converts a type string to lowercase, omits invalid runes, and shortens it if needed.
func TypeLower(s string) string {
	return Type(strings.ToLower(s))
}

// ShortType omits invalid runes, ensures a maximum length of 8 characters, and returns the result.
func ShortType(s string) string {
	return Clip(ASCII(s), ClipShortType)
}

// ShortTypeLower converts a short type string to lowercase, omits invalid runes, and shortens it if needed.
func ShortTypeLower(s string) string {
	return ShortType(strings.ToLower(s))
}

// LogType returns an entity type string for logging.
func LogType(entityType string) string {
	if entityType == "" {
		return "<unknown-type>"
	}

	return entityType
}
