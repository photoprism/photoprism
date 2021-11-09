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

// QueryTooShort tests if a search query is too short.
func QueryTooShort(q string) bool {
	q = strings.Trim(q, "- '")

	return q != Empty && len(q) < 3 && IsLatin(q)
}
