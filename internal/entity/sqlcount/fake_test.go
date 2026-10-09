package sqlcount

import (
	"context"
	"database/sql/driver"
	"errors"
)

// fakeDriver is a minimal driver whose connections and statements implement only the required interfaces.
type fakeDriver struct{ conn *fakeConn }

// Open returns the fake connection.
func (d *fakeDriver) Open(string) (driver.Conn, error) { return d.conn, nil }

// fakeConn records calls to the optional interfaces it may implement.
type fakeConn struct {
	resets, pings int
	valid         bool
}

// Prepare returns a fake statement.
func (c *fakeConn) Prepare(query string) (driver.Stmt, error) { return &fakeStmt{}, nil }

// Close does nothing.
func (c *fakeConn) Close() error { return nil }

// Begin returns an error, as the tests do not use transactions.
func (c *fakeConn) Begin() (driver.Tx, error) { return nil, errors.New("not supported") }

// fakeConnFull adds session reset, validation and ping to fakeConn.
type fakeConnFull struct{ fakeConn }

// ResetSession records the reset.
func (c *fakeConnFull) ResetSession(context.Context) error { c.resets++; return nil }

// IsValid returns the configured validity.
func (c *fakeConnFull) IsValid() bool { return c.valid }

// Ping records the ping.
func (c *fakeConnFull) Ping(context.Context) error { c.pings++; return nil }

// fakeStmt implements only the plain driver.Stmt methods.
type fakeStmt struct{ args []driver.Value }

// Close does nothing.
func (s *fakeStmt) Close() error { return nil }

// NumInput accepts any number of arguments.
func (s *fakeStmt) NumInput() int { return -1 }

// Exec records the arguments.
func (s *fakeStmt) Exec(args []driver.Value) (driver.Result, error) {
	s.args = args
	return driver.RowsAffected(1), nil
}

// Query records the arguments and returns no rows.
func (s *fakeStmt) Query(args []driver.Value) (driver.Rows, error) {
	s.args = args
	return &fakeRows{}, nil
}

// fakeRows is an empty result set.
type fakeRows struct{}

// Columns returns no columns.
func (r *fakeRows) Columns() []string { return nil }

// Close does nothing.
func (r *fakeRows) Close() error { return nil }

// Next reports the end of the result set.
func (r *fakeRows) Next([]driver.Value) error { return errors.New("EOF") }
