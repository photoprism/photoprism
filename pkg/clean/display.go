package clean

import (
	"strings"
	"unicode"
)

// Display replaces whitespace controls with a space, removes other control characters, and marks
// the remaining non-printable characters with a question mark, for text shown in a terminal or table.
// Unlike Log, it neither quotes nor replaces punctuation, and it keeps the spaces it is given.
func Display(s string) string {
	if s == "" {
		return s
	}

	return strings.Map(func(r rune) rune {
		switch {
		case r == '\t', r == '\n', r == '\v', r == '\f', r == '\r':
			return unsafeSpace
		case unsafeDropRune(r):
			return -1
		case unicode.Is(unicode.Zs, r):
			return r
		case unsafeRune(r):
			return unsafeMarker
		default:
			return r
		}
	}, s)
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
