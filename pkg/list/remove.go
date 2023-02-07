package list

// Remove removes a string from a list and returns it.
func Remove(list []string, s string) []string {
	if len(list) == 0 || s == "" {
		return list
	} else if s == All {
		return []string{}
	}

	result := make([]string, 0, len(list))

	// Find matches.
	for i := range list {
		if s != list[i] {
			result = append(result, list[i])
		}
	}

	return result
}
