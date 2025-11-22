package txt

import (
	"github.com/photoprism/photoprism/pkg/txt/clip"
)

const (
	// Ellipsis contains the typographic ellipsis character.
	Ellipsis = "…"
	// ClipCountry limits country codes.
	ClipCountry = 2
	// ClipRole limits role names.
	ClipRole = 32
	// ClipPasscode limits passcodes.
	ClipPasscode = 36
	// ClipKeyword limits keyword length.
	ClipKeyword = 40
	// ClipIP limits IP address strings.
	ClipIP = 48
	// ClipRealm limits realm names.
	ClipRealm = 64
	// ClipUsername limits usernames.
	ClipUsername = 64
	// ClipPassword limits passwords.
	ClipPassword = 72
	// ClipSlug limits URL slugs.
	ClipSlug = 80
	// ClipType limits type labels.
	ClipType = 100
	// ClipCategory limits category names.
	ClipCategory = 100
	// ClipTokenName limits token names.
	ClipTokenName = 128
	// ClipDefault is the default clipping length.
	ClipDefault = 160
	// ClipName limits standard names.
	ClipName = 160
	// ClipLongName limits long names.
	ClipLongName = 200
	// ClipError limits error strings.
	ClipError = 255
	// ClipEmail limits email addresses.
	ClipEmail = 255
	// ClipPath limits filesystem paths.
	ClipPath = 500
	// ClipComment limits comment strings.
	ClipComment = 512
	// ClipURL limits URLs.
	ClipURL = 512
	// ClipLog limits log messages.
	ClipLog = 512
	// ClipFlags limits combined CLI flags.
	ClipFlags = 767
	// ClipShortText limits short text blocks.
	ClipShortText = 1024
	// ClipText limits medium text blocks.
	ClipText = 2048
	// ClipLongText limits long text blocks.
	ClipLongText = 4096
)

// Clip limits a string to the given number of runes and removes all leading and trailing spaces.
func Clip(s string, size int) string {
	return clip.Runes(s, size)
}

// Shorten limits a character string to the specified number of runes and adds a suffix if it has been shortened.
func Shorten(s string, size int, suffix string) string {
	return clip.Shorten(s, size, suffix)
}
