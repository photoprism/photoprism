package clean

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// unsafeSpace replaces a space that is not one.
const unsafeSpace = ' '

// markRune is what a function passed to markText returns for a rune it wants replaced with its marker.
const markRune rune = -2

// markerDigits are the hexadecimal digits of a marker.
const markerDigits = "0123456789ABCDEF"

// unsafeDropRune reports whether a rune is a C0 control or DEL, which the sanitizers for stored values
// remove outright.
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

// unsafePathRune reports whether a rune is removed from file and path names: C0 and C1 controls, DEL,
// line and paragraph separators, and bidi embeddings, overrides and isolates. Every other character may
// occur in a real name and is kept, so that a cleaned name still matches the file system.
func unsafePathRune(r rune) bool {
	switch {
	case r < 0x20, r >= 0x7F && r <= 0x9F:
		return true
	case r == 0x2028, r == 0x2029:
		return true
	case r >= 0x202A && r <= 0x202E, r >= 0x2066 && r <= 0x2069:
		return true
	}

	return false
}

// unsafeSpaceRune reports whether a rune renders as a space without being one, so that two
// different names cannot render identically. It is checked before unsafeRune, which would
// otherwise mark these: Go considers no space printable except the ASCII one.
func unsafeSpaceRune(r rune) bool {
	return r != ' ' && unicode.Is(unicode.Zs, r)
}

// unsafeBreakRune reports whether a rune breaks a line or separates text like a tab, so that text shown on
// one line renders it as a space.
func unsafeBreakRune(r rune) bool {
	switch r {
	case '\t', '\n', '\v', '\f', '\r', 0x85, 0x2028, 0x2029:
		return true
	}

	return false
}

// markText returns s with each rune replaced by what f returns for it, where markRune stands for the
// rune's "[U+XXXX]" marker. A byte that is not valid UTF-8 is replaced with its "[0xXX]" marker. f may be
// called twice for a rune, so it must not count calls.
func markText(s string, f func(r rune) rune) string {
	i := 0

	for i < len(s) {
		r, n := utf8.DecodeRuneInString(s[i:])

		if (r == utf8.RuneError && n == 1) || f(r) != r {
			break
		}

		i += n
	}

	if i == len(s) {
		return s
	}

	var b strings.Builder

	b.Grow(len(s) + 16)
	b.WriteString(s[:i])

	for i < len(s) {
		r, n := utf8.DecodeRuneInString(s[i:])

		switch m := f(r); {
		case r == utf8.RuneError && n == 1:
			b.WriteString("[0x")
			b.WriteByte(markerDigits[s[i]>>4])
			b.WriteByte(markerDigits[s[i]&0xF])
			b.WriteByte(']')
		case m == markRune:
			writeMarker(&b, r)
		case m == r:
			b.WriteString(s[i : i+n])
		default:
			b.WriteRune(m)
		}

		i += n
	}

	return b.String()
}

// writeMarker writes the "[U+XXXX]" marker of a rune, with four to six hexadecimal digits.
func writeMarker(b *strings.Builder, r rune) {
	digits := 4

	for digits < 6 && r>>(4*digits) > 0 {
		digits++
	}

	b.WriteString("[U+")

	for d := digits - 1; d >= 0; d-- {
		b.WriteByte(markerDigits[(r>>(4*d))&0xF])
	}

	b.WriteByte(']')
}

// markerLen returns the length of the marker that s starts with, or 0 if it starts with none.
func markerLen(s string) int {
	if len(s) < 6 || s[0] != '[' {
		return 0
	}

	switch s[1:3] {
	case "U+":
		for i := 3; i < len(s) && i <= 9; i++ {
			if s[i] == ']' {
				if i >= 7 {
					return i + 1
				}

				return 0
			} else if !strings.ContainsRune(markerDigits, rune(s[i])) {
				return 0
			}
		}
	case "0x":
		if strings.ContainsRune(markerDigits, rune(s[3])) && strings.ContainsRune(markerDigits, rune(s[4])) && s[5] == ']' {
			return 6
		}
	}

	return 0
}

// markedUnitLen returns the length of the marker or character that s starts with, so that shortening
// sanitized text never splits a marker.
func markedUnitLen(s string) int {
	if n := markerLen(s); n > 0 {
		return n
	}

	_, n := utf8.DecodeRuneInString(s)

	return n
}
