package clean

import "strings"

var durationRunes = map[rune]bool{
	':': true,
	'-': true,
	'd': true,
	'h': true,
	'm': true,
	's': true,
	'n': true,
	'µ': true,
}

// Duration removes non-duration characters from a string and returns the result.
func Duration(s string) string {
	if s == "" {
		return ""
	}

	valid := false
	skipDot := false

	// Remove invalid characters.
	s = strings.Map(func(r rune) rune {
		if !skipDot && (r == ',' || r == '.') {
			skipDot = true
			return '.'
		} else if durationRunes[r] {
			skipDot = false
			return r
		} else if r < '0' || r > '9' {
			return -1
		}

		valid = true

		return r
	}, s)

	if !valid {
		return ""
	}

	return s
}
