package clean

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/txt"
)

// Name returns the sanitized and capitalized names.
func Name(name string) string {
	// Empty or too long?
	if name == "" || reject(name, txt.ClipDefault) {
		return ""
	}

	// Remove unwanted characters.
	name = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		switch r {
		case '"', '$', '%', '\\', '*', '`', ';', '<', '>', '{', '}':
			return -1
		}
		return r
	}, name)

	name = strings.TrimSpace(name)

	// Now empty?
	if name == "" {
		return ""
	}

	// Shorten and capitalize.
	return txt.Clip(txt.Title(name), txt.ClipDefault)
}
