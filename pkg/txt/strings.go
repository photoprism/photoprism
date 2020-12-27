package txt

import (
	"strings"
)

// Bool casts a string to bool.
func Bool(s string) bool {
	s = strings.TrimSpace(s)

	if s == "" || s == "0" || s == "false" || s == "no" {
		return false
	}

	return true
}
