package list

// Any matches everything.
const Any = "*"

// All is kept for backward compatibility, but deprecated.
const All = Any

// Contains tests if a string is contained in the list.
func Contains(list []string, s string) bool {
	if len(list) == 0 || s == "" {
		return false
	} else if s == Any {
		return true
	}

	// Find matches.
	for i := range list {
		if s == list[i] || list[i] == Any {
			return true
		}
	}

	return false
}

// ContainsAny tests if two lists have at least one common entry.
func ContainsAny(l, s []string) bool {
	if len(l) == 0 || len(s) == 0 {
		return false
	}

	// If second list contains All, it's a wildcard match.
	if s[0] == Any {
		return true
	}
	for j := 1; j < len(s); j++ {
		if s[j] == Any {
			return true
		}
	}

	// Build a set from the smaller slice for O(n+m) intersection.
	a, b := l, s
	if len(a) > len(b) {
		a, b = b, a
	}
	set := make(map[string]struct{}, len(a))
	for i := range a {
		set[a[i]] = struct{}{}
	}
	for j := range b {
		if _, ok := set[b[j]]; ok {
			return true
		}
	}
	return false
}
