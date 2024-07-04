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

// Yes tests if a string represents "yes".
func Yes(s string) bool {
	if s == "" {
		return false
	}

	s = strings.ToLower(strings.TrimSpace(s))

	return strings.IndexAny(s, "ytjposiд") == 0
}

// No tests if a string represents "no".
func No(s string) bool {
	if s == "" {
		return false
	}

	s = strings.ToLower(strings.TrimSpace(s))

	return strings.IndexAny(s, "0nhufeн") == 0
}

// New tests if a string represents "new".
func New(s string) bool {
	if s == "" {
		return false
	}

	s = strings.ToLower(strings.TrimSpace(s))

	return s == EnNew
}
