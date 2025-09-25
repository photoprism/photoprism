package provisioner

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/config"
)

// Credentials contains the connection details returned when ensuring a node database.
type Credentials struct {
	Driver    string
	Host      string
	Port      int
	Name      string
	User      string
	Password  string
	DSN       string
	RotatedAt string
}

// GetCredentials ensures a per-node database and user exist with minimal grants.
// - Requires a MySQL/MariaDB admin connection (this package maintains it).
// - Returns created=true if the database schema did not exist before.
// - If rotate is true or created, rotates the user password and includes it (and DSN) in the result.
func GetCredentials(ctx context.Context, conf *config.Config, nodeUUID, nodeName string, rotate bool) (Credentials, bool, error) {
	out := Credentials{}

	// Normalize provisioner driver to lower-case to accept variants like "MySQL"/"MariaDB".
	DatabaseDriver = strings.ToLower(DatabaseDriver)

	switch DatabaseDriver {
	case config.MySQL, config.MariaDB:
		// ok
	case config.SQLite3, config.Postgres:
		return out, false, errors.New("database must be MySQL/MariaDB for auto-provisioning")
	default:
		// Driver is configured externally for the provisioner (decoupled from app config).
		return out, false, fmt.Errorf("unsupported auto-provisioning database driver: %s", DatabaseDriver)
	}

	// Compute deterministic names and a candidate password.
	dbName, dbUser, dbPass := GenerateCreds(conf, nodeUUID, nodeName)

	// Extra safety: enforce allowed identifier charset.
	if !identRe.MatchString(dbName) || !identRe.MatchString(dbUser) {
		return out, false, errors.New("invalid generated database identifiers")
	}

	// Get (or open) admin DB handle.
	db, err := GetDB(ctx)
	if err != nil {
		return out, false, err
	}

	// 1) Determine if database already exists (values can be parameterized).
	var count int
	{
		c, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := db.QueryRowContext(
			c,
			"SELECT COUNT(*) FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?",
			dbName,
		).Scan(&count); err != nil {
			return out, false, err
		}
	}
	created := count == 0

	// 2) Create database schema if needed (identifier must be quoted).
	qDB, err := quoteIdent(dbName)
	if err != nil {
		return out, created, err
	}
	createDB := "CREATE DATABASE IF NOT EXISTS " + qDB + " CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"
	if err := execTimeout(ctx, db, 15*time.Second, createDB); err != nil {
		return out, created, err
	}

	// 3) Ensure user exists.
	acc, err := quoteAccount("%", dbUser) // user@'%'
	if err != nil {
		return out, created, err
	}
	pass, err := quoteString(dbPass)
	if err != nil {
		return out, created, err
	}

	createUser := "CREATE USER IF NOT EXISTS " + acc + " IDENTIFIED BY " + pass
	if err := execTimeout(ctx, db, 10*time.Second, createUser); err != nil {
		return out, created, err
	}

	// 4) Rotate or set password explicitly on first creation.
	if rotate || created {
		alterUser := "ALTER USER " + acc + " IDENTIFIED BY " + pass
		if err := execTimeout(ctx, db, 10*time.Second, alterUser); err != nil {
			return out, created, err
		}
		out.Password = dbPass
		out.RotatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	// 5) Grant privileges on schema.
	grant := "GRANT ALL PRIVILEGES ON " + qDB + ".* TO " + acc
	if err := execTimeout(ctx, db, 10*time.Second, grant); err != nil {
		return out, created, err
	}

	// 6) Optional on modern MariaDB/MySQL; harmless if included.
	if err := execTimeout(ctx, db, 5*time.Second, "FLUSH PRIVILEGES"); err != nil {
		return out, created, err
	}

	// Compose credentials.
	out.Host = DatabaseHost
	out.Port = DatabasePort
	out.Name = dbName
	out.User = dbUser
	out.Driver = DatabaseDriver

	if out.Password != "" {
		out.DSN = BuildDSN(DatabaseDriver, out.Host, out.Port, out.User, out.Password, out.Name)
	}

	return out, created, nil
}
