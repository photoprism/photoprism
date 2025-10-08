package txt

import (
	"regexp"
	"strings"
)

var UnknownStateCode = "zz"
var UnknownCountryCode = "zz"
var CountryWordsRegexp = regexp.MustCompile("[\\p{L}]{2,}")

// AmbiguousCountries contains location keywords that also occur as popular given names.
// They require additional context before we can safely treat them as country hints.
var AmbiguousCountries = map[string]string{
	"vienna":       "at",
	"london":       "gb",
	"sydney":       "au",
	"dallas":       "us",
	"houston":      "us",
	"orlando":      "us",
	"jordan":       "jo",
	"rafic hariri": "lb",
	"sofia":        "bg",
	"sana'a":       "ye",
	"sanaa":        "ye",
	"sana a":       "ye",
	"riad":         "sa",
	"riyadh":       "sa",
	"milan":        "it",
	"venice":       "it",
	"trinidad":     "tt",
	"valencia":     "es",
	"alberta":      "ca",
	"ben gurion":   "il",
	"haifa":        "il",
	"paris":        "fr",
	"chad":         "td",
	"samaria":      "ps",
}

// CountryCode attempts to find a matching country code for a given string.
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
