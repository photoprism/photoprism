/*
This package contains PhotoPrism database queries.

Additional information can be found in our Developer Guide:

https://github.com/photoprism/photoprism/wiki
*/
package query

import (
	"fmt"
	"strings"

	"github.com/gosimple/slug"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/txt"

	"github.com/jinzhu/gorm"
)

var log = event.Log

// About 1km ('good enough' for now)
const SearchRadius = 0.009

// Query searches given an originals path and a db instance.
type Query struct {
	db *gorm.DB
}

// SearchCount is the total number of search hits.
type SearchCount struct {
	Total int
}

// New returns a new Query type with a given path and db instance.
func New(db *gorm.DB) *Query {
	q := &Query{
		db: db,
	}

	return q
}

// Db returns a database connection instance.
func Db() *gorm.DB {
	return entity.Db()
}

// UnscopedDb returns an unscoped database connection instance.
func UnscopedDb() *gorm.DB {
	return entity.Db().Unscoped()
}

// LikeAny returns a where condition that matches any keyword in search.
func LikeAny(col, search string) (where string) {
	var wheres []string

	words := txt.UniqueKeywords(search)

	if len(words) == 0 {
		return ""
	}

	for _, w := range words {
		if len(w) > 3 {
			wheres = append(wheres, fmt.Sprintf("%s LIKE '%s%%'", col, w))
		} else {
			wheres = append(wheres, fmt.Sprintf("%s = '%s'", col, w))
		}
	}

	return strings.Join(wheres, " OR ")
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
		words = append(words, slug.Make(strings.TrimSpace(w)))
	}

	if len(words) == 0 {
		return ""
	}

	for _, w := range words {
		wheres = append(wheres, fmt.Sprintf("%s = '%s'", col, w))
	}

	return strings.Join(wheres, " OR ")
}
