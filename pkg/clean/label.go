package clean

import (
	"strings"
	"unicode"

	"github.com/photoprism/photoprism/pkg/txt"

	"github.com/gosimple/slug"
)

// LabelName cleans a string so that it can be compared against other LabelNames to ensure
// that LabelSlug clashes are only created appropriately
func LabelName(s string) string {
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "@", "at")
	s = strings.ReplaceAll(s, "&", "and")
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return s
	}
	if s[0] == txt.SlugEncoded && txt.ContainsAlnumLower(s[1:]) {
		return txt.Clip(s, txt.ClipSlug)
	}

	if slug.Make(s) == "" {
		return s
	}

	cleanLabel := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) || unicode.IsLetter(r) || unicode.IsSpace(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, s)

	return cleanLabel
}
