package query

import (
	"fmt"
	"strings"

	"github.com/gosimple/slug"
	"github.com/photoprism/photoprism/pkg/txt"

	"github.com/jinzhu/inflection"
)

// LikeAny returns a single where condition matching the search keywords.
func LikeAny(col, keywords string) (wheres []string) {
	keywords = strings.ReplaceAll(keywords, Or, " ")
	keywords = strings.ReplaceAll(keywords, OrEn, " ")
	keywords = strings.ReplaceAll(keywords, AndEn, And)

	for _, k := range strings.Split(keywords, And) {
		var orWheres []string

		words := txt.UniqueKeywords(k)

		if len(words) == 0 {
			continue
		}

		for _, w := range words {
			if len(w) > 3 {
				orWheres = append(orWheres, fmt.Sprintf("%s LIKE '%s%%'", col, w))
			} else {
				orWheres = append(orWheres, fmt.Sprintf("%s = '%s'", col, w))
			}

			if !txt.ContainsASCIILetters(w) {
				continue
			}

			singular := inflection.Singular(w)

			if singular != w {
				orWheres = append(orWheres, fmt.Sprintf("%s = '%s'", col, singular))
			}
		}

		if len(orWheres) > 0 {
			wheres = append(wheres, strings.Join(orWheres, " OR "))
		}
	}

	return wheres
}

// LikeAll returns a list of where conditions matching all search keywords.
func LikeAll(col, keywords string) (wheres []string) {
	words := txt.UniqueKeywords(keywords)

	if len(words) == 0 {
		return wheres
	}

	for _, w := range words {
		if len(w) > 3 {
			wheres = append(wheres, fmt.Sprintf("%s LIKE '%s%%'", col, w))
		} else {
			wheres = append(wheres, fmt.Sprintf("%s = '%s'", col, w))
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
