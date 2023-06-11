package txt

import (
	"strings"
)

const (
	Ellipsis      = "…"
	ClipCountry   = 2
	ClipRole      = 32
	ClipKeyword   = 40
	ClipIP        = 48
	ClipRealm     = 64
	ClipUserName  = 64
	ClipPassword  = 72
	ClipSlug      = 80
	ClipCategory  = 100
	ClipTokenName = 128
	ClipDefault   = 160
	ClipName      = 160
	ClipLongName  = 200
	ClipEmail     = 255
	ClipPath      = 500
	ClipComment   = 512
	ClipURL       = 512
	ClipLog       = 512
	ClipFlags     = 767
	ClipShortText = 1024
	ClipText      = 2048
	ClipLongText  = 4096
)

// Clip shortens a string to the given number of runes, and removes all leading and trailing white space.
func Clip(s string, size int) string {
	s = strings.TrimSpace(s)

	if s == "" || size <= 0 {
		return ""
	}

	runes := []rune(s)

	if len(runes) > size {
		s = string(runes[0:size])
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
