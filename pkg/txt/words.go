package txt

import (
	"regexp"
	"sort"
	"strings"
)

var KeywordsRegexp = regexp.MustCompile("[\\p{L}\\-]{3,}")

// Words returns a slice of words with at least 3 characters from a string, dashes count as character ("ile-de-france").
func Words(s string) (results []string) {
	return KeywordsRegexp.FindAllString(s, -1)
}

// ReplaceSpaces replaces all spaces with another string.
func ReplaceSpaces(s string, char string) string {
	return strings.Replace(s, " ", char, -1)
}

var FilenameKeywordsRegexp = regexp.MustCompile("[\\p{L}]{3,}")

// FilenameWords returns a slice of words with at least 3 characters from a string ("ile", "france").
func FilenameWords(s string) (results []string) {
	return FilenameKeywordsRegexp.FindAllString(s, -1)
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

// UniqueWords sorts and filters a string slice for unique words.
func UniqueWords(words []string) (results []string) {
	last := ""

	sort.Strings(words)

	for _, w := range words {
		if len(w) < 3 || w == last {
			continue
		}

		last = w

		results = append(results, w)
	}

	return results
}

// UniqueKeywords returns a slice of unique and sorted keywords without stopwords.
func UniqueKeywords(s string) (results []string) {
	last := ""

	words := Keywords(s)

	sort.Strings(words)

	for _, w := range words {
		w = strings.ToLower(w)

		if len(w) < 3 || w == last {
			continue
		}

		last = w

		results = append(results, w)
	}

	return results
}
