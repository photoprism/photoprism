/*
This package contains PhotoPrism database queries.

Additional information can be found in our Developer Guide:

https://github.com/photoprism/photoprism/wiki
*/
package query

import (
	"github.com/photoprism/photoprism/internal/event"

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
