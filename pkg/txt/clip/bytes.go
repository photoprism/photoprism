package clip

import "strings"

// Bytes limits a string to at most maxBytes bytes without splitting a UTF-8
// rune, and removes all leading and trailing spaces. It is intended for columns
// whose limit is measured in bytes rather than characters (e.g. VARBINARY), so a
// multi-byte value cannot overflow the column and trigger a write error.
func Bytes(s string, maxBytes int) string {
	s = strings.TrimSpace(s)

	if s == "" || maxBytes <= 0 {
		return ""
	}

	// Already within budget (also the only path that can return a string ending
	// at len(s) rather than at an interior rune boundary).
	if len(s) <= maxBytes {
		return s
	}

	// Keep the longest prefix that ends on a rune boundary at or below maxBytes.
	// Ranging over a string yields the byte index at the start of each rune, so
	// cut always lands on a complete-rune boundary.
	cut := 0
	for i := range s {
		if i > maxBytes {
			break
		}
		cut = i
	}

	return strings.TrimSpace(s[:cut])
}
