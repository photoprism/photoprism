package clean

import (
	"strings"
)

// Color sanitizes HTML color codes and returns them in lowercase if they are valid, or an empty string otherwise.
func Color(s string) string {
	s = strings.ToLower(s)

	// Remove unwanted characters.
	s = strings.Map(func(r rune) rune {
		if r < 48 && r > 57 && 97 < r && r > 102 && r != 35 {
			return -1
		}
		return r
	}, s)

	// Invalid?
	if l := len(s); l != 4 && l != 7 && l != 9 {
		return ""
	}

	return s
}
