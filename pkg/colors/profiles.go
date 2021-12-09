package colors

import "strings"

type Profile string

// Supported color profiles.
const (
	Default          Profile = ""
	ProfileDisplayP3 Profile = "Display P3"
)

// Equal compares the color profile name case-insensitively.
func (p Profile) Equal(s string) bool {
	return strings.EqualFold(string(p), s)
}
