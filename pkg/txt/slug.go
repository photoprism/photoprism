package txt

import (
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gosimple/slug"
)

// SlugCharset defines the allowed characters for encoded slugs.
const SlugCharset = "abcdefghijklmnopqrstuvwxyz123456"

// SlugEncoded is the prefix character indicating an encoded slug.
const SlugEncoded = '_'

// SlugEncoding defines default encoding for slug generation.
var SlugEncoding = base32.NewEncoding(SlugCharset).WithPadding(base32.NoPadding)

// Slug converts a string to a valid slug with a max length of 80 runes.
func Slug(s string) string {
	s = strings.TrimSpace(s)

	if s == "" || s == "-" {
		return s
	}

	if s[0] == SlugEncoded && ContainsAlnumLower(s[1:]) {
		return Clip(s, ClipSlug)
	}

	result := slug.Make(s)

	if result == "" {
		result = string(SlugEncoded) + SlugEncoding.EncodeToString([]byte(s))
	} else if tokens := slugRuneTokens(s); tokens != "" {
		result += "-" + tokens
		result = clipSlugWithHash(result, s)
	}

	return Clip(result, ClipSlug)
}

// slugRuneTokens returns stable rune tokens for non-ASCII symbols that slug.Make would drop.
func slugRuneTokens(s string) string {
	tokens := make([]string, 0, 4)

	for _, r := range s {
		if r <= unicode.MaxASCII || slug.Make(string(r)) != "" {
			continue
		}

		switch {
		case unicode.IsSymbol(r):
		case r == '\u200d':
		case r >= '\ufe00' && r <= '\ufe0f':
		case r >= '\U0001f3fb' && r <= '\U0001f3ff':
		default:
			continue
		}

		tokens = append(tokens, fmt.Sprintf("u%x", r))
	}

	return strings.Join(tokens, "-")
}

// clipSlugWithHash clips long slugs and appends a deterministic hash suffix to avoid collisions.
func clipSlugWithHash(result, source string) string {
	if utf8.RuneCountInString(result) <= ClipSlug {
		return result
	}

	hash := sha256.Sum256([]byte(source))
	suffix := fmt.Sprintf("%x", hash[:4])
	prefixLen := ClipSlug - len(suffix) - 1

	if prefixLen <= 0 {
		return Clip(result, ClipSlug)
	}

	prefix := strings.TrimRight(Clip(result, prefixLen), "-")

	if prefix == "" {
		return Clip(result, ClipSlug)
	}

	return prefix + "-" + suffix
}

// SlugToTitle converts a slug back to a title
func SlugToTitle(s string) string {
	if s == "" {
		return ""
	}

	if s[0] == SlugEncoded {
		title, err := SlugEncoding.DecodeString(s[1:])

		if len(title) > 0 && err == nil {
			return string(title)
		}
	}

	return Title(strings.Join(Words(s), " "))
}
