package list

// Add adds a string to the list if it does not exist yet.
func Add(list []string, s string) []string {
	if s == "" {
		return list
	} else if len(list) == 0 {
		return []string{s}
	} else if Contains(list, s) {
		return list
	}

	return append(list, s)
}
