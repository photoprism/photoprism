package txt

import (
	"strings"
)

// Bool casts a string to bool.
func Bool(s string) bool {
	s = strings.TrimSpace(s)

	if s == "" || No(s) {
		return false
	}

	return true
}

// Yes returns true if a string represents "yes".
func Yes(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))

	return strings.IndexAny(s, "ytjosiд") == 0
}

// No returns true if a string represents "no".
func No(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))

	return strings.IndexAny(s, "0nhfeн") == 0
}
