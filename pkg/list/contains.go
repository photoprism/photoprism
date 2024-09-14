package list

const All = "*"

// Contains tests if a string is contained in the list.
func Contains(list []string, s string) bool {
	if len(list) == 0 || s == "" {
		return false
	} else if s == All {
		return true
	}

	// Find matches.
	for i := range list {
		if s == list[i] || list[i] == All {
			return true
		}
	}

	return false
}

// ContainsAny tests if two lists have at least one common entry.
func ContainsAny(l, s []string) bool {
	if len(l) == 0 || len(s) == 0 {
		return false
	} else if s[0] == All {
		return true
	}

	// Find matches.
	for i := range l {
		for j := range s {
			if s[j] == l[i] || s[j] == All {
				return true
			}
		}
	}

	// Not found.
	return false
}
