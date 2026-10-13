package clean

import (
	"strings"
	"unicode"
)

// Type replaces whitespace controls with a space, omits other control and non-ASCII characters,
// and ensures a maximum length of 64 characters.
func Type(s string) string {
	if s == "" {
		return s
	}

	result, _, _ := typeString(s, LengthType)

	return result
}

// typeString keeps printable ASCII, replaces whitespace controls with a space, drops other bytes, and limits
// the trimmed result to maxLen bytes. It reports whether a character inside the trimmed input was removed or
// replaced, and whether the result was shortened.
func typeString(s string, maxLen int) (result string, changed, clipped bool) {
	s = strings.TrimSpace(s)

	var dst []byte

	for i := 0; i < len(s); i++ {
		b := s[i]

		if b >= 0x20 && b < 0x7F {
			if dst != nil {
				dst = append(dst, b)
			}

			continue
		}

		if dst == nil {
			dst = make([]byte, i, len(s))
			copy(dst, s[:i])
		}

		switch b {
		case '\t', '\n', '\v', '\f', '\r':
			dst = append(dst, ' ')
		}

		changed = true
	}

	if dst != nil {
		s = strings.TrimSpace(string(dst))
	}

	if len(s) > maxLen {
		s = strings.TrimSpace(s[:maxLen])
		clipped = true
	}

	return s, changed, clipped
}

// TypeUnicode collapses whitespace, removes control, non-printable and other unsafe characters
// except U+200C and U+200D, and limits the result to 64 characters while keeping non-ASCII letters.
func TypeUnicode(s string) string {
	if s == "" {
		return s
	}

	buf := make([]rune, 0, len([]rune(s)))
	lastWasSpace := false

	for _, r := range s {
		if len(buf) >= LengthType {
			break
		}

		if unicode.IsSpace(r) {
			if len(buf) == 0 || lastWasSpace {
				continue
			}
			buf = append(buf, ' ')
			lastWasSpace = true
			continue
		}

		if unsafeDropRune(r) || unsafeRune(r) {
			continue
		}

		switch r {
		case '`', '\\', '|', '"', '\'', '?', '*', '<', '>', '{', '}':
			continue
		}

		buf = append(buf, r)
		lastWasSpace = false
	}

	for len(buf) > 0 && unicode.IsSpace(buf[len(buf)-1]) {
		buf = buf[:len(buf)-1]
	}

	return string(buf)
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

// ShortType replaces whitespace controls with a space, omits other control and non-ASCII characters,
// and ensures a maximum length of 8 characters.
func ShortType(s string) string {
	if s == "" {
		return s
	}

	result, _, _ := typeString(s, LengthShortType)

	return result
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
