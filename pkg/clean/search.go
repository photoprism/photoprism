package clean

import (
	"regexp"
	"strings"
)

// spaced returns the string padded with a space left and right.
func spaced(s string) string {
	return Space + s + Space
}

// replace performs a case-insensitive string replacement.
func replace(subject string, search string, replace string) string {
	return regexp.MustCompile("(?i)"+search).ReplaceAllString(subject, replace)
}

// SearchString replaces search operator with default symbols.
func SearchString(s string) string {
	if s == "" || reject(s, LengthLimit) {
		return Empty
	}

	// Normalize.
	s = strings.ReplaceAll(s, "%%", "%")
	s = strings.ReplaceAll(s, "%", "*")
	s = strings.ReplaceAll(s, "**", "*")

	// Trim.
	return strings.Trim(s, "|\\<>\n\r\t")
}

// SearchQuery replaces search operator with default symbols.
func SearchQuery(s string) string {
	if s == "" || reject(s, LengthLimit) {
		return Empty
	}

	// Normalize.
	s = replace(s, spaced(EnOr), Or)
	s = replace(s, spaced(EnOr), Or)
	s = replace(s, spaced(EnAnd), And)
	s = replace(s, spaced(EnWith), And)
	s = replace(s, spaced(EnIn), And)
	s = replace(s, spaced(EnAt), And)
	s = strings.ReplaceAll(s, SpacedPlus, And)
	s = strings.ReplaceAll(s, "%%", "%")
	s = strings.ReplaceAll(s, "%", "*")
	s = strings.ReplaceAll(s, "**", "*")

	// Trim.
	return strings.Trim(s, "|${}\\<>: \n\r\t")
}
