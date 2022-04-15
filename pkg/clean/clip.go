package clean

import "strings"

const (
	ClipType      = 64
	ClipShortType = 8
)

// Clip shortens a string to the given number of characters, and removes all leading and trailing white space.
func Clip(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	l := len(s)

	if l <= maxLen {
		return s
	} else {
		return strings.TrimSpace(s[:maxLen])
	}
}
