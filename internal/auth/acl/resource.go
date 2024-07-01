package acl

import "strings"

// Resource represents a resource for which roles can be granted Permission.
type Resource string

// String returns the type as string.
func (r Resource) String() string {
	if r == "" {
		return "default"
	}

	return string(r)
}

// LogId returns an identifier string for use in log messages.
func (r Resource) LogId() string {
	return r.String()
}

// Equal checks if the type matches.
func (r Resource) Equal(s string) bool {
	return strings.EqualFold(s, r.String())
}

// NotEqual checks if the type is different.
func (r Resource) NotEqual(s string) bool {
	return !r.Equal(s)
}
