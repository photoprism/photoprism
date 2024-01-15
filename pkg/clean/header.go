package clean

// Header sanitizes a string for use in request or response headers.
func Header(s string) string {
	if s == "" || len(s) > MaxLength {
		return ""
	}

	result := make([]rune, 0, len(s))

	for _, r := range s {
		if r > 32 && r < 127 {
			result = append(result, r)
		}
	}

	return string(result)
}
