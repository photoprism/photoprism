package clean

import (
	"fmt"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

// JSONEscape returns JSON text with DEL and other characters that do not print, except spaces and
// U+200C and U+200D, as well as invalid bytes, written as \u escapes, so that the decoded values stay exact. encoding/json
// already escapes C0 controls in strings, and the characters escaped here occur only in strings.
func JSONEscape(data []byte) []byte {
	for i := 0; i < len(data); {
		r, size := utf8.DecodeRune(data[i:])

		if !jsonEscapeRune(r, size) {
			i += size
			continue
		}

		out := make([]byte, 0, len(data)+16)
		out = append(out, data[:i]...)

		for j := i; j < len(data); {
			r, size = utf8.DecodeRune(data[j:])

			if r == utf8.RuneError && size <= 1 {
				out = append(out, `\ufffd`...)
			} else if !jsonEscapeRune(r, size) {
				out = append(out, data[j:j+size]...)
			} else if r1, r2 := utf16.EncodeRune(r); r1 != unicode.ReplacementChar {
				out = fmt.Appendf(out, `\u%04x\u%04x`, r1, r2)
			} else {
				out = fmt.Appendf(out, `\u%04x`, r)
			}

			j += size
		}

		return out
	}

	return data
}

// jsonEscapeRune reports whether a decoded rune is written as an escape by JSONEscape. C0 controls are
// left as they are: encoding/json escapes them in strings, so any in its output is whitespace between values.
// An invalid byte is written as \ufffd, the escape json.Marshal uses for invalid UTF-8 in strings.
func jsonEscapeRune(r rune, size int) bool {
	switch {
	case r == utf8.RuneError && size <= 1:
		return true
	case r < 0x20:
		return false
	case r == 0x7F:
		return true
	case unsafeSpaceRune(r):
		return false
	default:
		return unsafeRune(r)
	}
}
