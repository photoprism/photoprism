package txt

import (
	"regexp"
	"sort"
	"strings"
)

var KeywordsRegexp = regexp.MustCompile("[\\p{L}\\d\\-']{1,}")

// UnknownWord returns true if the string does not seem to be a real word.
func UnknownWord(s string) bool {
	if len(s) > 3 || !ContainsASCIILetters(s) {
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
	if s == "" {
		return results
	}

	for _, w := range KeywordsRegexp.FindAllString(s, -1) {
		w = strings.Trim(w, "- '")

		if w == "" || len(w) < 2 && IsLatin(w) || IsNumeric(w) {
			continue
		}

		results = append(results, w)
	}

	return results
}

// Keywords returns a slice of keywords without stopwords but including dashes.
func Keywords(s string) (results []string) {
	if s == "" {
		return results
	}

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

// ReplaceSpaces replaces all spaces with another string.
func ReplaceSpaces(s string, char string) string {
	return strings.Replace(s, " ", char, -1)
}

var FilenameKeywordsRegexp = regexp.MustCompile("[\\p{L}]{1,}")

// FilenameWords returns a slice of words with at least 3 characters from a string ("ile", "france").
func FilenameWords(s string) (results []string) {
	if s == "" {
		return results
	}

	for _, s := range FilenameKeywordsRegexp.FindAllString(s, -1) {
		if len(s) < 3 && IsLatin(s) {
			continue
		}

		results = append(results, s)
	}

	return results
}

// FilenameKeywords returns a slice of keywords without stopwords.
func FilenameKeywords(s string) (results []string) {
	if s == "" {
		return results
	}

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

// UniqueWords sorts and filters a string slice for unique words.
func UniqueWords(words []string) (results []string) {
	last := ""

	SortCaseInsensitive(words)

	for _, w := range words {
		w = strings.Trim(strings.ToLower(w), "- '")

		if w == "" || len(w) < 2 && IsLatin(w) || w == last {
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

		if len(w) < 2 && IsLatin(w) || w == last || strings.Contains(remove, w) {
			continue
		}

		last = w

		results = append(results, w)
	}

	return results
}

// AddToWords add words to a string slice and returns the sorted result.
func AddToWords(existing []string, words string) []string {
	w := Words(words)

	if len(w) < 1 {
		return existing
	}

	return UniqueWords(append(existing, w...))
}

// MergeWords merges two keyword strings separated by ", ".
func MergeWords(w1, w2 string) string {
	return strings.Join(AddToWords(Words(w1), w2), ", ")
}

// UniqueKeywords returns a slice of unique and sorted keywords without stopwords.
func UniqueKeywords(s string) (results []string) {
	if s == "" {
		return results
	}

	last := ""

	words := Keywords(s)

	SortCaseInsensitive(words)

	for _, w := range words {
		w = strings.ToLower(w)

		if len(w) < 3 && IsLatin(w) || w == last {
			continue
		}

		last = w

		results = append(results, w)
	}

	return results
}

// SortCaseInsensitive performs a case-insensitive slice sort.
func SortCaseInsensitive(words []string) {
	sort.Slice(words, func(i, j int) bool { return strings.ToLower(words[i]) < strings.ToLower(words[j]) })
}

// StopwordsOnly tests if the string contains stopwords only.
func StopwordsOnly(s string) bool {
	s = strings.TrimSpace(s)

	if s == "" {
		return false
	}

	for _, w := range Words(s) {
		w = strings.ToLower(w)

		if UnknownWord(w) {
			continue
		}

		if _, ok := StopWords[w]; ok == false {
			return false
		}
	}

	return true
}
