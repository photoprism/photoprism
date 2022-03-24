package txt

import (
	"strings"
)

const (
	Ellipsis        = "…"
	ClipCountryCode = 2
	ClipKeyword     = 40
	ClipUsername    = 64
	ClipSlug        = 80
	ClipCategory    = 100
	ClipPlace       = 128
	ClipDefault     = 160
	ClipName        = 160
	ClipTitle       = 200
	ClipVarchar     = 255
	ClipPath        = 500
	ClipLabel       = 500
	ClipQuery       = 1000
	ClipDescription = 16000
)

// Clip shortens a string to the given number of runes, and removes all leading and trailing white space.
func Clip(s string, size int) string {
	s = strings.TrimSpace(s)

	if s == "" || size <= 0 {
		return ""
	}

	runes := []rune(s)

	if len(runes) > size {
		s = string(runes[0 : size-1])
	}

	return strings.TrimSpace(s)
}

// Shorten shortens a string with suffix.
func Shorten(s string, size int, suffix string) string {
	if suffix == "" {
		suffix = Ellipsis
	}

	l := len(suffix)

	if len(s) < size || size < l+1 {
		return s
	}

	return Clip(s, size-l) + suffix
}
