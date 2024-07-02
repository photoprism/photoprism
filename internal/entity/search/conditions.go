package search

import (
	"fmt"
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"

	"github.com/jinzhu/inflection"
)

// Like escapes a string for use in a query.
func Like(s string) string {
	return strings.Trim(clean.SqlString(s), " |&*%")
}

// LikeAny returns a single where condition matching the search words.
func LikeAny(col, s string, keywords, exact bool) (wheres []string) {
	if s == "" {
		return wheres
	}

	s = txt.StripOr(clean.SearchQuery(s))

	var wildcardThreshold int

	if exact {
		wildcardThreshold = -1
	} else if keywords {
		wildcardThreshold = 4
	} else {
		wildcardThreshold = 2
	}

	for _, k := range strings.Split(s, txt.And) {
		var orWheres []string
		var words []string

		if keywords {
			words = txt.UniqueKeywords(k)
		} else {
			words = txt.UniqueWords(strings.Fields(k))
		}

		if len(words) == 0 {
			continue
		}

		for _, w := range words {
			if wildcardThreshold > 0 && len(w) >= wildcardThreshold {
				orWheres = append(orWheres, fmt.Sprintf("%s LIKE '%s%%'", col, Like(w)))
			} else {
				orWheres = append(orWheres, fmt.Sprintf("%s LIKE '%s'", col, Like(w)))
			}

			if !keywords || !txt.ContainsASCIILetters(w) {
				continue
			}

			singular := inflection.Singular(w)

			if singular != w {
				orWheres = append(orWheres, fmt.Sprintf("%s LIKE '%s'", col, Like(singular)))
			}
		}

		if len(orWheres) > 0 {
			wheres = append(wheres, strings.Join(orWheres, " OR "))
		}
	}

	return wheres
}

// LikeAnyKeyword returns a single where condition matching the search keywords.
func LikeAnyKeyword(col, s string) (wheres []string) {
	return LikeAny(col, s, true, false)
}

// LikeAnyWord returns a single where condition matching the search word.
func LikeAnyWord(col, s string) (wheres []string) {
	return LikeAny(col, s, false, false)
}

// LikeAll returns a list of where conditions matching all search words.
func LikeAll(col, s string, keywords, exact bool) (wheres []string) {
	if s == "" {
		return wheres
	}

	var words []string
	var wildcardThreshold int

	if keywords {
		words = txt.UniqueKeywords(s)
		wildcardThreshold = 4
	} else {
		words = txt.UniqueWords(strings.Fields(s))
		wildcardThreshold = 2
	}

	if len(words) == 0 {
		return wheres
	} else if exact {
		wildcardThreshold = -1
	}

	for _, w := range words {
		if wildcardThreshold > 0 && len(w) >= wildcardThreshold {
			wheres = append(wheres, fmt.Sprintf("%s LIKE '%s%%'", col, Like(w)))
		} else {
			wheres = append(wheres, fmt.Sprintf("%s LIKE '%s'", col, Like(w)))
		}
	}

	return wheres
}

// LikeAllKeywords returns a list of where conditions matching all search keywords.
func LikeAllKeywords(col, s string) (wheres []string) {
	return LikeAll(col, s, true, false)
}

// LikeAllWords returns a list of where conditions matching all search words.
func LikeAllWords(col, s string) (wheres []string) {
	return LikeAll(col, s, false, false)
}

// LikeAllNames returns a list of where conditions matching all names.
func LikeAllNames(cols Cols, s string) (wheres []string) {
	if len(cols) == 0 || len(s) < 1 {
		return wheres
	}

	for _, k := range strings.Split(s, txt.And) {
		var orWheres []string

		for _, w := range strings.Split(k, txt.Or) {
			w = strings.TrimSpace(w)

			if w == txt.EmptyString {
				continue
			}

			for _, c := range cols {
				if strings.Contains(w, txt.Space) {
					orWheres = append(orWheres, fmt.Sprintf("%s LIKE '%s%%'", c, Like(w)))
				} else {
					orWheres = append(orWheres, fmt.Sprintf("%s LIKE '%%%s%%'", c, Like(w)))
				}
			}
		}

		if len(orWheres) > 0 {
			wheres = append(wheres, strings.Join(orWheres, " OR "))
		}
	}

	return wheres
}

// AnySlug returns a where condition that matches any slug in search.
func AnySlug(col, search, sep string) (where string) {
	if search == "" {
		return ""
	}

	if sep == "" {
		sep = " "
	}

	var wheres []string
	var words []string

	for _, w := range strings.Split(search, sep) {
		w = strings.TrimSpace(w)

		words = append(words, txt.Slug(w))

		if !txt.ContainsASCIILetters(w) {
			continue
		}

		singular := inflection.Singular(w)

		if singular != w {
			words = append(words, txt.Slug(singular))
		}
	}

	if len(words) == 0 {
		return ""
	}

	for _, w := range words {
		wheres = append(wheres, fmt.Sprintf("%s = '%s'", col, Like(w)))
	}

	return strings.Join(wheres, " OR ")
}

// AnyInt returns a where condition that matches any integer within a range.
func AnyInt(col, numbers, sep string, min, max int) (where string) {
	if numbers == "" {
		return ""
	}

	if sep == "" {
		sep = txt.Or
	}

	var matches []int
	var wheres []string

	for _, n := range strings.Split(numbers, sep) {
		i := txt.Int(n)

		if i == 0 || i < min || i > max {
			continue
		}

		matches = append(matches, i)
	}

	if len(matches) == 0 {
		return ""
	}

	for _, n := range matches {
		wheres = append(wheres, fmt.Sprintf("%s = %d", col, n))
	}

	return strings.Join(wheres, " OR ")
}

// OrLike returns a where condition and values for finding multiple terms combined with OR.
func OrLike(col, s string) (where string, values []interface{}) {
	if txt.Empty(col) || txt.Empty(s) {
		return "", []interface{}{}
	}

	s = strings.ReplaceAll(s, "*", "%")
	s = strings.ReplaceAll(s, "%%", "%")

	terms := strings.Split(s, txt.Or)
	values = make([]interface{}, len(terms))

	if l := len(terms); l == 0 {
		return "", []interface{}{}
	} else if l == 1 {
		values[0] = terms[0]
	} else {
		for i := range terms {
			values[i] = strings.TrimSpace(terms[i])
		}
	}

	like := fmt.Sprintf("%s LIKE ?", col)
	where = like + strings.Repeat(" OR "+like, len(terms)-1)

	return where, values
}

// Split splits a search string into separate values and trims whitespace.
func Split(s string, sep string) (result []string) {
	if s == "" {
		return []string{}
	}

	// Trim separator and split.
	s = strings.Trim(s, sep)
	v := strings.Split(s, sep)

	if len(v) <= 1 {
		return v
	}

	result = make([]string, 0, len(v))

	for i := range v {
		if t := strings.TrimSpace(v[i]); t != "" {
			result = append(result, t)
		}
	}

	return result
}

// SplitOr splits a search string into separate OR values for an IN condition.
func SplitOr(s string) (values []string) {
	return Split(s, txt.Or)
}

// SplitAnd splits a search string into separate AND values.
func SplitAnd(s string) (values []string) {
	return Split(s, txt.And)
}
