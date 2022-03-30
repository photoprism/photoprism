package list

// Excludes tests if a string is not contained in the list.
func Excludes(list []string, s string) bool {
	if len(list) == 0 || s == "" {
		return false
	}

	return !Contains(list, s)
}

// ExcludesAny tests if two lists exclude each other.
func ExcludesAny(l, s []string) bool {
	if len(l) == 0 || len(s) == 0 {
		return false
	}

	return !ContainsAny(l, s)
}
