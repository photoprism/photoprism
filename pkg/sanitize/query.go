package sanitize

import "strings"

// spaced returns the string padded with a space left and right.
func spaced(s string) string {
	return Space + s + Space
}

// Query replaces search operator with default symbols.
func Query(s string) string {
	if s == "" || len(s) > 1024 || strings.Contains(s, "${") {
		return Empty
	}

	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, spaced(EnOr), Or)
	s = strings.ReplaceAll(s, spaced(EnAnd), And)
	s = strings.ReplaceAll(s, spaced(EnWith), And)
	s = strings.ReplaceAll(s, spaced(EnIn), And)
	s = strings.ReplaceAll(s, spaced(EnAt), And)
	s = strings.ReplaceAll(s, SpacedPlus, And)
	s = strings.ReplaceAll(s, "%", "*")

	return strings.Trim(s, "+&|_-=!@$%^(){}\\<>,.;: ")
}
