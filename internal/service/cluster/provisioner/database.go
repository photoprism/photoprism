package provisioner

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql" // register MySQL driver
)

// devDatabasePort returns the dev MariaDB port, honoring MARIADB_PORT when set so
// the provisioner stays aligned with compose.yaml overrides.
func devDatabasePort() int {
	if v := os.Getenv("MARIADB_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n < 65536 {
			return n
		}
	}
	return 4001
}

// ProvisionDSN specifies the admin DSN used for auto-provisioning, for example:
// root:insecure@tcp(127.0.0.1:3306)/photoprism?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true
var ProvisionDSN = fmt.Sprintf("root:photoprism@tcp(mariadb:%d)/photoprism?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true", devDatabasePort())

// DatabaseHost is the hostname of the admin server used for provisioning operations.
var DatabaseHost = "mariadb"

// DatabasePort is the port of the admin server used for provisioning operations.
var DatabasePort = devDatabasePort()

// DatabaseDriver indicates the SQL driver used for provisioning (independent from the app DB driver).
var DatabaseDriver = "mysql"

// ErrNotConfigured is returned by GetDB if no provisioning DSN is set.
var ErrNotConfigured = errors.New("auto-provisioning database is not configured")

// -----------------------------------------------------------------------------
// Persistent auto-provisioning *sql.DB connection with liveness checks
// -----------------------------------------------------------------------------

var (
	dbConn  *sql.DB
	dbDSN   string
	dbMutex sync.RWMutex
)

// GetDB returns a pooled auto-provisioning connection, opening (or reopening) if needed.
// It pings with a short timeout before returning, and reopens the pool when ProvisionDSN has changed.
func GetDB(ctx context.Context) (*sql.DB, error) {
	// Fast path with read lock.
	dbMutex.RLock()
	db, connDSN := dbConn, dbDSN
	dbMutex.RUnlock()

	if db != nil {
		if connDSN == ProvisionDSN {
			if err := pingWithTimeout(ctx, db, 3*time.Second); err == nil {
				return db, nil
			}
		}
		// Ping failed or DSN changed -> close & rebuild.
		_ = db.Close()
		clearDB(db)
	}

	connDSN = ProvisionDSN

	if connDSN == "" {
		return nil, ErrNotConfigured
	}

	var err error

	db, err = sql.Open("mysql", connDSN)
	if err != nil {
		return nil, err
	}

	// No connection is kept idle, so each statement uses the account's current privileges, which the
	// server applies to new connections only, and a connection is closed once it has been used.
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(0)

	// Verify connection.
	if pingErr := pingWithTimeout(ctx, db, 5*time.Second); pingErr != nil {
		_ = db.Close()
		return nil, pingErr
	}

	return storeDB(db, connDSN), nil
}

// storeDB stores a newly opened pool unless another caller stored one for the same DSN first,
// in which case it closes db and returns the stored pool.
func storeDB(db *sql.DB, connDSN string) *sql.DB {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if dbConn != nil && dbDSN == connDSN {
		_ = db.Close()
		return dbConn
	} else if dbConn != nil {
		_ = dbConn.Close()
	}

	dbConn, dbDSN = db, connDSN
	return db
}

// clearDB removes the cached pool if it is still db.
func clearDB(db *sql.DB) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if dbConn == db {
		dbConn, dbDSN = nil, ""
	}
}

// pingWithTimeout validates liveness by issuing a ping bounded by d.
func pingWithTimeout(ctx context.Context, db *sql.DB, d time.Duration) error {
	c, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	return db.PingContext(c)
}

// -----------------------------------------------------------------------------
// Quoting & validation helpers
// -----------------------------------------------------------------------------

// Allow only safe characters in generated identifiers (you can tighten/loosen this).
var identRe = regexp.MustCompile(`^[a-z0-9\-_.]+$`)

// quoteIdent wraps an identifier in backticks after validating its characters.
func quoteIdent(s string) (string, error) {
	if s == "" {
		return "", errors.New("empty identifier")
	}
	if !identRe.MatchString(s) {
		return "", fmt.Errorf("invalid identifier %q", s)
	}
	// Backtick-escape any accidental backticks (shouldn't happen with identRe).
	return "`" + strings.ReplaceAll(s, "`", "``") + "`", nil
}

// quoteString escapes and quotes a string literal for SQL statements. Backslashes and control characters
// are refused, since only single quotes are escaped.
func quoteString(s string) (string, error) {
	if strings.ContainsRune(s, '\x00') {
		return "", errors.New("string contains NUL")
	} else if strings.ContainsFunc(s, func(r rune) bool { return r == '\\' || r < 0x20 || r == 0x7f }) {
		return "", errors.New("string contains unsupported characters")
	}
	// SQL-92 string literal quoting: single quotes doubled.
	return "'" + strings.ReplaceAll(s, "'", "''") + "'", nil
}

// quoteAccount formats a user@host identifier using SQL quoting rules.
func quoteAccount(host, user string) (string, error) {
	u, err := quoteString(user)
	if err != nil {
		return "", fmt.Errorf("invalid user: %w", err)
	}
	h, err := quoteString(host)
	if err != nil {
		return "", fmt.Errorf("invalid host: %w", err)
	}
	return u + "@" + h, nil
}

// execTimeout executes stmt with a deadline by wrapping the call in a cancelable context.
func execTimeout(ctx context.Context, db *sql.DB, d time.Duration, stmt string) error {
	c, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	_, err := db.ExecContext(c, stmt)
	return err
}
