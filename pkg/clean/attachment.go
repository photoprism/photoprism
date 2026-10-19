package clean

import (
	"path/filepath"
	"strings"
	"unicode"
)

// AttachmentName returns a file name for a download or archive entry: whitespace controls become a space,
// path separators and other characters Windows does not allow in names become an underscore, other control
// and non-printable characters are removed, and the extension and non-ASCII characters are kept. Leading
// dots are removed, and a name without a stem is named "download".
func AttachmentName(s string) string {
	s = strings.Map(func(r rune) rune {
		switch {
		case r == '\t', r == '\n', r == '\v', r == '\f', r == '\r':
			return ' '
		case r == '/', r == '\\', r == ':', r == '*', r == '?', r == '"', r == '<', r == '>', r == '|':
			return '_'
		case unsafeDropRune(r):
			return -1
		case unicode.Is(unicode.Zs, r):
			return r
		case unsafeRune(r):
			return -1
		default:
			return r
		}
	}, s)

	s = strings.TrimSpace(s)

	if ext := filepath.Ext(s); len(ext) > 1 && strings.TrimLeft(strings.TrimSuffix(s, ext), ". ") == "" {
		return "download" + ext
	} else if s = strings.TrimLeft(s, ". "); s == "" {
		return "download"
	}

	return s
}
