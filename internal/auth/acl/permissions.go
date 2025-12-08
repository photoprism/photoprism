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

// First returns the first permission as a string. When the slice is empty it defaults to ActionUse.
func (perm Permissions) First() string {
	if len(perm) == 0 {
		return ActionUse.String()
	}

	return perm[0].String()
}

// Contains reports whether the specified permission or wildcard is present in this set.
func (perm Permissions) Contains(p Permission) bool {
	if len(perm) == 0 || p == "" {
		return false
	} else if p == Any {
		return true
	}

	// Find matches.
	for i := range perm {
		if p == perm[i] || perm[i] == Any {
			return true
		}
	}

	return false
}
