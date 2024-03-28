package acl

import "strings"

// Permissions represents a list of permissions.
type Permissions []Permission

// String returns the permissions as a comma-separated string.
func (perm Permissions) String() string {
	s := make([]string, len(perm))

	for i := range perm {
		s[i] = perm[i].String()
	}

	return strings.Join(s, ", ")
}
