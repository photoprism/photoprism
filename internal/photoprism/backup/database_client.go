package backup

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/clean"
)

// mariadbPasswordEnv is the environment variable the MariaDB and MySQL clients read a password from.
const mariadbPasswordEnv = "MYSQL_PWD" // #nosec G101 environment variable name, not a credential

// mariadbConn holds the values a MariaDB client command connects with. Ssl is set unless the server is known
// not to offer zero-configuration TLS, SslVerify if the client verifies its certificate, and SslRequest if a
// MariaDB client is asked to use TLS without verification, which it only does if the server offers it.
type mariadbConn struct {
	Bin        string
	Socket     string
	Host       string
	Port       string
	User       string
	Name       string
	Password   string
	Ssl        bool
	SslUnknown bool
	SslVerify  bool
	SslRequest bool
}

// newMariadbConn returns the connection values of the configured database for the specified client.
func newMariadbConn(c *config.Config, bin string) mariadbConn {
	conn := mariadbConn{
		Bin:      bin,
		User:     c.DatabaseUser(),
		Name:     c.DatabaseName(),
		Password: c.DatabasePassword(),
		Ssl:      c.DatabaseSsl(),
	}

	// TLS is only skipped for a server known to be older than MariaDB 11.4.
	if !conn.Ssl && c.DatabaseVersion() == "" {
		conn.Ssl, conn.SslUnknown = true, true
	}

	conn.Socket, conn.Host, conn.Port = mariadbTarget(c.DatabaseServer(), c.DatabaseHost(), c.DatabasePortString())

	if conn.Socket == "" {
		conn.setClientSsl(mariadbClientVersion(bin))
	}

	return conn
}

// setClientSsl sets how a client of the specified version uses TLS. A MariaDB 11.4+ client verifies the
// zero-configuration certificate against the password; otherwise a MariaDB client requests TLS without
// verification, so it still encrypts the connection where the server offers TLS. A client that could not
// be checked or identified is asked to verify, so the connection fails instead of going unverified.
func (conn *mariadbConn) setClientSsl(major, minor int, kind clientKind) {
	switch zeroConf := major > 11 || major == 11 && minor >= 4; {
	case kind == clientOther:
		return
	case kind == clientUnknown:
		conn.SslVerify = conn.Ssl && conn.Password != ""
	case conn.Ssl && zeroConf:
		conn.SslVerify = conn.Password != ""
	default:
		conn.SslRequest = true
	}
}

// mariadbTarget returns what a client connects to: a socket path, or a host and port.
func mariadbTarget(server, host, port string) (socket, h, p string) {
	if strings.HasPrefix(server, "/") {
		return server, "", ""
	}

	return "", host, port
}

// String returns the connection for logging, without the password it carries.
func (conn mariadbConn) String() string {
	if conn.Socket != "" {
		return conn.User + "@" + conn.Socket + "/" + conn.Name
	}

	return conn.User + "@" + conn.Host + ":" + conn.Port + "/" + conn.Name
}

// Cmd returns the client command for the connection, with the specified flags and the database name
// appended. The password is passed in the environment, keeping it out of the command, and --no-defaults
// comes first, as the client requires. A MariaDB 11.4+ client is passed --ssl-verify-server-cert, so it
// verifies the zero-configuration certificate against the password it reads from the environment.
func (conn mariadbConn) Cmd(args ...string) *exec.Cmd {
	a := make([]string, 0, len(args)+10)
	a = append(a, "--no-defaults")

	if conn.Socket != "" {
		a = append(a, "--protocol", "socket", "-S", conn.Socket)
	} else {
		a = append(a, "--protocol", "tcp")

		switch {
		case conn.Ssl && conn.SslVerify && conn.Password != "":
			a = append(a, "--ssl-verify-server-cert")
		case conn.SslRequest:
			a = append(a, "--ssl", "--skip-ssl-verify-server-cert")
		case !conn.Ssl:
			a = append(a, "--skip-ssl")
		}

		a = append(a, "-h", conn.Host, "-P", conn.Port)
	}

	a = append(a, "-u", conn.User)
	a = append(a, args...)
	a = append(a, conn.Name)

	cmd := exec.Command(conn.Bin, a...) // #nosec G204 database connection parameters from trusted config
	cmd.Env = conn.env()

	return cmd
}

// env returns the environment the client runs with: the current one, carrying the connection's
// password, or none where it has no password. The variable is removed before it is set, so the value
// a client receives is the configured one and never a value this process inherited.
func (conn mariadbConn) env() []string {
	parent := os.Environ()
	env := make([]string, 0, len(parent)+1)

	for _, v := range parent {
		if !strings.HasPrefix(v, mariadbPasswordEnv+"=") {
			env = append(env, v)
		}
	}

	if conn.Password == "" {
		return env
	}

	return append(env, mariadbPasswordEnv+"="+conn.Password)
}

// clientDiagnostics logs the leading warnings a client wrote to stderr and returns its remaining lines,
// each sanitized with the password masked. Warnings are logged whether or not the client succeeded, so
// that a connection without verified TLS is always reported.
func clientDiagnostics(stderr, password, action string) (lines []string) {
	for _, line := range strings.Split(clean.Secrets(stderr, password), "\n") {
		if line = strings.TrimSpace(line); line == "" {
			continue
		} else if len(lines) == 0 && strings.HasPrefix(line, "WARNING:") {
			log.Warnf("%s: %s", action, clean.ErrorFull(errors.New(strings.TrimPrefix(line, "WARNING:"))))
		} else {
			lines = append(lines, clean.ErrorFull(errors.New(line)))
		}
	}

	return lines
}

// clientError returns the diagnostics a failed client wrote to stderr as an error, joined with "; ", or
// nil if there are none besides its warnings.
func clientError(stderr, password, action string) error {
	if lines := clientDiagnostics(stderr, password, action); len(lines) > 0 {
		return errors.New(strings.Join(lines, "; "))
	}

	return nil
}

// logDatabaseSsl reports whether the server offers zero-configuration TLS, which decides how the
// client connects and where it reads the password.
// see https://mariadb.org/mission-impossible-zero-configuration-ssl/
func logDatabaseSsl(conn mariadbConn, action string) {
	switch {
	case conn.Socket != "":
		return
	case !conn.Ssl && conn.SslRequest:
		log.Warnf("%s: zero-configuration ssl not supported by the server, using unverified ssl if available", action)
		return
	case !conn.Ssl:
		log.Warnf("%s: zero-configuration ssl not supported by the server", action)
		return
	}

	subject := "server supports zero-configuration ssl"

	if conn.SslUnknown {
		subject = "server version unknown, expecting zero-configuration ssl"
	}

	switch {
	case conn.Password == "":
		log.Warnf("%s: %s, but it cannot be verified without a password", action, subject)
	case !conn.SslVerify:
		log.Warnf("%s: %s, but %s cannot verify it", action, subject, filepath.Base(conn.Bin))
	default:
		log.Infof("%s: %s", action, subject)
	}
}

// mariadbClientVersionRegexp matches the server version a MariaDB client was built with in its --version output.
var mariadbClientVersionRegexp = regexp.MustCompile(`(\d+)\.(\d+)\.\d+(?:-\d+)?-MariaDB`)

// mysqlClientVersionRegexp matches the --version output of MySQL and Percona clients, e.g.
// "Ver 8.0.36 for Linux on x86_64 (MySQL Community Server - GPL)" or "Ver 14.14 Distrib 5.7.44, for Linux".
var mysqlClientVersionRegexp = regexp.MustCompile(`(?i)\(MySQL [A-Za-z ]+Server|Percona Server|Ver \d+\.\d+\.\d+[-+\w.]* for |Distrib \d+\.\d+\.\d+(?:-[\w.]+)?,`)

// clientKind is what a client version check found out about a client.
type clientKind int

const (
	clientUnknown clientKind = iota // the client could not be checked or identified
	clientOther                     // the client is identified as a MySQL or Percona client
	clientMariadb                   // the client is a MariaDB client
)

// clientVersion is the cached result of a client version check.
type clientVersion struct {
	major, minor int
	kind         clientKind
}

// clientVersions caches client version checks, keyed by binary.
var clientVersions sync.Map

// mariadbClientVersion returns the major and minor version of a MariaDB client, and what kind of client bin
// is. MariaDB 11.4+ clients verify zero-configuration TLS certificates against the password. A completed
// check is cached for each binary; one that failed is not, so the next command checks again.
// see https://mariadb.org/mission-impossible-zero-configuration-ssl/
func mariadbClientVersion(bin string) (major, minor int, kind clientKind) {
	if v, ok := clientVersions.Load(bin); ok {
		cv := v.(clientVersion)
		return cv.major, cv.minor, cv.kind
	}

	ctx, cancel := context.WithTimeout(context.Background(), clientProbeTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin, "--no-defaults", "--version") // #nosec G204 configured client binary
	cmd.WaitDelay = clientProbeWaitDelay
	out, err := cmd.Output()

	// A client that exited while a child kept its output open still ran.
	if ctx.Err() != nil {
		err = ctx.Err()
	} else if errors.Is(err, exec.ErrWaitDelay) {
		err = nil
	}

	if err != nil {
		log.Warnf("database: failed to check the version of %s (%s)", filepath.Base(bin), clean.Error(err))
		return 0, 0, clientUnknown
	}

	var cv clientVersion

	if m := mariadbClientVersionRegexp.FindStringSubmatch(string(out)); len(m) == 3 {
		cv.major, _ = strconv.Atoi(m[1])
		cv.minor, _ = strconv.Atoi(m[2])
		cv.kind = clientMariadb
	} else if mysqlClientVersionRegexp.Match(out) {
		cv.kind = clientOther
	} else {
		log.Warnf("database: failed to identify the version of %s", filepath.Base(bin))
	}

	clientVersions.Store(bin, cv)

	return cv.major, cv.minor, cv.kind
}
