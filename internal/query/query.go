/*
Package query provides frequently used database queries for use in commands and API.

Copyright (c) 2018 - 2023 PhotoPrism UG. All rights reserved.

	This program is free software: you can redistribute it and/or modify
	it under Version 3 of the GNU Affero General Public License (the "AGPL"):
	<https://docs.photoprism.app/license/agpl>

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	The AGPL is supplemented by our Trademark and Brand Guidelines,
	which describe how our Brand Assets may be used:
	<https://www.photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>
*/
package query

import (
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
)

var log = event.Log

const (
	MySQL   = "mysql"
	SQLite3 = "sqlite3"
)

// Cols represents a list of database columns.
type Cols []string

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

// DbDialect returns the sql database dialect name.
func DbDialect() string {
	return Db().Dialect().GetName()
}

// BatchSize returns the maximum query parameter number based on the current sql database dialect.
func BatchSize() int {
	switch DbDialect() {
	case SQLite3:
		return 333
	default:
		return 1000
	}
}

// logErr logs an error and keeps quiet otherwise.
func logErr(prefix, action string, err error) {
	if err != nil {
		log.Errorf("%s: %s (%s)", prefix, err, action)
	}
}
