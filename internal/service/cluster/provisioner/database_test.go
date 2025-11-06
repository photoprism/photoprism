package provisioner

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// --- Quoting helpers ---------------------------------------------------------

func TestQuoteIdent_Valid(t *testing.T) {
	cases := []string{
		"cluster_db",
		"my.db-1_2",
		"a",
		"z9-._",
	}
	for _, c := range cases {
		got, err := quoteIdent(c)
		assert.NoError(t, err, c)
		assert.Equal(t, "`"+c+"`", got, c)
	}
}

func TestQuoteIdent_Invalid(t *testing.T) {
	cases := []string{"", "UPPER", "bad name", "semi;colon", "ä", "`tick`"}
	for _, c := range cases {
		_, err := quoteIdent(c)
		assert.Error(t, err, c)
	}
}

func TestQuoteString_BasicEscapes(t *testing.T) {
	// Single quotes are doubled, others are left as-is.
	in := "O'Reilly & partners"
	got, err := quoteString(in)
	assert.NoError(t, err)
	assert.Equal(t, "'O''Reilly & partners'", got)

	in2 := "back`tick and normal"
	got2, err := quoteString(in2)
	assert.NoError(t, err)
	assert.Equal(t, "'back`tick and normal'", got2)
}

func TestQuoteString_RejectsNUL(t *testing.T) {
	_, err := quoteString("nul\x00here")
	assert.Error(t, err)
}

func TestQuoteAccount(t *testing.T) {
	got, err := quoteAccount("%", "cluster_user")
	assert.NoError(t, err)
	assert.Equal(t, "'cluster_user'@'%'", got)
}

func TestQuoteAccount_Errors(t *testing.T) {
	_, err := quoteAccount("%", "bad\x00user")
	assert.Error(t, err)
	_, err = quoteAccount("bad\x00host", "ok")
	assert.Error(t, err)
}

// --- Timeout helpers (using test drivers) -----------------------------------

// sleepyDriver implements ExecContext that waits for ctx.Done(), allowing us to
// verify deadline enforcement in execTimeout without a real DB.
type sleepyDriver struct{}

func (sleepyDriver) Open(name string) (driver.Conn, error) { return sleepyConn{}, nil }

type sleepyConn struct{}

func (sleepyConn) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("not implemented")
}
func (sleepyConn) Close() error              { return nil }
func (sleepyConn) Begin() (driver.Tx, error) { return nil, errors.New("not implemented") }

// Implement ExecerContext so database/sql uses Exec directly.
func (sleepyConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func TestExecTimeout_DeadlineExceeded(t *testing.T) {
	drv := "sleepy_exec_deadline"
	sql.Register(drv, sleepyDriver{})
	db, err := sql.Open(drv, "")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	start := time.Now()
	err = execTimeout(ctx, db, 50*time.Millisecond, "DO 1")
	dur := time.Since(start)
	assert.Error(t, err)
	// Should be close to the deadline, not seconds.
	assert.Less(t, dur, 500*time.Millisecond)
}

// fastDriver returns immediately and captures the last query to ensure that
// execTimeout forwards the statement correctly.
type fastDriver struct{ last *string }

func (f fastDriver) Open(name string) (driver.Conn, error) { return fastConn{last: f.last}, nil }

type fastConn struct{ last *string }

func (c fastConn) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("not implemented")
}
func (c fastConn) Close() error              { return nil }
func (c fastConn) Begin() (driver.Tx, error) { return nil, errors.New("not implemented") }
func (c fastConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if c.last != nil {
		*c.last = query
	}
	return driver.RowsAffected(1), nil
}

func TestExecTimeout_ForwardsStatement(t *testing.T) {
	var last string
	drv := "fast_exec"
	sql.Register(drv, fastDriver{last: &last})
	db, err := sql.Open(drv, "")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	stmt := "DO 1"
	err = execTimeout(ctx, db, 200*time.Millisecond, stmt)
	assert.NoError(t, err)
	assert.Equal(t, stmt, last)
}

// pingyDriver implements driver.Pinger so PingContext respects context; one
// variant waits for cancellation, the other returns quickly.
type pingyDriver struct{ wait bool }

func (p pingyDriver) Open(name string) (driver.Conn, error) { return pingyConn{wait: p.wait}, nil }

type pingyConn struct{ wait bool }

func (c pingyConn) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("not implemented")
}
func (c pingyConn) Close() error              { return nil }
func (c pingyConn) Begin() (driver.Tx, error) { return nil, errors.New("not implemented") }

func (c pingyConn) Ping(ctx context.Context) error {
	if c.wait {
		<-ctx.Done()
		return ctx.Err()
	}
	return nil
}

func TestPingWithTimeout_DeadlineExceeded(t *testing.T) {
	drv := "pingy_wait"
	sql.Register(drv, pingyDriver{wait: true})
	db, err := sql.Open(drv, "")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	start := time.Now()
	err = pingWithTimeout(ctx, db, 50*time.Millisecond)
	dur := time.Since(start)
	assert.Error(t, err)
	assert.Less(t, dur, 500*time.Millisecond)
}

func TestPingWithTimeout_Succeeds(t *testing.T) {
	drv := "pingy_fast"
	sql.Register(drv, pingyDriver{wait: false})
	db, err := sql.Open(drv, "")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	err = pingWithTimeout(ctx, db, 100*time.Millisecond)
	assert.NoError(t, err)
}
