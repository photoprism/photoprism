package provisioner

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// ProvisionDSN specifies the admin DSN used for auto-provisioning, for example:
// root:insecure@tcp(127.0.0.1:3306)/photoprism?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true
var ProvisionDSN = "root:photoprism@tcp(mariadb:4001)/photoprism?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true"

// DatabaseHost is the hostname of the admin server used for provisioning operations.
var DatabaseHost = "mariadb"

// DatabasePort is the port of the admin server used for provisioning operations.
var DatabasePort = 4001

// DatabaseDriver indicates the SQL driver used for provisioning (independent from the app DB driver).
var DatabaseDriver = "mysql"

// -----------------------------------------------------------------------------
// Persistent auto-provisioning *sql.DB connection with liveness checks
// -----------------------------------------------------------------------------

var (
	dbConn  *sql.DB
	dbMutex sync.RWMutex
)

// GetDB returns a pooled auto-provisioning connection, opening (or reopening) if needed.
// It pings with a short timeout before returning to ensure liveness.
func GetDB(ctx context.Context) (*sql.DB, error) {
	// Fast path with read lock.
	dbMutex.RLock()
	db := dbConn
	dbMutex.RUnlock()

	if db != nil {
		if err := pingWithTimeout(ctx, db, 3*time.Second); err == nil {
			return db, nil
		}
		// Ping failed -> close & rebuild.
		_ = db.Close()
		setDB(nil)
	}

	var err error

	db, err = sql.Open("mysql", ProvisionDSN)
	if err != nil {
		return nil, err
	}

	// Reasonable pool settings; adjust for your environment.
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	// Verify connection.
	if pingErr := pingWithTimeout(ctx, db, 5*time.Second); pingErr != nil {
		_ = db.Close()
		return nil, pingErr
	}

	setDB(db)
	return db, nil
}

// setDB stores the shared provisioning connection under write lock.
func setDB(db *sql.DB) {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	dbConn = db
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

// quoteString escapes and quotes a string literal for SQL statements.
func quoteString(s string) (string, error) {
	if strings.ContainsRune(s, '\x00') {
		return "", errors.New("string contains NUL")
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
