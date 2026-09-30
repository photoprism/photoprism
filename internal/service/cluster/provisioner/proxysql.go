package provisioner

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
)

const (
	// DefaultProxyHostgroup routes instance connections to the primary (writer) backend hostgroup.
	DefaultProxyHostgroup = 10
	// DefaultProxyFrontend enables clients to authenticate through ProxySQL; required for instance users.
	DefaultProxyFrontend = 1
	// DefaultProxyBackend keeps instance users from authenticating against upstream servers directly.
	DefaultProxyBackend = 0
	// DefaultProxyMaxConnections caps concurrent connections per instance to avoid exhausting ProxySQL.
	DefaultProxyMaxConnections = 200
	// DefaultProxyUseSSL toggles ProxySQL's SSL flag for instance accounts (0 = disabled by default).
	DefaultProxyUseSSL = 0
	// DefaultProxyComment labels provisioned users so operators can distinguish auto-managed accounts.
	DefaultProxyComment = "Portal provisioned instance"
)

// ProxyOptions describes the ProxySQL mysql_users attributes to apply when syncing instance accounts.
type ProxyOptions struct {
	Hostgroup      int
	Frontend       int
	Backend        int
	MaxConnections int
	UseSSL         int
	Comment        string
}

// ProvisionProxyDSN specifies the optional ProxySQL admin DSN (port 6032 by default) for keeping user accounts in sync.
var ProvisionProxyDSN = ""

// ProvisionProxyOptions stores the current defaults used when synchronizing ProxySQL instance accounts.
var ProvisionProxyOptions = ProxyOptions{
	Hostgroup:      DefaultProxyHostgroup,
	Frontend:       DefaultProxyFrontend,
	Backend:        DefaultProxyBackend,
	MaxConnections: DefaultProxyMaxConnections,
	UseSSL:         DefaultProxyUseSSL,
	Comment:        DefaultProxyComment,
}

// SyncProxyUser ensures the ProxySQL mysql_users entry matches the provided schema and credentials.
// When pass is empty the existing password is preserved, allowing non-rotating syncs that only adjust metadata.
func SyncProxyUser(ctx context.Context, proxyDSN, schema, user, pass string, opts ProxyOptions) (err error) {
	adminDsn, err := normalizeProxyDSN(proxyDSN)
	if err != nil {
		return err
	}

	db, err := sql.Open("mysql", adminDsn)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, db.Close())
	}()

	password := pass
	if password == "" {
		if err := db.QueryRowContext(ctx, "SELECT password FROM mysql_users WHERE username = ?", user).Scan(&password); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("proxysql: existing user not found and password not provided")
			}
			return err
		}
	}

	if opts.Comment == "" {
		opts.Comment = DefaultProxyComment
	}

	if _, err = db.ExecContext(ctx, "DELETE FROM mysql_users WHERE username = ?", user); err != nil {
		return err
	}

	if _, err = db.ExecContext(ctx, `
		INSERT INTO mysql_users (
			username, password, active, use_ssl, default_hostgroup,
			default_schema, schema_locked, transaction_persistent,
			fast_forward, backend, frontend, max_connections, attributes, comment
		) VALUES (
			?, ?, 1, ?, ?, ?,
			0, 1,
			0, ?, ?, ?, '{}', ?
		)
	`, user, password, opts.UseSSL, opts.Hostgroup, schema, opts.Backend, opts.Frontend, opts.MaxConnections, opts.Comment); err != nil {
		return err
	}

	return applyProxySQL(ctx, db)
}

// DropProxyUser removes the mysql_users record for a instance and reloads ProxySQL runtime/disk.
func DropProxyUser(ctx context.Context, proxyDSN, user string) (err error) {
	adminDsn, err := normalizeProxyDSN(proxyDSN)
	if err != nil {
		return err
	}

	db, err := sql.Open("mysql", adminDsn)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, db.Close())
	}()

	if _, err = db.ExecContext(ctx, "DELETE FROM mysql_users WHERE username = ?", user); err != nil {
		return err
	}

	return applyProxySQL(ctx, db)
}

// applyProxySQL reloads mysql_users into ProxySQL runtime and persists the changes to disk.
func applyProxySQL(ctx context.Context, db *sql.DB) error {
	for _, stmt := range []string{
		"LOAD MYSQL USERS TO RUNTIME",
		"SAVE MYSQL USERS TO DISK",
	} {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

// proxyParamRules lists the DSN parameters accepted for the ProxySQL admin interface and checks their
// values. The character set is limited to UTF-8.
var proxyParamRules = dsn.ParamRules{
	"charset":           dsn.ValidCharset,
	"collation":         dsn.ValidCollation,
	"parseTime":         dsn.ValidBool,
	"interpolateParams": dsn.ValidBool,
	"timeout":           dsn.ValidDuration,
	"readTimeout":       dsn.ValidDuration,
	"writeTimeout":      dsn.ValidDuration,
	"maxAllowedPacket":  dsn.ValidPacketSize,
	"tls":               dsn.ValidTLS,
}

// normalizeProxyDSN returns a ProxySQL admin DSN with the accepted parameters only, adding
// interpolateParams=true so prepared statements work, and charset=utf8mb4 if no charset or collation is set.
func normalizeProxyDSN(proxyDsn string) (string, error) {
	if proxyDsn == "" {
		return "", nil
	} else if !validProxyDSN(proxyDsn) {
		return "", errors.New("proxysql: invalid admin dsn")
	}

	query := dsn.Query(proxyDsn)
	params, dropped, err := dsn.FilterParams(query, proxyParamRules)

	if err != nil {
		return "", fmt.Errorf("proxysql: %w", err)
	} else if names := dsn.LoggableParamNames(dropped); len(names) > 0 {
		log.Warnf("proxysql: ignored %d unsupported admin dsn parameters %s", len(dropped), clean.LogNames(names))
	} else if len(dropped) > 0 {
		log.Warnf("proxysql: ignored %d unsupported admin dsn parameters", len(dropped))
	}

	defaults := "charset=utf8mb4&interpolateParams=true"

	if dsn.HasParam(params, "charset") || dsn.HasParam(params, "collation") {
		defaults = "interpolateParams=true"
	}

	return strings.TrimSuffix(proxyDsn[:len(proxyDsn)-len(query)], "?") + "?" + dsn.MergeParams(params, defaults), nil
}

// validProxyDSN reports whether the MySQL driver can parse a DSN, including one with a parameter it
// refuses by panicking.
func validProxyDSN(s string) (valid bool) {
	defer func() {
		if recover() != nil {
			valid = false
		}
	}()

	_, err := mysql.ParseDSN(s)

	return err == nil
}
