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

// SQLParam cleans and preps a string for use as a parameter in a query.  Pre and Post are used to add a wild card character for like's.
func SQLParam(s, pre, post string) string {
	return pre + strings.Trim(s, " |&*%") + post
}

// LikeAny returns a slice of or'd where conditions matching the search words, and slices of the variable parameters for each where condition.
// It returns a slice of where statements, and slices of slice the parameter values.
// Expectation is that each set of results will be fed into gorm.Expr
// eg. gorm.Expr(wheres[0], valuesSlice[0]...)
func LikeAny(col, s string, keywords, exact bool) (wheres []string, valuesSlice [][]interface{}) {
	if s == "" {
		return wheres, valuesSlice
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
		var orValues []interface{}
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
				orValues = append(orValues, SQLParam(w, "", "%"))
			} else {
				orWheres = append(orWheres, fmt.Sprintf("%s LIKE ?", col))
				orValues = append(orValues, SQLParam(w, "", ""))
			}

			if !keywords || !txt.ContainsASCIILetters(w) {
				continue
			}

			singular := inflection.Singular(w)

			if singular != w {
				orWheres = append(orWheres, fmt.Sprintf("%s LIKE ?", col))
				orValues = append(orValues, SQLParam(singular, "", ""))
			}
		}

		if len(orWheres) > 0 {
			wheres = append(wheres, strings.Join(orWheres, " OR "))
			valuesSlice = append(valuesSlice, orValues)
		}
	}

	return wheres, valuesSlice
}

// LikeAnyKeyword returns a slice of or'd where conditions matching the search keywords, and slices of the variable parameters for each where condition.
// It returns a slice of where statements, and slices of slice the parameter values.
// Expectation is that each set of results will be fed into gorm.Expr
// eg. gorm.Expr(wheres[0], valuesSlice[0]...)
func LikeAnyKeyword(col, s string) (wheres []string, valuesSlice [][]interface{}) {
	return LikeAny(col, s, true, false)
}

// LikeAnyWord returns a returns a slice of or'd where conditions matching the search word, and slices of the variable parameters for each where condition.
// It returns a slice of where statements, and slices of slice the parameter values.
// Expectation is that each set of results will be fed into gorm.Expr
// eg. gorm.Expr(wheres[0], valuesSlice[0]...)
func LikeAnyWord(col, s string) (wheres []string, valuesSlice [][]interface{}) {
	return LikeAny(col, s, false, false)
}

// LikeAll returns a slice of where conditions matching all the search words, and slices of the variable parameters for each where condition.
// It returns a slice of where statements, and slices of slice the parameter values.
// Expectation is that each set of results will be fed into gorm.Expr
// eg. gorm.Expr(wheres[0], valuesSlice[0]...)
func LikeAll(col, s string, keywords, exact bool) (wheres []string, valuesSlice [][]interface{}) {
	if s == "" {
		return wheres, valuesSlice
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
		return wheres, valuesSlice
	} else if exact {
		wildcardThreshold = -1
	}

	for _, w := range words {
		var value []interface{}
		if wildcardThreshold > 0 && len(w) >= wildcardThreshold {
			wheres = append(wheres, fmt.Sprintf("%s LIKE ?", col))
			value = append(value, SQLParam(w, "", "%"))
		} else {
			wheres = append(wheres, fmt.Sprintf("%s LIKE ?", col))
			value = append(value, SQLParam(w, "", ""))
		}
		valuesSlice = append(valuesSlice, value)
	}

	return wheres, valuesSlice
}

// LikeAllKeywords returns a slice of where conditions matching all the search keywords, and slices of the variable parameters for each where condition.
// It returns a slice of where statements, and slices of slice the parameter values.
// Expectation is that each set of results will be fed into gorm.Expr
// eg. gorm.Expr(wheres[0], valuesSlice[0]...)
func LikeAllKeywords(col, s string) (wheres []string, valuesSlice [][]interface{}) {
	return LikeAll(col, s, true, false)
}

// LikeAllWords returns a slice of where conditions matching all the search words, and slices of the variable parameters for each where condition.
// It returns a slice of where statements, and slices of slice the parameter values.
// Expectation is that each set of results will be fed into gorm.Expr
// eg. gorm.Expr(wheres[0], valuesSlice[0]...)
func LikeAllWords(col, s string) (wheres []string, valuesSlice [][]interface{}) {
	return LikeAll(col, s, false, false)
}

// LikeAllNames returns a slice of where conditions matching all the names, and slices of the variable parameters for each where condition.
// It returns a slice of where statements, and slices of slice the parameter values.
// Expectation is that each set of results will be fed into gorm.Expr
// eg. gorm.Expr(wheres[0], valuesSlice[0]...)
func LikeAllNames(cols Cols, s string) (wheres []string, valuesSlice [][]interface{}) {
	if len(cols) == 0 || len(s) < 1 {
		return wheres, valuesSlice
	}

	for _, k := range txt.UnTrimmedSplitWithEscape(s, txt.AndRune, txt.EscapeRune) {
		var orWheres []string
		var orValues []interface{}

		for _, w := range txt.UnTrimmedSplitWithEscape(k, txt.OrRune, txt.EscapeRune) {
			w = strings.TrimSpace(w)

			if w == txt.EmptyString {
				continue
			}

			for _, c := range cols {
				if strings.Contains(w, txt.Space) {
					orWheres = append(orWheres, fmt.Sprintf("%s LIKE ?", c))
					orValues = append(orValues, SQLParam(w, "", "%"))
				} else {
					orWheres = append(orWheres, fmt.Sprintf("%s LIKE ?", c))
					orValues = append(orValues, SQLParam(w, "%", "%"))
				}
			}
		}

		if len(orWheres) > 0 {
			wheres = append(wheres, strings.Join(orWheres, " OR "))
			valuesSlice = append(valuesSlice, orValues)
		}
	}

	return wheres, valuesSlice
}

// AnySlug returns a where condition that matches any slug in the search and the slice of the variable parameters for the where condition.
// It returns a where statement, and a slice of the parameter values.
// Expectation is that each set of results will be fed into gorm.Expr
// eg. gorm.Expr(where, values...)
func AnySlug(col, search, sep string) (where string, values []interface{}) {
	if search == "" {
		return "", values
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
		return "", values
	}

	for _, w := range words {
		wheres = append(wheres, fmt.Sprintf("%s = ?", col))
		values = append(values, SQLParam(w, "", ""))
	}

	return strings.Join(wheres, " OR "), values
}

// AnyInt returns a where condition that matches any integer in the search and the slice of the variable parameters for the where condition.
// It returns a where statement, and a slice of the parameter values.
// Expectation is that each set of results will be fed into gorm.Expr
// eg. gorm.Expr(where, values...)
func AnyInt(col, numbers, sep string, low, high int) (where string, values []interface{}) {
	if numbers == "" {
		return "", values
	}

	if sep == "" {
		sep = txt.Or
	}

	var matches []int
	var wheres []string

	for _, n := range strings.Split(numbers, sep) {
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

// OrLike returns a where condition and values for finding multiple terms combined with OR and the slice of the variable parameters for the where condition.
// It returns a where statement, and a slice of the parameter values.
// Expectation is that each set of results will be fed into gorm.Expr
// eg. gorm.Expr(where, values...)
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

// OrLikeCols returns a where condition and values for finding multiple terms combined with OR and the slice of the variable parameters for the where condition.
// It returns a where statement, and a slice of the parameter values.
// Expectation is that each set of results will be fed into gorm.Expr
// eg. gorm.Expr(where, values...)
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

// SplitOr splits a search string into separate OR values for an IN condition.
func SplitOr(s string) (values []string) {
	return txt.TrimmedSplitWithEscape(s, txt.OrRune, txt.EscapeRune)
}

// SplitAnd splits a search string into separate AND values.
func SplitAnd(s string) (values []string) {
	return txt.TrimmedSplitWithEscape(s, txt.AndRune, txt.EscapeRune)
}
