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

// TypeLowerUnderscore converts a string to a lowercase type string and replaces spaces with underscores.
func TypeLowerUnderscore(s string) string {
	return strings.ReplaceAll(TypeLower(s), " ", "_")
}

// ShortType omits invalid runes, ensures a maximum length of 8 characters, and returns the result.
func ShortType(s string) string {
	return Clip(ASCII(s), ClipShortType)
}

// ShortTypeLower converts a short type string to lowercase, omits invalid runes, and shortens it if needed.
func ShortTypeLower(s string) string {
	return ShortType(strings.ToLower(s))
}

// ShortTypeLowerUnderscore converts a string to a short lowercase type string and replaces spaces with underscores.
func ShortTypeLowerUnderscore(s string) string {
	return strings.ReplaceAll(ShortTypeLower(s), " ", "_")
}
