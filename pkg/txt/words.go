package txt

import (
	"regexp"
	"sort"
	"strings"
)

var KeywordsRegexp = regexp.MustCompile("[\\p{L}\\-]{3,}")

// UnknownWord returns true if the string does not seem to be a real word.
func UnknownWord(s string) bool {
	if len(s) > 3 || !ASCII(s) {
		return false
	}

	s = strings.ToLower(s)

	if _, ok := ShortWords[s]; ok {
		return false
	}

	if _, ok := SpecialWords[s]; ok {
		return false
	}

	return true
}

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

// FilenameKeywords returns a slice of keywords without stopwords.
func FilenameKeywords(s string) (results []string) {
	for _, w := range FilenameWords(s) {
		w = strings.ToLower(w)

		if UnknownWord(w) {
			continue
		}

		if _, ok := StopWords[w]; ok == false {
			results = append(results, w)
		}
	}

	return results
}

// Keywords returns a slice of keywords without stopwords but including dashes.
func Keywords(s string) (results []string) {
	for _, w := range Words(s) {
		w = strings.ToLower(w)

		if UnknownWord(w) {
			continue
		}

		if _, ok := StopWords[w]; ok == false {
			results = append(results, w)
		}
	}

	return results
}

// UniqueWords sorts and filters a string slice for unique words.
func UniqueWords(words []string) (results []string) {
	last := ""

	SortCaseInsensitive(words)

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

// RemoveFromWords removes words from a string slice and returns the sorted result.
func RemoveFromWords(words []string, remove string) (results []string) {
	remove = strings.ToLower(remove)
	last := ""

	SortCaseInsensitive(words)

	for _, w := range words {
		w = strings.ToLower(w)

		if len(w) < 3 || w == last || strings.Contains(remove, w) {
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

	SortCaseInsensitive(words)

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

// Sorts string slice case insensitive.
func SortCaseInsensitive(words []string) {
	sort.Slice(words, func(i, j int) bool { return strings.ToLower(words[i]) < strings.ToLower(words[j]) })
}
