package acl

import "strings"

// Permission represents a single ability.
type Permission string

// String returns the type as string.
func (p Permission) String() string {
	return strings.ReplaceAll(string(p), "_", " ")
}

// LogId returns an identifier string for use in log messages.
func (p Permission) LogId() string {
	return p.String()
}

// Equal checks if the type matches.
func (p Permission) Equal(s string) bool {
	return strings.EqualFold(s, p.String())
}

// NotEqual checks if the type is different.
func (p Permission) NotEqual(s string) bool {
	return !p.Equal(s)
}
