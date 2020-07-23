package txt

import (
	"regexp"
	"strings"
)

var ContainsNumberRegexp = regexp.MustCompile("\\d+")

// ContainsNumber returns true if string contains a number.
func ContainsNumber(s string) bool {
	return ContainsNumberRegexp.MatchString(s)
}

// Bool casts a string to bool.
func Bool(s string) bool {
	s = strings.TrimSpace(s)

	if s == "" || s == "0" || s == "false" || s == "no" {
		return false
	}

	return true
}

// ASCII returns true if the string only contains ascii chars without whitespace, numbers, and punctuation marks.
func ASCII(s string) bool {
	for _, r := range s {
		if (r < 65 || r > 90) && (r < 97 || r > 122) {
			return false
		}
	}

	return true
}
