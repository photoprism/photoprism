package clean

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/txt/clip"
)

// Type omits invalid runes, ensures a maximum length of 32 characters, and returns the result.
func Type(s string) string {
	if s == "" {
		return s
	}

	return clip.Chars(ASCII(s), LengthType)
}

// TypeLower converts a type string to lowercase, omits invalid runes, and shortens it if needed.
func TypeLower(s string) string {
	if s == "" {
		return s
	}

	return Type(strings.ToLower(s))
}

// TypeLowerUnderscore converts a string to a lowercase type string and replaces spaces with underscores.
func TypeLowerUnderscore(s string) string {
	if s == "" {
		return s
	}

	return strings.ReplaceAll(TypeLower(s), " ", "_")
}

// ShortType omits invalid runes, ensures a maximum length of 8 characters, and returns the result.
func ShortType(s string) string {
	if s == "" {
		return s
	}

	return clip.Chars(ASCII(s), LengthShortType)
}

// ShortTypeLower converts a short type string to lowercase, omits invalid runes, and shortens it if needed.
func ShortTypeLower(s string) string {
	if s == "" {
		return s
	}

	return ShortType(strings.ToLower(s))
}

// ShortTypeLowerUnderscore converts a string to a short lowercase type string and replaces spaces with underscores.
func ShortTypeLowerUnderscore(s string) string {
	if s == "" {
		return s
	}

	return strings.ReplaceAll(ShortTypeLower(s), " ", "_")
}
