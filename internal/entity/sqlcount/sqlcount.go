/*
Package sqlcount counts the SQL statements a database/sql connection pool executes, so tests can pin
how many statements an operation issues independently of the ORM version. It is meant to be imported
from tests only.

Copyright (c) 2018 - 2026 PhotoPrism UG. All rights reserved.

	This program is free software: you can redistribute it and/or modify
	it under Version 3 of the GNU Affero General Public License (the "AGPL"):
	<https://docs.photoprism.app/license/agpl>

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	The AGPL is supplemented by our Trademark and Brand Guidelines,
	which describe how our Brand Assets may be used:
	<https://www.photoprism.app/trademark/>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>
*/
package sqlcount

import (
	"database/sql"
	"errors"
	"sync"
)

// Counter records the statements executed through the connections of one pool.
type Counter struct {
	mu         sync.Mutex
	enabled    bool
	statements []string
}

// Open returns a connection pool for the registered driver and data source name whose statements
// are recorded by the returned Counter while it is started.
func Open(driverName, dataSourceName string) (*sql.DB, *Counter, error) {
	if driverName == "" {
		return nil, nil, errors.New("sqlcount: driver name is empty")
	}

	// sql.Open validates the driver name without connecting to the database.
	db, err := sql.Open(driverName, dataSourceName)

	if err != nil {
		return nil, nil, err
	}

	drv := db.Driver()

	if err = db.Close(); err != nil {
		return nil, nil, err
	}

	c := &Counter{}

	return sql.OpenDB(&connector{driver: drv, dsn: dataSourceName, counter: c}), c, nil
}

// Start discards the statements recorded so far and starts recording.
func (c *Counter) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.enabled = true
	c.statements = nil
}

// Stop stops recording and returns the statements recorded since Start.
func (c *Counter) Stop() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.enabled = false

	return c.statements
}

// Statements returns a copy of the statements recorded since Start.
func (c *Counter) Statements() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return append([]string(nil), c.statements...)
}

// Count returns the number of statements recorded since Start.
func (c *Counter) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.statements)
}

// record adds a statement if recording is enabled.
func (c *Counter) record(query string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.enabled {
		c.statements = append(c.statements, query)
	}
}
