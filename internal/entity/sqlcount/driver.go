package sqlcount

import (
	"context"
	"database/sql/driver"
	"errors"
)

// connector opens connections of the wrapped driver that report their statements to a Counter.
type connector struct {
	driver  driver.Driver
	dsn     string
	counter *Counter
}

// Connect opens a connection of the wrapped driver.
func (c *connector) Connect(ctx context.Context) (driver.Conn, error) {
	var conn driver.Conn
	var err error

	if dc, ok := c.driver.(driver.DriverContext); ok {
		var inner driver.Connector

		if inner, err = dc.OpenConnector(c.dsn); err != nil {
			return nil, err
		}

		conn, err = inner.Connect(ctx)
	} else {
		conn, err = c.driver.Open(c.dsn)
	}

	if err != nil {
		return nil, err
	}

	return &countingConn{Conn: conn, counter: c.counter}, nil
}

// Driver returns the wrapped driver.
func (c *connector) Driver() driver.Driver {
	return c.driver
}

// countingConn wraps a driver connection. It implements neither driver.ExecerContext nor
// driver.QueryerContext, so database/sql prepares every statement and runs it through countingStmt.
// Counts are per prepared execution: SQLite then runs only the first of several statements in one string.
type countingConn struct {
	driver.Conn
	counter *Counter
}

// Prepare returns a prepared statement that reports each execution.
func (c *countingConn) Prepare(query string) (driver.Stmt, error) {
	return c.PrepareContext(context.Background(), query)
}

// PrepareContext returns a prepared statement that reports each execution.
func (c *countingConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	var stmt driver.Stmt
	var err error

	if pc, ok := c.Conn.(driver.ConnPrepareContext); ok {
		stmt, err = pc.PrepareContext(ctx, query)
	} else {
		stmt, err = c.Conn.Prepare(query)
	}

	// A statement the database rejects when preparing it is counted here, as it never executes.
	if err != nil {
		c.counter.recordResult(query, err)
		return nil, err
	}

	return &countingStmt{Stmt: stmt, conn: c.Conn, query: query, counter: c.counter}, nil
}

// BeginTx starts a transaction on the wrapped connection.
func (c *countingConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if bc, ok := c.Conn.(driver.ConnBeginTx); ok {
		return bc.BeginTx(ctx, opts)
	}

	return c.Conn.Begin() //nolint:staticcheck // Fallback for drivers without ConnBeginTx.
}

// CheckNamedValue lets the wrapped connection convert arguments, as it would without the wrapper.
func (c *countingConn) CheckNamedValue(v *driver.NamedValue) error {
	if nc, ok := c.Conn.(driver.NamedValueChecker); ok {
		return nc.CheckNamedValue(v)
	}

	return driver.ErrSkip
}

// Ping forwards the connection check, so db.Ping reaches the database.
func (c *countingConn) Ping(ctx context.Context) error {
	if p, ok := c.Conn.(driver.Pinger); ok {
		return p.Ping(ctx)
	}

	return nil
}

// ResetSession forwards the session reset, so the pool discards connections the driver reports as bad.
func (c *countingConn) ResetSession(ctx context.Context) error {
	if sr, ok := c.Conn.(driver.SessionResetter); ok {
		return sr.ResetSession(ctx)
	}

	return nil
}

// IsValid forwards the validity check, so the pool discards connections the driver reports as invalid.
func (c *countingConn) IsValid() bool {
	if v, ok := c.Conn.(driver.Validator); ok {
		return v.IsValid()
	}

	return true
}

// countingStmt wraps a prepared statement and reports each execution to a Counter.
type countingStmt struct {
	driver.Stmt
	conn    driver.Conn
	query   string
	counter *Counter
}

// Exec executes the statement and records it.
func (s *countingStmt) Exec(args []driver.Value) (driver.Result, error) {
	res, err := s.Stmt.Exec(args) //nolint:staticcheck // Required by the driver.Stmt interface.
	s.counter.recordResult(s.query, err)
	return res, err
}

// Query executes the statement and records it.
func (s *countingStmt) Query(args []driver.Value) (driver.Rows, error) {
	rows, err := s.Stmt.Query(args) //nolint:staticcheck // Required by the driver.Stmt interface.
	s.counter.recordResult(s.query, err)
	return rows, err
}

// ExecContext executes the statement and records it.
func (s *countingStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (res driver.Result, err error) {
	defer func() { s.counter.recordResult(s.query, err) }()

	if ec, ok := s.Stmt.(driver.StmtExecContext); ok {
		return ec.ExecContext(ctx, args)
	}

	values, err := namedValues(args)

	if err != nil {
		return nil, err
	}

	return s.Stmt.Exec(values) //nolint:staticcheck // Fallback for drivers without StmtExecContext.
}

// QueryContext executes the statement and records it.
func (s *countingStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (rows driver.Rows, err error) {
	defer func() { s.counter.recordResult(s.query, err) }()

	if qc, ok := s.Stmt.(driver.StmtQueryContext); ok {
		return qc.QueryContext(ctx, args)
	}

	values, err := namedValues(args)

	if err != nil {
		return nil, err
	}

	return s.Stmt.Query(values) //nolint:staticcheck // Fallback for drivers without StmtQueryContext.
}

// CheckNamedValue converts arguments like the wrapped statement or, failing that, its connection.
// database/sql consults the statement first and never reaches the connection once it implements this.
func (s *countingStmt) CheckNamedValue(v *driver.NamedValue) error {
	if nc, ok := s.Stmt.(driver.NamedValueChecker); ok {
		return nc.CheckNamedValue(v)
	} else if nc, ok = s.conn.(driver.NamedValueChecker); ok {
		return nc.CheckNamedValue(v)
	}

	return driver.ErrSkip
}

// namedValues converts positional named values for drivers that only accept plain values.
func namedValues(args []driver.NamedValue) ([]driver.Value, error) {
	values := make([]driver.Value, len(args))

	for i, arg := range args {
		if arg.Name != "" {
			return nil, errors.New("sqlcount: named arguments are not supported")
		}

		values[i] = arg.Value
	}

	return values, nil
}
