package projection

import (
	"strings"
)

// Type represents a visual projection type.
type Type string

// String returns the type as string.
func (t Type) String() string {
	return string(t)
}

// Unknown checks if the type is unknown.
func (t Type) Unknown() bool {
	return t == Unknown
}

// Equal checks if the type matches.
func (t Type) Equal(s string) bool {
	return strings.EqualFold(s, t.String())
}

// NotEqual checks if the type is different.
func (t Type) NotEqual(s string) bool {
	return !t.Equal(s)
}
