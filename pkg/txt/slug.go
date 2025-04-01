package txt

import (
	"encoding/base32"
	"strings"

	"github.com/gosimple/slug"
)

const SlugCharset = "abcdefghijklmnopqrstuvwxyz123456"
const SlugEncoded = '_'

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
	}

	return Clip(result, ClipSlug)
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
