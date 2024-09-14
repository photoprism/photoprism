package clean

// ASCII removes all non-ascii characters from a string and returns it.
func ASCII(s string) string {
	if s == "" {
		return ""
	}

	result := make([]rune, 0, len(s))

	for _, r := range s {
		if r <= 127 {
			result = append(result, r)
		}
	}

	return string(result)
}
