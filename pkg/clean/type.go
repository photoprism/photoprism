package clean

import (
	"strings"
	"unicode"

	"github.com/photoprism/photoprism/pkg/txt/clip"
)

// Type omits invalid runes, ensures a maximum length of 64 characters, and returns the result.
func Type(s string) string {
	if s == "" {
		return s
	}

	return clip.Chars(ASCII(s), LengthType)
}

// TypeUnderscore replaces whitespace, dividers, quotes, brackets, and other special characters with an underscore.
func TypeUnderscore(s string) string {
	if s == "" {
		return s
	}

	s = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return '_'
		}

		switch r {
		case '-', '`', '~', '\\', '|', '"', '\'', '?', '*', '<', '>', '{', '}':
			return '_'
		default:
			return r
		}
	}, s)

	return s
}

// TypeDash replaces whitespace, dividers, quotes, brackets, and other special characters with a dash.
func TypeDash(s string) string {
	if s == "" {
		return s
	}

	s = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return '-'
		}

		switch r {
		case '_', '`', '~', '\\', '|', '"', '\'', '?', '*', '<', '>', '{', '}':
			return '-'
		default:
			return r
		}
	}, s)

	return s
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

	return TypeUnderscore(TypeLower(s))
}

// TypeLowerDash converts a string to a lowercase type string and replaces spaces with dashes.
func TypeLowerDash(s string) string {
	if s == "" {
		return s
	}

	return TypeDash(TypeLower(s))
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

	return TypeUnderscore(ShortTypeLower(s))
}

// ShortTypeLowerDash converts a string to a short lowercase type string and replaces spaces with dashes.
func ShortTypeLowerDash(s string) string {
	if s == "" {
		return s
	}

	return TypeDash(ShortTypeLower(s))
}
