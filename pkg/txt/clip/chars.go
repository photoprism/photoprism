package clip

import "strings"

// Chars limits a string to the specified number of characters and removes all leading and trailing spaces.
func Chars(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	l := len(s)

	if l <= maxLen {
		return s
	} else {
		return strings.TrimSpace(s[:maxLen])
	}
}
