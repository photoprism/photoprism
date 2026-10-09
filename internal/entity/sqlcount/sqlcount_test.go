package sqlcount

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// openTestDb returns a counted SQLite database with an empty table t.
func openTestDb(t *testing.T) (*sql.DB, *Counter) {
	t.Helper()

	db, c, err := Open("sqlite3", filepath.Join(t.TempDir(), "count.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("CREATE TABLE t (id INTEGER PRIMARY KEY, name TEXT)")
	require.NoError(t, err)

	return db, c
}

func TestOpen(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, c := openTestDb(t)
		require.NotNil(t, c)
		require.NoError(t, db.Ping())
	})
	t.Run("EmptyDriver", func(t *testing.T) {
		db, c, err := Open("", "")
		assert.Error(t, err)
		assert.Nil(t, db)
		assert.Nil(t, c)
	})
	t.Run("UnknownDriver", func(t *testing.T) {
		db, c, err := Open("sqlcount-unknown", "")
		assert.Error(t, err)
		assert.Nil(t, db)
		assert.Nil(t, c)
	})
}

func TestCounter(t *testing.T) {
	t.Run("ExecAndQuery", func(t *testing.T) {
		db, c := openTestDb(t)

		c.Start()
		_, err := db.Exec("INSERT INTO t (name) VALUES (?)", "a")
		require.NoError(t, err)

		var n int
		require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM t WHERE name = ?", "a").Scan(&n))
		assert.Equal(t, 1, n)

		assert.Equal(t, 2, c.Count())
		assert.Equal(t, []string{"INSERT INTO t (name) VALUES (?)", "SELECT COUNT(*) FROM t WHERE name = ?"}, c.Statements())
		assert.Len(t, c.Stop(), 2)
	})
	t.Run("NotStarted", func(t *testing.T) {
		db, c := openTestDb(t)

		_, err := db.Exec("INSERT INTO t (name) VALUES (?)", "a")
		require.NoError(t, err)
		assert.Equal(t, 0, c.Count())
	})
	t.Run("Stopped", func(t *testing.T) {
		db, c := openTestDb(t)

		c.Start()
		_, err := db.Exec("INSERT INTO t (name) VALUES (?)", "a")
		require.NoError(t, err)
		assert.Len(t, c.Stop(), 1)

		_, err = db.Exec("INSERT INTO t (name) VALUES (?)", "b")
		require.NoError(t, err)
		assert.Equal(t, 1, c.Count())
	})
	t.Run("StartResets", func(t *testing.T) {
		db, c := openTestDb(t)

		c.Start()
		_, err := db.Exec("INSERT INTO t (name) VALUES (?)", "a")
		require.NoError(t, err)

		c.Start()
		assert.Equal(t, 0, c.Count())
	})
	t.Run("Transaction", func(t *testing.T) {
		db, c := openTestDb(t)

		c.Start()
		tx, err := db.BeginTx(context.Background(), nil)
		require.NoError(t, err)
		_, err = tx.Exec("INSERT INTO t (name) VALUES (?)", "a")
		require.NoError(t, err)
		require.NoError(t, tx.Commit())

		// Transaction control is not a statement.
		assert.Equal(t, []string{"INSERT INTO t (name) VALUES (?)"}, c.Statements())
	})
	t.Run("FailedStatement", func(t *testing.T) {
		db, c := openTestDb(t)

		c.Start()
		_, err := db.Exec("INSERT INTO missing (name) VALUES (?)", "a")
		assert.Error(t, err)

		// A statement that cannot be prepared never reaches the database as an execution.
		assert.Equal(t, 0, c.Count())
	})
}

func TestNamedValues(t *testing.T) {
	t.Run("Positional", func(t *testing.T) {
		values, err := namedValues([]driver.NamedValue{{Ordinal: 1, Value: "a"}, {Ordinal: 2, Value: int64(2)}})
		require.NoError(t, err)
		assert.Equal(t, []driver.Value{"a", int64(2)}, values)
	})
	t.Run("Named", func(t *testing.T) {
		_, err := namedValues([]driver.NamedValue{{Name: "a", Ordinal: 1, Value: "a"}})
		assert.Error(t, err)
	})
}

func TestMySQL(t *testing.T) {
	dsn := os.Getenv("PHOTOPRISM_TEST_DSN")

	if os.Getenv("PHOTOPRISM_TEST_DRIVER") != "mysql" || dsn == "" {
		t.Skip("PHOTOPRISM_TEST_DRIVER is not mysql")
	}

	db, c, err := Open("mysql", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	c.Start()

	// The connection converts uint64 values with the high bit set, which the default converter refuses.
	var v uint64
	require.NoError(t, db.QueryRow("SELECT ?", uint64(1)<<63).Scan(&v))
	assert.Equal(t, uint64(1)<<63, v)
	assert.Equal(t, 1, c.Count())
}

func TestOpenGorm(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		p, err := OpenGorm("sqlite3", filepath.Join(t.TempDir(), "gorm.db"))
		require.NoError(t, err)
		t.Cleanup(func() { _ = p.Close() })

		require.NoError(t, p.Db().Exec("CREATE TABLE t (id INTEGER PRIMARY KEY, name TEXT)").Error)

		p.Counter.Start()
		require.NoError(t, p.Db().Exec("INSERT INTO t (name) VALUES (?)", "a").Error)

		var n int64
		require.NoError(t, p.Db().Table("t").Count(&n).Error)
		assert.Equal(t, int64(1), n)
		assert.Equal(t, 2, p.Counter.Count())
	})
	t.Run("UnknownDriver", func(t *testing.T) {
		p, err := OpenGorm("sqlcount-unknown", "")
		assert.Error(t, err)
		assert.Nil(t, p)
	})
}

func TestCountingStmt(t *testing.T) {
	newStmt := func() (*countingStmt, *fakeStmt, *Counter) {
		inner := &fakeStmt{}
		c := &Counter{}
		c.Start()
		return &countingStmt{Stmt: inner, conn: &fakeConn{}, query: "q", counter: c}, inner, c
	}

	t.Run("Exec", func(t *testing.T) {
		s, inner, c := newStmt()
		_, err := s.Exec([]driver.Value{"a"}) //nolint:staticcheck // Tests the deprecated interface method.
		require.NoError(t, err)
		assert.Equal(t, []driver.Value{"a"}, inner.args)
		assert.Equal(t, []string{"q"}, c.Statements())
	})
	t.Run("Query", func(t *testing.T) {
		s, inner, c := newStmt()
		_, err := s.Query([]driver.Value{"a"}) //nolint:staticcheck // Tests the deprecated interface method.
		require.NoError(t, err)
		assert.Equal(t, []driver.Value{"a"}, inner.args)
		assert.Equal(t, 1, c.Count())
	})
	t.Run("ExecContextFallback", func(t *testing.T) {
		s, inner, c := newStmt()
		_, err := s.ExecContext(context.Background(), []driver.NamedValue{{Ordinal: 1, Value: "a"}})
		require.NoError(t, err)
		assert.Equal(t, []driver.Value{"a"}, inner.args)
		assert.Equal(t, 1, c.Count())
	})
	t.Run("QueryContextFallback", func(t *testing.T) {
		s, inner, c := newStmt()
		_, err := s.QueryContext(context.Background(), []driver.NamedValue{{Ordinal: 1, Value: "a"}})
		require.NoError(t, err)
		assert.Equal(t, []driver.Value{"a"}, inner.args)
		assert.Equal(t, 1, c.Count())
	})
	t.Run("NamedArgument", func(t *testing.T) {
		s, _, _ := newStmt()
		_, err := s.ExecContext(context.Background(), []driver.NamedValue{{Name: "a", Ordinal: 1, Value: "a"}})
		assert.Error(t, err)
	})
	t.Run("CheckNamedValueSkip", func(t *testing.T) {
		s, _, _ := newStmt()
		assert.ErrorIs(t, s.CheckNamedValue(&driver.NamedValue{Value: "a"}), driver.ErrSkip)
	})
}

func TestCountingConn(t *testing.T) {
	t.Run("Forwarded", func(t *testing.T) {
		inner := &fakeConnFull{fakeConn{valid: false}}
		c := &countingConn{Conn: inner, counter: &Counter{}}

		require.NoError(t, c.ResetSession(context.Background()))
		require.NoError(t, c.Ping(context.Background()))
		assert.Equal(t, 1, inner.resets)
		assert.Equal(t, 1, inner.pings)
		assert.False(t, c.IsValid())
	})
	t.Run("Defaults", func(t *testing.T) {
		c := &countingConn{Conn: &fakeConn{}, counter: &Counter{}}

		assert.NoError(t, c.ResetSession(context.Background()))
		assert.NoError(t, c.Ping(context.Background()))
		assert.True(t, c.IsValid())
		assert.ErrorIs(t, c.CheckNamedValue(&driver.NamedValue{Value: "a"}), driver.ErrSkip)
	})
	t.Run("BeginFallback", func(t *testing.T) {
		c := &countingConn{Conn: &fakeConn{}, counter: &Counter{}}
		_, err := c.BeginTx(context.Background(), driver.TxOptions{})
		assert.Error(t, err)
	})
}

func TestConnector(t *testing.T) {
	inner := &fakeConnFull{fakeConn{valid: true}}
	drv := &fakeDriver{conn: &inner.fakeConn}
	c := &connector{driver: drv, dsn: "fake", counter: &Counter{}}

	assert.Same(t, drv, c.Driver())

	conn, err := c.Connect(context.Background())
	require.NoError(t, err)
	assert.IsType(t, &countingConn{}, conn)
}
