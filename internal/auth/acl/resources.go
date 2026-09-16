package acl

import "strings"

// Resources represents a list of resources, for checks that accept any one of them.
type Resources []Resource

// String returns the resources as a comma-separated string.
func (res Resources) String() string {
	s := make([]string, len(res))

	for i := range res {
		s[i] = res[i].String()
	}

	return strings.Join(s, ", ")
}

// First returns the first resource. An empty list defaults to ResourceDefault.
func (res Resources) First() Resource {
	if len(res) == 0 {
		return ResourceDefault
	}

	return res[0]
}
