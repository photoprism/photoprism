package txt

import (
	"strings"
)

const (
	Or     = "|"
	And    = "&"
	Plus   = " + "
	OrEn   = " or "
	AndEn  = " and "
	WithEn = " with "
	InEn   = " in "
	AtEn   = " at "
	Space  = " "
	Empty  = ""
)

// NormalizeQuery replaces search operator with default symbols.
func NormalizeQuery(s string) string {
	s = strings.ToLower(Clip(s, ClipQuery))
	s = strings.ReplaceAll(s, OrEn, Or)
	s = strings.ReplaceAll(s, AndEn, And)
	s = strings.ReplaceAll(s, WithEn, And)
	s = strings.ReplaceAll(s, InEn, And)
	s = strings.ReplaceAll(s, AtEn, And)
	s = strings.ReplaceAll(s, Plus, And)
	s = strings.ReplaceAll(s, "%", "*")
	return strings.Trim(s, "+&|_-=!@$%^(){}\\<>,.;: ")
}

// QueryTooShort tests if a search query is too short.
func QueryTooShort(q string) bool {
	q = strings.Trim(q, "- '")

	return q != Empty && len(q) < 3 && IsLatin(q)
}
