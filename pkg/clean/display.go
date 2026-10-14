package clean

import (
	"strings"
	"unicode"
)

// Display replaces line breaks and tabs with a space and writes other control and non-printable characters
// as a "[U+XXXX]" marker, and a byte that is not valid UTF-8 as "[0xXX]", for text shown in a terminal or
// table. Unlike Log, it neither quotes nor replaces punctuation, and it keeps the spaces it is given.
func Display(s string) string {
	return markText(s, displayRune)
}

// DisplayRunes renders s like Display, shortened to at most size characters like ClipRunes. It reads at most
// 4*size bytes of s after leading spaces, since each character of s shows as at least one character.
func DisplayRunes(s string, size int) string {
	if size <= 0 {
		return ""
	}

	return ClipRunes(Display(shortenBytes(strings.TrimLeftFunc(s, unicode.IsSpace), 4*size, "")), size)
}

// displayRune returns what Display writes for a rune.
func displayRune(r rune) rune {
	switch {
	case r >= ' ' && r < 0x7F:
		return r
	case unsafeBreakRune(r):
		return unsafeSpace
	case unicode.Is(unicode.Zs, r):
		return r
	case unsafeRune(r):
		return markRune
	default:
		return r
	}
}

// DisplayLines applies Display to each line of s, so that multi-line text such as an error message keeps its
// line breaks.
func DisplayLines(s string) string {
	if !strings.Contains(s, "\n") {
		return Display(s)
	}

	lines := strings.Split(s, "\n")

	for i, line := range lines {
		lines[i] = Display(strings.TrimSuffix(line, "\r"))
	}

	return strings.Join(lines, "\n")
}
