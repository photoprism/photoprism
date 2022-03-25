package sanitize

import (
	"bytes"
)

// sqlSpecialBytes contains special bytes to escape in SQL search queries.
// see https://mariadb.com/kb/en/string-literals/
var sqlSpecialBytes = []byte{34, 39, 92, 95} // ", ', \, _

// SqlString escapes a string for use in an SQL query.
func SqlString(s string) string {
	var i int
	for i = 0; i < len(s); i++ {
		if bytes.Contains(sqlSpecialBytes, []byte{s[i]}) {
			break
		}
	}

	// No special characters found, return original string.
	if i >= len(s) {
		return s
	}

	b := make([]byte, 2*len(s)-i)
	copy(b, s[:i])
	j := i
	for ; i < len(s); i++ {
		if s[i] < 31 {
			// Ignore control chars.
			continue
		}
		if bytes.Contains(sqlSpecialBytes, []byte{s[i]}) {
			b[j] = '\\'
			j++
		}
		b[j] = s[i]
		j++
	}
	return string(b[:j])
}
