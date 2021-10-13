package txt

import (
	"strings"
)

const (
	Empty      = ""
	Space      = " "
	Or         = "|"
	And        = "&"
	Plus       = "+"
	SpacedPlus = Space + Plus + Space
)

// Spaced returns the string padded with a space left and right.
func Spaced(s string) string {
	return Space + s + Space
}

// StripOr removes or operators from a query.
func StripOr(s string) string {
	s = strings.ReplaceAll(s, Or, Space)
	return s
}

// NormalizeQuery replaces search operator with default symbols.
func NormalizeQuery(s string) string {
	s = strings.ToLower(Clip(s, ClipQuery))
	s = strings.ReplaceAll(s, Spaced(EnOr), Or)
	s = strings.ReplaceAll(s, Spaced(EnAnd), And)
	s = strings.ReplaceAll(s, Spaced(EnWith), And)
	s = strings.ReplaceAll(s, Spaced(EnIn), And)
	s = strings.ReplaceAll(s, Spaced(EnAt), And)
	s = strings.ReplaceAll(s, SpacedPlus, And)
	s = strings.ReplaceAll(s, "%", "*")
	return strings.Trim(s, "+&|_-=!@$%^(){}\\<>,.;: ")
}

// QueryTooShort tests if a search query is too short.
func QueryTooShort(q string) bool {
	q = strings.Trim(q, "- '")

	return q != Empty && len(q) < 3 && IsLatin(q)
}
