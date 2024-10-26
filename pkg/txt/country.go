package txt

import (
	"regexp"
	"strings"
)

var UnknownStateCode = "zz"
var UnknownCountryCode = "zz"
var CountryWordsRegexp = regexp.MustCompile("[\\p{L}]{2,}")

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
