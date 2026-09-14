package clean

import "unicode"

// unsafeSpace replaces a space that is not one.
const unsafeSpace = ' '

// unsafeMarker replaces a character that is removed for being invisible, so that removing one
// cannot join what it separated or hide that anything was there.
const unsafeMarker = '?'

// unsafeDropRune reports whether a rune is dropped outright rather than marked. These are the
// characters the sanitizers have always removed, and their absence is what callers expect.
func unsafeDropRune(r rune) bool {
	return r < 0x20 || r == 0x7F
}

// unsafeRune reports whether a rune must not reach operator-facing text. Everything Go does not
// consider printable qualifies, which covers the control, format, private-use and unassigned
// categories without naming each one. U+200C and U+200D are the exception, because both form
// letters and emoji in names people really have.
func unsafeRune(r rune) bool {
	if r == 0x200C || r == 0x200D {
		return false
	}

	return !unicode.IsPrint(r)
}

// unsafeSpaceRune reports whether a rune renders as a space without being one, so that two
// different names cannot render identically. It is checked before unsafeRune, which would
// otherwise mark these: Go considers no space printable except the ASCII one.
func unsafeSpaceRune(r rune) bool {
	return r != ' ' && unicode.Is(unicode.Zs, r)
}
