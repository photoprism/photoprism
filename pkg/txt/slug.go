package txt

import (
	"strings"

	"github.com/gosimple/slug"
)

// Slug converts a string to a valid slug with a max length of 80 runes.
func Slug(s string) string {
	s = strings.TrimSpace(s)

	if s == "" {
		return ""
	}

	return Clip(slug.Make(s), ClipSlug)
}

// SlugToTitle converts a slug back to a title
func SlugToTitle(s string) string {
	if s == "" {
		return ""
	}

	return Title(strings.Join(Words(s), " "))
}
