package clean

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/txt"
)

// Name sanitizes a name string.
func Name(name string) string {
	// Empty or too long?
	if name == "" || reject(name, txt.ClipDefault) {
		return ""
	}

	var prev rune

	// Remove unwanted characters.
	name = strings.Map(func(r rune) rune {
		if r == ' ' && (prev == 0 || prev == ' ') {
			return -1
		}

		prev = r

		if r < 32 || r == 127 {
			return -1
		}

		switch r {
		case '"', '$', '%', '\\', '*', '`', ';', '<', '>', '{', '}':
			return -1
		}
		return r
	}, name)

	// OK?
	if name = strings.TrimSpace(name); name == "" {
		return ""
	}

	// Make sure name isn't too long.
	return txt.Clip(name, txt.ClipDefault)
}

// NameCapitalized sanitizes and capitalizes a name.
func NameCapitalized(name string) string {
	return txt.Title(Name(name))
}
