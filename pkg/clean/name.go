package clean

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/txt"
)

// Name sanitizes a name string and clips it to at most txt.ClipDefault
// runes; injection patterns (`${`, `ldap://`) drop the value entirely.
// The cap is character-counted — callers that store the result in a
// byte-counted column must apply their own byte clip.
func Name(name string) string {
	if name == "" || reject(name, 0) {
		return ""
	}

	var prev rune

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

	return txt.Clip(name, txt.ClipDefault)
}

// NameCapitalized sanitizes and capitalizes a name.
func NameCapitalized(name string) string {
	return txt.Title(Name(name))
}

// DlName sanitizes a download name string.
func DlName(name string) string {
	if name == "" {
		return ""
	}

	name = strings.ReplaceAll(name, "...", "…")

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
		case '.', '|', '?', '"', '$', '%', '/', '\\', '*', '`', ':', ';', '<', '>', '{', '}':
			return -1
		}
		return r
	}, name)

	// Shorten string if longer than 255 runes.
	if name = strings.TrimSpace(name); name == "" {
		return ""
	} else if runes := []rune(name); len(runes) > txt.ClipPath {
		name = string(runes[0:txt.ClipPath])
	}

	return strings.TrimSpace(name)
}
