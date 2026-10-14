package backup

import (
	"context"
	"net"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
)

// postgresEnvParams maps the connection parameters a PostgreSQL client command is passed to the libpq
// environment variables that set them.
var postgresEnvParams = map[string]string{
	"sslmode":                  "PGSSLMODE",
	"sslrootcert":              "PGSSLROOTCERT",
	"sslcert":                  "PGSSLCERT",
	"sslkey":                   "PGSSLKEY",
	"sslcrl":                   "PGSSLCRL",
	"sslcertmode":              "PGSSLCERTMODE",
	"sslsni":                   "PGSSLSNI",
	"sslnegotiation":           "PGSSLNEGOTIATION",
	"channel_binding":          "PGCHANNELBINDING",
	"gssencmode":               "PGGSSENCMODE",
	"target_session_attrs":     "PGTARGETSESSIONATTRS",
	"options":                  "PGOPTIONS",
	"require_auth":             "PGREQUIREAUTH",
	"ssl_min_protocol_version": "PGSSLMINPROTOCOLVERSION",
	"ssl_max_protocol_version": "PGSSLMAXPROTOCOLVERSION",
	"passfile":                 "PGPASSFILE",
	"connect_timeout":          "PGCONNECT_TIMEOUT",
}

// postgresServiceEnv lists the environment variables that would select a connection from a service file or
// address, which a client command never reads, as the app does not either.
var postgresServiceEnv = []string{"PGHOSTADDR", "PGSERVICE", "PGSERVICEFILE"}

// postgresRestrictVersions lists the first release of each major version whose psql can restrict backslash
// commands with "\restrict"; later major versions all can.
var postgresRestrictVersions = map[int]int{13: 22, 14: 19, 15: 14, 16: 10, 17: 6}

// postgresVersionRegexp matches the version a PostgreSQL client prints, e.g. "psql (PostgreSQL) 18.6".
var postgresVersionRegexp = regexp.MustCompile(`\(PostgreSQL\) (\d+)\.(\d+)`)

// postgresConn holds the values a PostgreSQL client command connects with. Host is a host name or address, or
// the directory of a Unix socket, and Params holds the connection parameters listed in postgresEnvParams.
type postgresConn struct {
	Bin      string
	Host     string
	Port     string
	User     string
	Name     string
	Password string
	Params   map[string]string
}

// newPostgresConn returns the connection values of the configured database for the specified client.
func newPostgresConn(c *config.Config, bin string) postgresConn {
	d := c.DatabaseDSN()
	parsed := dsn.Parse(d)

	conn := postgresConn{
		Bin:      bin,
		Host:     c.DatabaseHost(),
		User:     c.DatabaseUser(),
		Name:     c.DatabaseName(),
		Password: c.DatabasePassword(),
		Params:   postgresParams(parsed.Params, strings.Contains(d, "://")),
	}

	// A configured DSN may leave the port, user, database and password to the environment, as for the app;
	// a generated one names them.
	if c.HasDatabaseDSN() {
		conn.Port, conn.User, conn.Name, conn.Password = "", parsed.User, parsed.Name, parsed.Password

		if _, port, err := net.SplitHostPort(c.DatabaseServer()); err == nil && port != "" {
			conn.Port = port
		}
	} else {
		conn.Port = c.DatabasePortString()
	}

	// Like the driver of the app, a client verifies the server certificate with the system's certificates
	// unless the configuration names one. A client accepts them only for full verification.
	if conn.Params["sslmode"] == "verify-full" && conn.Params["sslrootcert"] == "" && os.Getenv("PGSSLROOTCERT") == "" {
		conn.Params["sslrootcert"] = "system"
	}

	return conn
}

// postgresParams returns the connection parameters a client is passed from the parameters of a DSN, given as
// a URL query or as space-separated key=value pairs with optionally quoted values.
func postgresParams(params string, query bool) map[string]string {
	result := make(map[string]string)

	if query {
		values, _ := url.ParseQuery(params)

		for key := range postgresEnvParams {
			if v := values.Get(key); v != "" {
				result[key] = v
			}
		}

		return result
	}

	for _, pair := range splitPostgresParams(params) {
		if key, value, found := strings.Cut(pair, "="); found && postgresEnvParams[key] != "" && value != "" {
			result[key] = value
		}
	}

	return result
}

// splitPostgresParams splits space-separated key=value pairs, keeping spaces in quoted values and removing
// the quotes.
func splitPostgresParams(s string) (pairs []string) {
	var b strings.Builder

	quote := rune(0)

	for _, r := range s {
		switch {
		case quote != 0 && r == quote:
			quote = 0
		case quote == 0 && (r == '"' || r == '\''):
			quote = r
		case quote == 0 && (r == ' ' || r == '\t' || r == '\n'):
			if b.Len() > 0 {
				pairs = append(pairs, b.String())
				b.Reset()
			}
		default:
			b.WriteRune(r)
		}
	}

	if b.Len() > 0 {
		pairs = append(pairs, b.String())
	}

	return pairs
}

// postgresRestrictSupported reports whether a psql client can restrict backslash commands, based on the
// version it prints.
func postgresRestrictSupported(bin string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), clientProbeTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin, "--version") // #nosec G204 configured client binary
	cmd.WaitDelay = clientProbeWaitDelay

	out, err := cmd.Output()

	if err != nil {
		return false, err
	}

	m := postgresVersionRegexp.FindSubmatch(out)

	if m == nil {
		return false, nil
	}

	major, _ := strconv.Atoi(string(m[1]))
	minor, _ := strconv.Atoi(string(m[2]))

	if first, ok := postgresRestrictVersions[major]; ok {
		return minor >= first, nil
	}

	return major > 17, nil
}

// String returns the connection for logging, without the password it carries.
func (conn postgresConn) String() string {
	if conn.Port == "" {
		return conn.User + "@" + conn.Host + "/" + conn.Name
	}

	return conn.User + "@" + net.JoinHostPort(conn.Host, conn.Port) + "/" + conn.Name
}

// Cmd returns the client command for the connection with the specified flags. The connection is passed in the
// environment, which keeps the password out of the command.
func (conn postgresConn) Cmd(args ...string) *exec.Cmd {
	cmd := exec.Command(conn.Bin, args...) // #nosec G204 database connection parameters from trusted config
	cmd.Env = conn.env()

	return cmd
}

// env returns the environment the client runs with: the current one, as the driver of the app also reads its
// defaults from it, with the server, port, database, user and password of the connection where it has them,
// and the parameters of the connection, which take precedence over inherited values.
func (conn postgresConn) env() []string {
	values := [][2]string{{"PGHOST", conn.Host}, {"PGPORT", conn.Port}, {"PGUSER", conn.User},
		{"PGDATABASE", conn.Name}, {"PGPASSWORD", conn.Password}}

	parent := os.Environ()
	env := make([]string, 0, len(parent)+len(values)+len(conn.Params))

	for _, v := range parent {
		name, _, _ := strings.Cut(v, "=")

		if slices.Contains(postgresServiceEnv, name) || slices.ContainsFunc(values, func(kv [2]string) bool {
			return kv[0] == name && kv[1] != ""
		}) {
			continue
		}

		env = append(env, v)
	}

	for _, kv := range values {
		if kv[1] != "" {
			env = append(env, kv[0]+"="+kv[1])
		}
	}

	for key, value := range conn.Params {
		env = append(env, postgresEnvParams[key]+"="+value)
	}

	return env
}

// privilegeCheckTimeout bounds the query that checks the privileges of the user a restore connects as.
var privilegeCheckTimeout = 5 * time.Second

// postgresServerPrivileges selects the session user, and whether it may create roles or is a member of a
// superuser role or of a predefined role that grants access to server programs or files.
const postgresServerPrivileges = "SELECT session_user, COALESCE((SELECT rolcreaterole FROM pg_roles " +
	"WHERE rolname = session_user), FALSE) OR EXISTS (SELECT 1 FROM pg_roles r WHERE (r.rolsuper OR " +
	"r.rolname IN ('pg_execute_server_program', 'pg_read_server_files', 'pg_write_server_files')) " +
	"AND pg_has_role(session_user, r.oid, 'MEMBER'))"

// warnPostgresServerPrivileges logs a warning if the database user has privileges beyond the database, as
// restoring with a role that only owns the database is recommended. The restore does not depend on the check.
func warnPostgresServerPrivileges(db *gorm.DB) {
	if db == nil {
		log.Debugf("restore: skipped checking the role of the database user, as the database is not connected")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), privilegeCheckTimeout)
	defer cancel()

	sqlDb, err := db.DB()

	if err != nil {
		log.Warnf("restore: failed to check the role of the database user (%s)", clean.Error(err))
		return
	}

	var user string
	var privileged bool

	if err = sqlDb.QueryRowContext(ctx, postgresServerPrivileges).Scan(&user, &privileged); err != nil {
		log.Warnf("restore: failed to check the role of the database user (%s)", clean.Error(err))
	} else if privileged {
		log.Warnf("restore: database user %s has server-side privileges, restoring with a role that only owns "+
			"the database is recommended", clean.Log(user))
	}
}

// postgresDumpArgs returns the flags a PostgreSQL dump is created with: a plain SQL script that drops the
// existing objects before it creates them, without owners and privileges, so it restores as any user.
func postgresDumpArgs() []string {
	return []string{"--no-password", "--format=plain", "--clean", "--if-exists", "--no-owner", "--no-privileges"}
}

// postgresRestoreArgs returns the flags a PostgreSQL dump is restored with: read from the standard input in a
// single transaction that stops at the first error, so a failed restore changes nothing, and without the
// output of the statements or the statement and values in an error message.
func postgresRestoreArgs() []string {
	return []string{"--no-password", "--no-psqlrc", "--quiet", "--set=ON_ERROR_STOP=1", "--set=VERBOSITY=terse",
		"--set=SHOW_CONTEXT=never", "--single-transaction", "--output=" + os.DevNull, "--file=-"}
}
