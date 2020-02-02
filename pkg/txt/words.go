package txt

import (
	"regexp"
	"strings"
)

var KeywordsRegexp = regexp.MustCompile("[\\p{L}]{3,}")

// Words returns a slice of words with at least 3 characters from a string.
func Words(s string) (results []string) {
	return KeywordsRegexp.FindAllString(s, -1)
}

// Keywords returns a slice of keywords without stopwords.
func Keywords(s string) (results []string) {
	for _, w := range Words(s) {
		w = strings.ToLower(w)

		if _, ok := Stopwords[w]; ok == false {
			results = append(results, w)
		}
	}

	return results
}
