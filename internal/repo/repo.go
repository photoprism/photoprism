/*
This package contains PhotoPrism database queries.

Additional information can be found in our Developer Guide:

https://github.com/photoprism/photoprism/wiki
*/
package repo

import (
	"github.com/photoprism/photoprism/internal/event"

	"github.com/jinzhu/gorm"
)

var log = event.Log

// About 1km ('good enough' for now)
const SearchRadius = 0.009

// Repo searches given an originals path and a db instance.
type Repo struct {
	originalsPath string
	db            *gorm.DB
}

// SearchCount is the total number of search hits.
type SearchCount struct {
	Total int
}

// New returns a new Repo type with a given path and db instance.
func New(originalsPath string, db *gorm.DB) *Repo {
	instance := &Repo{
		originalsPath: originalsPath,
		db:            db,
	}

	return instance
}
