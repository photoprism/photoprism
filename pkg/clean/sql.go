package clean

// SqlSpecial checks if the byte must be escaped/omitted in SQL.
func SqlSpecial(b byte) (special bool, omit bool) {
	if b < 32 {
		return true, true
	}

	switch b {
	case '"', '\'', '\\':
		return true, false
	default:
		return false, false
	}
}

// SqlClean strips bytes that must be omitted from an SQL parameter value
// (currently control characters below 0x20). Quote characters are preserved
// unchanged because the caller is expected to pass the result through a
// parameterized query, where escaping is the driver's responsibility.
func SqlClean(s string) string {
	var i int
	for i = 0; i < len(s); i++ {
		if _, omit := SqlSpecial(s[i]); omit {
			break
		}
	}

	// Return unchanged if no omittable characters were found.
	if i >= len(s) {
		return s
	}

	b := make([]byte, len(s))

	copy(b, s[:i])

	j := i

	for ; i < len(s); i++ {
		if _, omit := SqlSpecial(s[i]); omit {
			continue
		}
		b[j] = s[i]
		j++
	}

	return string(b[:j])
}
