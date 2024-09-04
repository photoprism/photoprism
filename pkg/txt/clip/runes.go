package clip

import "strings"

// Runes limits a string to the given number of runes and removes all leading and trailing spaces.
func Runes(s string, size int) string {
	s = strings.TrimSpace(s)

	if s == "" || size <= 0 {
		return ""
	}

	runes := []rune(s)

	if len(runes) > size {
		s = string(runes[0:size])
	}

	return strings.TrimSpace(s)
}
