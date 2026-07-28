package clean

import (
	"path/filepath"
	"strings"
	"unicode"
)

// FileNameRedacted returns a file name with its last dash-separated segment replaced by a placeholder,
// keeping the extension so entries stay recognizable. Use it for generated, single-use names.
func FileNameRedacted(s string) string {
	ext := filepath.Ext(s)
	name := strings.TrimSuffix(s, ext)

	if i := strings.LastIndex(name, "-"); i > 0 {
		return name[:i+1] + "***" + ext
	}

	return "***" + ext
}

// FileName removes invalid character from a filename string.
func FileName(s string) string {
	if s == "" || reject(s, 512) || strings.Contains(s, "/") || strings.Contains(s, "..") {
		return ""
	}

	// Trim whitespace.
	s = strings.TrimSpace(s)

	// Remove non-printable and other potentially problematic characters.
	// The following characters must never be used in a filename: / \ > < | : &
	s = strings.Map(func(r rune) rune {
		if !unicode.IsPrint(r) {
			return -1
		}

		switch r {
		case '~', '/', '\\', ':', '|', '"', '?', '*', '<', '>', '{', '}':
			return -1
		default:
			return r
		}
	}, s)

	if s == "." || s == ".." {
		return ""
	}

	return s
}
