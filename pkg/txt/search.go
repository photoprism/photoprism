package txt

// SearchTerms returns a bool map with all terms as key.
func SearchTerms(s string) map[string]bool {
	result := make(map[string]bool)

	if s == "" {
		return result
	}

	for _, w := range UniqueKeywords(s) {
		result[w] = true
	}

	return result
}
