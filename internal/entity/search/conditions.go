package search

import (
	"fmt"
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"

	"github.com/jinzhu/inflection"
)

// SqlParam sanitizes user input for use as a LIKE-clause bind value. The
// surrounding pre/post strings are concatenated verbatim so callers can add
// SQL wildcards (e.g. "%") without exposing the underlying value to string
// interpolation.
func SqlParam(s, pre, post string) string {
	return pre + strings.Trim(clean.SqlClean(s), " |&*%") + post
}

// LikeAny builds OR-chained LIKE predicates for a text column. The input string
// may contain AND / OR separators; keywords trigger stemming and plural
// normalization while exact mode disables wildcard suffixes.
// The returned wheres and values are aligned 1:1; callers feed each pair into
// gorm.Expr(wheres[i], values[i]...).
func LikeAny(col, s string, keywords, exact bool) (wheres []string, values [][]any) {
	if s == "" {
		return wheres, values
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
		var orValues []any
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
				orWheres = append(orWheres, fmt.Sprintf("%s LIKE ?", col))
				orValues = append(orValues, SqlParam(w, "", "%"))
			} else {
				orWheres = append(orWheres, fmt.Sprintf("%s LIKE ?", col))
				orValues = append(orValues, SqlParam(w, "", ""))
			}

			if !keywords || !txt.ContainsASCIILetters(w) {
				continue
			}

			singular := inflection.Singular(w)

			if singular != w {
				orWheres = append(orWheres, fmt.Sprintf("%s LIKE ?", col))
				orValues = append(orValues, SqlParam(singular, "", ""))
			}
		}

		if len(orWheres) > 0 {
			wheres = append(wheres, strings.Join(orWheres, " OR "))
			values = append(values, orValues)
		}
	}

	return wheres, values
}

// LikeAnyKeyword is a keyword-optimized wrapper around LikeAny.
func LikeAnyKeyword(col, s string) (wheres []string, values [][]any) {
	return LikeAny(col, s, true, false)
}

// LikeAnyWord matches whole words and keeps wildcard thresholds tuned for
// free-form text search instead of keyword lists.
func LikeAnyWord(col, s string) (wheres []string, values [][]any) {
	return LikeAny(col, s, false, false)
}

// LikeAll produces AND-chained LIKE predicates for every significant token in
// the search string. When exact is false, longer words receive a suffix
// wildcard to support prefix matches.
// The returned wheres and values are aligned 1:1; callers feed each pair into
// gorm.Expr(wheres[i], values[i]...).
func LikeAll(col, s string, keywords, exact bool) (wheres []string, values [][]any) {
	if s == "" {
		return wheres, values
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
		return wheres, values
	} else if exact {
		wildcardThreshold = -1
	}

	for _, w := range words {
		if wildcardThreshold > 0 && len(w) >= wildcardThreshold {
			wheres = append(wheres, fmt.Sprintf("%s LIKE ?", col))
			values = append(values, []any{SqlParam(w, "", "%")})
		} else {
			wheres = append(wheres, fmt.Sprintf("%s LIKE ?", col))
			values = append(values, []any{SqlParam(w, "", "")})
		}
	}

	return wheres, values
}

// LikeAllKeywords is LikeAll specialized for keyword search.
func LikeAllKeywords(col, s string) (wheres []string, values [][]any) {
	return LikeAll(col, s, true, false)
}

// LikeAllWords is LikeAll specialized for general word search.
func LikeAllWords(col, s string) (wheres []string, values [][]any) {
	return LikeAll(col, s, false, false)
}

// LikeAllNames splits a name query into AND-separated groups and generates
// prefix or substring matches against each provided column, keeping multi-word
// tokens intact so "John Doe" still matches full-name columns.
func LikeAllNames(cols Cols, s string) (wheres []string, values [][]any) {
	if len(cols) == 0 || len(s) < 1 {
		return wheres, values
	}

	for _, k := range txt.UnTrimmedSplitWithEscape(s, txt.AndRune, txt.EscapeRune) {
		var orWheres []string
		var orValues []any

		for _, w := range txt.UnTrimmedSplitWithEscape(k, txt.OrRune, txt.EscapeRune) {
			w = strings.TrimSpace(w)

			if w == txt.EmptyString {
				continue
			}

			for _, c := range cols {
				if strings.Contains(w, txt.Space) {
					orWheres = append(orWheres, fmt.Sprintf("%s LIKE ?", c))
					orValues = append(orValues, SqlParam(w, "", "%"))
				} else {
					orWheres = append(orWheres, fmt.Sprintf("%s LIKE ?", c))
					orValues = append(orValues, SqlParam(w, "%", "%"))
				}
			}
		}

		if len(orWheres) > 0 {
			wheres = append(wheres, strings.Join(orWheres, " OR "))
			values = append(values, orValues)
		}
	}

	return wheres, values
}

// AnySlug converts human-friendly search terms into slugs and matches them
// against the provided slug column, including the singularized variant for
// plural words (e.g. "Cats" -> "cat").
func AnySlug(col, search, sep string) (where string, values []any) {
	if search == "" {
		return "", values
	}

	if sep == "" {
		sep = " "
	}

	var wheres []string
	var words []string

	for w := range strings.SplitSeq(search, sep) {
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
		return "", values
	}

	for _, w := range words {
		wheres = append(wheres, fmt.Sprintf("%s = ?", col))
		values = append(values, SqlParam(w, "", ""))
	}

	return strings.Join(wheres, " OR "), values
}

// AnyInt filters user-specified integers through an allowed range and returns
// an OR-chained equality predicate for the values that remain. Named low/high
// to avoid shadowing the predeclared min/max identifiers added in Go 1.21.
func AnyInt(col, numbers, sep string, low, high int) (where string, values []any) {
	if numbers == "" {
		return "", values
	}

	if sep == "" {
		sep = txt.Or
	}

	var matches []int
	var wheres []string

	for n := range strings.SplitSeq(numbers, sep) {
		i := txt.Int(n)

		if i == 0 || i < low || i > high {
			continue
		}

		matches = append(matches, i)
	}

	if len(matches) == 0 {
		return "", values
	}

	for _, n := range matches {
		wheres = append(wheres, fmt.Sprintf("%s = ?", col))
		values = append(values, n)
	}

	return strings.Join(wheres, " OR "), values
}

// OrLike prepares a parameterized OR/LIKE clause for a single column. Star (* )
// wildcards are mapped to SQL percent wildcards before returning the query and
// bind values.
func OrLike(col, s string) (where string, values []any) {
	if txt.Empty(col) || txt.Empty(s) {
		return "", []any{}
	}

	s = strings.ReplaceAll(s, "*", "%")
	s = strings.ReplaceAll(s, "%%", "%")

	terms := txt.UnTrimmedSplitWithEscape(s, txt.OrRune, txt.EscapeRune)
	values = make([]any, len(terms))

	if l := len(terms); l == 0 {
		return "", []any{}
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
func OrLikeCols(cols []string, s string) (where string, values []any) {
	if len(cols) == 0 || txt.Empty(s) {
		return "", []any{}
	}

	s = strings.ReplaceAll(s, "*", "%")
	s = strings.ReplaceAll(s, "%%", "%")

	terms := txt.UnTrimmedSplitWithEscape(s, txt.OrRune, txt.EscapeRune)

	if len(terms) == 0 {
		return "", []any{}
	}

	values = make([]any, len(terms)*len(cols))

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

// SplitAnd splits a search string on AND separators (&) while honoring escape
// sequences.
func SplitAnd(s string) (values []string) {
	return txt.TrimmedSplitWithEscape(s, txt.AndRune, txt.EscapeRune)
}
