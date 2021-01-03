package txt

import (
	"regexp"
	"unicode"
)

var ContainsNumberRegexp = regexp.MustCompile("\\d+")

// ContainsNumber returns true if string contains a number.
func ContainsNumber(s string) bool {
	return ContainsNumberRegexp.MatchString(s)
}

// ContainsLetters reports whether the string only contains letters.
func ContainsLetters(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}

	return true
}

// ContainsASCIILetters reports if the string only contains ascii chars without whitespace, numbers, and punctuation marks.
func ContainsASCIILetters(s string) bool {
	for _, r := range s {
		if (r < 65 || r > 90) && (r < 97 || r > 122) {
			return false
		}
	}

	return true
}

// ContainsSymbols reports whether the string only contains symbolic characters.
func ContainsSymbols(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if !unicode.IsSymbol(r) {
			return false
		}
	}

	return true
}
