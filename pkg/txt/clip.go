package txt

import (
	"github.com/photoprism/photoprism/pkg/txt/clip"
)

const (
	Ellipsis      = "…"
	ClipCountry   = 2
	ClipRole      = 32
	ClipPasscode  = 36
	ClipKeyword   = 40
	ClipIP        = 48
	ClipRealm     = 64
	ClipUsername  = 64
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

// Clip limits a string to the given number of runes and removes all leading and trailing spaces.
func Clip(s string, size int) string {
	return clip.Runes(s, size)
}

// Shorten limits a character string to the specified number of runes and adds a suffix if it has been shortened.
func Shorten(s string, size int, suffix string) string {
	return clip.Shorten(s, size, suffix)
}
