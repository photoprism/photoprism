package txt

import (
	"regexp"
	"strconv"
	"strings"
)

var CountryWordsRegexp = regexp.MustCompile("[\\p{L}]{2,}")

// Int returns a string as int or 0 if it can not be converted.
func Int(s string) int {
	if s == "" {
		return 0
	}

	result, err := strconv.ParseInt(s, 10, 64)

	if err != nil {
		return 0
	}

	return int(result)
}

// IsUInt returns true if a string only contains an unsigned integer.
func IsUInt(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if r < 48 || r > 57 {
			return false
		}
	}

	return true
}

// CountryCode tries to find a matching country code for a given string e.g. from a file oder directory name.
func CountryCode(s string) string {
	if s == "zz" {
		return "zz"
	}

	words := CountryWordsRegexp.FindAllString(s, -1)

	for i, w := range words {
		if i < len(words)-1 {
			search := strings.ToLower(w + " " + words[i+1])

			if code, ok := Countries[search]; ok {
				return code
			}
		}

		search := strings.ToLower(w)

		if code, ok := Countries[search]; ok {
			return code
		}
	}

	return "zz"
}
