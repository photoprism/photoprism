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
