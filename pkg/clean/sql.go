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

// SqlString escapes a string for use in an SQL query.
func SqlString(s string) string {
	var i int
	for i = 0; i < len(s); i++ {
		if found, _ := SqlSpecial(s[i]); found {
			break
		}
	}

	// Return if no special characters were found.
	if i >= len(s) {
		return s
	}

	b := make([]byte, 2*len(s)-i)

	copy(b, s[:i])

	j := i

	for ; i < len(s); i++ {
		if special, omit := SqlSpecial(s[i]); omit {
			// Omit control characters.
			continue
		} else if special {
			// Escape other special characters.
			// see https://mariadb.com/kb/en/string-literals/
			b[j] = s[i]
			j++
		}

		b[j] = s[i]
		j++
	}

	return string(b[:j])
}
