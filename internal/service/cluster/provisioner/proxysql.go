package provisioner

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

const (
	// DefaultProxyHostgroup routes tenant connections to the primary (writer) backend hostgroup.
	DefaultProxyHostgroup = 10
	// DefaultProxyFrontend enables clients to authenticate through ProxySQL; required for tenant users.
	DefaultProxyFrontend = 1
	// DefaultProxyBackend keeps tenant users from authenticating against upstream servers directly.
	DefaultProxyBackend = 0
	// DefaultProxyMaxConnections caps concurrent connections per tenant to avoid exhausting ProxySQL.
	DefaultProxyMaxConnections = 200
	// DefaultProxyUseSSL toggles ProxySQL's SSL flag for tenant accounts (0 = disabled by default).
	DefaultProxyUseSSL = 0
	// DefaultProxyComment labels provisioned users so operators can distinguish auto-managed accounts.
	DefaultProxyComment = "Portal provisioned tenant"
)

// ProxyOptions describes the ProxySQL mysql_users attributes to apply when syncing tenant accounts.
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

// ProvisionProxyOptions stores the current defaults used when synchronizing ProxySQL tenant accounts.
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
func SyncProxyUser(ctx context.Context, proxyDSN, schema, user, pass string, opts ProxyOptions) error {
	db, err := sql.Open("mysql", normalizeProxyDSN(proxyDSN))
	if err != nil {
		return err
	}
	defer db.Close()

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

// DropProxyUser removes the mysql_users record for a tenant and reloads ProxySQL runtime/disk.
func DropProxyUser(ctx context.Context, proxyDSN, user string) error {
	db, err := sql.Open("mysql", normalizeProxyDSN(proxyDSN))
	if err != nil {
		return err
	}
	defer db.Close()

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

// normalizeProxyDSN adds interpolateParams to ProxySQL admin DSNs when missing so prepared statements work.
func normalizeProxyDSN(proxyDsn string) string {
	if proxyDsn == "" || strings.Contains(proxyDsn, "interpolateParams=") {
		return proxyDsn
	}

	sep := "?"
	if strings.Contains(proxyDsn, "?") {
		if strings.HasSuffix(proxyDsn, "?") || strings.HasSuffix(proxyDsn, "&") {
			sep = ""
		} else {
			sep = "&"
		}
	}

	return proxyDsn + sep + "interpolateParams=true"
}
