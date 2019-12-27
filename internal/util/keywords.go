package util

import (
	"regexp"
	"strings"
)

var KeywordsRegexp = regexp.MustCompile("[\\p{L}]{3,}")

func Keywords(s string) (results []string) {
	all := KeywordsRegexp.FindAllString(s, -1)

	for _, w := range all {
		w = strings.ToLower(w)

		if _, ok := Stopwords[w]; ok == false {
			results = append(results, w)
		}
	}

	return results
}
