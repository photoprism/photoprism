/*

Package query contains frequently used database queries for use in commands and API.

Copyright (c) 2018 - 2020 Michael Mayer <hello@photoprism.org>

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published
    by the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.

    PhotoPrism™ is a registered trademark of Michael Mayer.  You may use it as required
    to describe our software, run your own server, for educational purposes, but not for
    offering commercial goods, products, or services without prior written permission.
    In other words, please ask.

Feel free to send an e-mail to hello@photoprism.org if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.org/developer-guide/

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
	"github.com/jinzhu/inflection"
)

var log = event.Log

const (
	MySQL  = "mysql"
	SQLite = "sqlite3"
)

// Max result limit for queries.
const MaxResults = 10000

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

// DbDialect returns the sql dialect name.
func DbDialect() string {
	return Db().Dialect().GetName()
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

		singular := inflection.Singular(w)

		if singular != w {
			wheres = append(wheres, fmt.Sprintf("%s = '%s'", col, singular))
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
		w = strings.TrimSpace(w)

		words = append(words, slug.Make(w))

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
