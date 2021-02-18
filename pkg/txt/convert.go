package txt

import (
	"regexp"
	"strconv"
	"strings"
)

var UnknownCountryCode = "zz"
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
func CountryCode(s string) (code string) {
	code = UnknownCountryCode

	if s == "" || s == UnknownCountryCode {
		return code
	}

	words := CountryWordsRegexp.FindAllString(s, -1)

	for i, w := range words {
		if i < len(words)-1 {
			search := strings.ToLower(w + " " + words[i+1])

			if match, ok := Countries[search]; ok {
				return match
			}
		}

		search := strings.ToLower(w)

		if match, ok := Countries[search]; ok {
			code = match
		}
	}

	return code
}
