package search

import (
	"fmt"
	"strings"

	"github.com/photoprism/photoprism/pkg/sanitize"

	"github.com/gosimple/slug"
	"github.com/photoprism/photoprism/pkg/txt"

	"github.com/jinzhu/inflection"
)

// LikeAny returns a single where condition matching the search words.
func LikeAny(col, s string, keywords, exact bool) (wheres []string) {
	if s == "" {
		return wheres
	}

	s = txt.StripOr(sanitize.SearchQuery(s))

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
			words = txt.UniqueWords(txt.Words(k))
		}

		if len(words) == 0 {
			continue
		}

		for _, w := range words {
			if wildcardThreshold > 0 && len(w) >= wildcardThreshold {
				orWheres = append(orWheres, fmt.Sprintf("%s LIKE '%s%%'", col, w))
			} else {
				orWheres = append(orWheres, fmt.Sprintf("%s LIKE '%s'", col, w))
			}

			if !keywords || !txt.ContainsASCIILetters(w) {
				continue
			}

			singular := inflection.Singular(w)

			if singular != w {
				orWheres = append(orWheres, fmt.Sprintf("%s LIKE '%s'", col, singular))
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
		words = txt.UniqueWords(txt.Words(s))
		wildcardThreshold = 2
	}

	if len(words) == 0 {
		return wheres
	} else if exact {
		wildcardThreshold = -1
	}

	for _, w := range words {
		if wildcardThreshold > 0 && len(w) >= wildcardThreshold {
			wheres = append(wheres, fmt.Sprintf("%s LIKE '%s%%'", col, w))
		} else {
			wheres = append(wheres, fmt.Sprintf("%s LIKE '%s'", col, w))
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

			if w == txt.Empty {
				continue
			}

			for _, c := range cols {
				if strings.Contains(w, txt.Space) {
					orWheres = append(orWheres, fmt.Sprintf("%s LIKE '%s%%'", c, w))
				} else {
					orWheres = append(orWheres, fmt.Sprintf("%s LIKE '%%%s%%'", c, w))
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

		words = append(words, slug.Make(w))

		if !txt.ContainsASCIILetters(w) {
			continue
		}

		singular := inflection.Singular(w)

		if singular != w {
			words = append(words, slug.Make(singular))
		}
	}

	if len(words) == 0 {
		return ""
	}

	for _, w := range words {
		wheres = append(wheres, fmt.Sprintf("%s = '%s'", col, w))
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
	if col == "" || s == "" {
		return "", []interface{}{}
	}

	s = strings.ReplaceAll(s, "*", "%")
	s = strings.ReplaceAll(s, "%%", "%")

	terms := strings.Split(s, txt.Or)
	values = make([]interface{}, len(terms))

	for i := range terms {
		values[i] = terms[i]
	}

	like := fmt.Sprintf("%s LIKE ?", col)
	where = like + strings.Repeat(" OR "+like, len(terms)-1)

	return where, values
}
