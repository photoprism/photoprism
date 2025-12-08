package search

import (
	"fmt"
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"

	"github.com/jinzhu/inflection"
)

// Like sanitizes user input so it can be safely interpolated into SQL LIKE
// expressions. It strips operators that we don't expect to persist in the
// statement and lets callers provide their own surrounding wildcards.
func Like(s string) string {
	return strings.Trim(clean.SqlString(s), " |&*%")
}

// LikeAny builds OR-chained LIKE predicates for a text column. The input string
// may contain AND / OR separators; keywords trigger stemming and plural
// normalization while exact mode disables wildcard suffixes.
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

	for _, k := range txt.UnTrimmedSplitWithEscape(s, txt.AndRune, txt.EscapeRune) {
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

// LikeAnyKeyword is a keyword-optimized wrapper around LikeAny.
func LikeAnyKeyword(col, s string) (wheres []string) {
	return LikeAny(col, s, true, false)
}

// LikeAnyWord matches whole words and keeps wildcard thresholds tuned for
// free-form text search instead of keyword lists.
func LikeAnyWord(col, s string) (wheres []string) {
	return LikeAny(col, s, false, false)
}

// LikeAll produces AND-chained LIKE predicates for every significant token in
// the search string. When exact is false, longer words receive a suffix
// wildcard to support prefix matches.
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

// LikeAllKeywords is LikeAll specialized for keyword search.
func LikeAllKeywords(col, s string) (wheres []string) {
	return LikeAll(col, s, true, false)
}

// LikeAllWords is LikeAll specialized for general word search.
func LikeAllWords(col, s string) (wheres []string) {
	return LikeAll(col, s, false, false)
}

// LikeAllNames splits a name query into AND-separated groups and generates
// prefix or substring matches against each provided column, keeping multi-word
// tokens intact so "John Doe" still matches full-name columns.
func LikeAllNames(cols Cols, s string) (wheres []string) {
	if len(cols) == 0 || len(s) < 1 {
		return wheres
	}

	for _, k := range txt.UnTrimmedSplitWithEscape(s, txt.AndRune, txt.EscapeRune) {
		var orWheres []string

		for _, w := range txt.UnTrimmedSplitWithEscape(k, txt.OrRune, txt.EscapeRune) {
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

// AnySlug converts human-friendly search terms into slugs and matches them
// against the provided slug column, including the singularized variant for
// plural words (e.g. "Cats" -> "cat").
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

// AnyInt filters user-specified integers through an allowed range and returns
// an OR-chained equality predicate for the values that remain.
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

// OrLike prepares a parameterised OR/LIKE clause for a single column. Star (* )
// wildcards are mapped to SQL percent wildcards before returning the query and
// bind values.
func OrLike(col, s string) (where string, values []interface{}) {
	if txt.Empty(col) || txt.Empty(s) {
		return "", []interface{}{}
	}

	s = strings.ReplaceAll(s, "*", "%")
	s = strings.ReplaceAll(s, "%%", "%")

	terms := txt.UnTrimmedSplitWithEscape(s, txt.OrRune, txt.EscapeRune)
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

// OrLikeCols behaves like OrLike but fans out the same search terms across
// multiple columns, preserving the order of values so callers can feed them to
// database/sql.
func OrLikeCols(cols []string, s string) (where string, values []interface{}) {
	if len(cols) == 0 || txt.Empty(s) {
		return "", []interface{}{}
	}

	s = strings.ReplaceAll(s, "*", "%")
	s = strings.ReplaceAll(s, "%%", "%")

	terms := txt.UnTrimmedSplitWithEscape(s, txt.OrRune, txt.EscapeRune)

	if len(terms) == 0 {
		return "", []interface{}{}
	}

	values = make([]interface{}, len(terms)*len(cols))

	for j := range terms {
		for i := range cols {
			values[j+i] = strings.TrimSpace(terms[j])
		}
	}

	wheres := make([]string, len(cols))

	for i, col := range cols {
		for j := range terms {
			k := len(terms) * i
			values[j+k] = terms[j]
		}
		like := fmt.Sprintf("%s LIKE ?", col)
		wheres[i] = like + strings.Repeat(" OR "+like, len(terms)-1)
	}

	return strings.Join(wheres, " OR "), values
}

// SplitOr splits a search string on OR separators (|) while respecting escape
// sequences so literals like "\|" survive unchanged.
func SplitOr(s string) (values []string) {
	return txt.TrimmedSplitWithEscape(s, txt.OrRune, txt.EscapeRune)
}

// SplitAnd splits a search string on AND separators (&) while honouring escape
// sequences.
func SplitAnd(s string) (values []string) {
	return txt.TrimmedSplitWithEscape(s, txt.AndRune, txt.EscapeRune)
}
