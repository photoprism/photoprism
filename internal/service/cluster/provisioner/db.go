package provisioner

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/config"
)

// Creds contains the connection details returned when ensuring a node database.
type Creds struct {
	Host          string
	Port          int
	Name          string
	User          string
	Password      string
	DSN           string
	LastRotatedAt string
}

var identRe = regexp.MustCompile(`^[a-z0-9\-_.]+$`)

func quoteIdent(s string) string { return "`" + strings.ReplaceAll(s, "`", "``") + "`" }

// EnsureNodeDatabase ensures a per-node database and user exist with minimal grants.
// - Requires MySQL/MariaDB driver on the portal.
// - Returns created=true if the database schema did not exist before.
// - If rotate is true or created, rotates the user password and includes it (and DSN) in the result.
func EnsureNodeDatabase(ctx context.Context, conf *config.Config, nodeName string, rotate bool) (Creds, bool, error) {
	out := Creds{}

	switch conf.DatabaseDriver() {
	case config.MySQL, config.MariaDB:
		// ok
	case config.SQLite3, config.Postgres:
		return out, false, errors.New("portal database must be MySQL/MariaDB for registration")
	default:
		return out, false, fmt.Errorf("unsupported portal database driver: %s", conf.DatabaseDriver())
	}

	// Compute deterministic names and a candidate password.
	dbName, dbUser, dbPass := GenerateCreds(conf, nodeName)

	// Extra safety: enforce allowed identifier charset.
	if !identRe.MatchString(dbName) || !identRe.MatchString(dbUser) {
		return out, false, errors.New("invalid generated database identifiers")
	}

	// Determine if database already exists.
	type res struct{ C int }
	var r res

	q := conf.Db().Unscoped()

	if err := q.Raw("SELECT COUNT(*) AS C FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?", dbName).Scan(&r).Error; err != nil {
		return out, false, err
	}

	created := r.C == 0

	// Create database schema if needed.
	if err := exec(q, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", quoteIdent(dbName))); err != nil {
		return out, created, err
	}

	// Create user if needed (host wildcard '%').
	if err := exec(q, fmt.Sprintf("CREATE USER IF NOT EXISTS '%s'@'%%' IDENTIFIED BY '%s'", dbUser, dbPass)); err != nil {
		return out, created, err
	}

	// Rotate or set password explicitly on first creation.
	if rotate || created {
		if err := exec(q, fmt.Sprintf("ALTER USER '%s'@'%%' IDENTIFIED BY '%s'", dbUser, dbPass)); err != nil {
			return out, created, err
		}
		out.Password = dbPass
		out.LastRotatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	// Grant privileges on schema.
	if err := exec(q, fmt.Sprintf("GRANT ALL PRIVILEGES ON %s.* TO '%s'@'%%'", quoteIdent(dbName), dbUser)); err != nil {
		return out, created, err
	}

	// Optional on modern MariaDB, harmless if included.
	if err := exec(q, "FLUSH PRIVILEGES"); err != nil {
		return out, created, err
	}

	out.Host = conf.DatabaseHost()
	out.Port = conf.DatabasePort()
	out.Name = dbName
	out.User = dbUser

	if out.Password != "" {
		out.DSN = BuildDSN(out.Host, out.Port, out.User, out.Password, out.Name)
	}

	return out, created, nil
}

func exec(db *gorm.DB, stmt string) error {
	if stmt == "" {
		return nil
	}

	// Use a no-op scan into a struct to execute raw SQL with gorm v1.
	var nop struct{}
	return db.Raw(stmt).Scan(&nop).Error
}
