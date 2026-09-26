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

// mariadbConn holds the values a MariaDB client command connects with. Ssl is set if the server offers
// zero-configuration TLS, and SslVerify if the client can verify its certificate.
type mariadbConn struct {
	Bin       string
	Socket    string
	Host      string
	Port      string
	User      string
	Name      string
	Password  string
	Ssl       bool
	SslVerify bool
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

	conn.Socket, conn.Host, conn.Port = mariadbTarget(c.DatabaseServer(), c.DatabaseHost(), c.DatabasePortString())

	if conn.Ssl && conn.Socket == "" && conn.Password != "" {
		conn.SslVerify = clientVerifiesSsl(bin)
	}

	return conn
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

		if !conn.Ssl {
			a = append(a, "--skip-ssl")
		} else if conn.SslVerify && conn.Password != "" {
			a = append(a, "--ssl-verify-server-cert")
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

// clientError returns the diagnostics a failed client wrote to stderr as an error, or nil if there are
// none. Leading warnings are logged on their own, the remaining lines are joined with "; ", and each
// line is sanitized with the password masked.
func clientError(stderr, password, action string) error {
	var lines []string

	for _, line := range strings.Split(clean.Secrets(stderr, password), "\n") {
		if line = strings.TrimSpace(line); line == "" {
			continue
		} else if len(lines) == 0 && strings.HasPrefix(line, "WARNING:") {
			log.Warnf("%s: %s", action, clean.ErrorFull(errors.New(strings.TrimPrefix(line, "WARNING:"))))
		} else {
			lines = append(lines, clean.ErrorFull(errors.New(line)))
		}
	}

	if len(lines) == 0 {
		return nil
	}

	return errors.New(strings.Join(lines, "; "))
}

// logDatabaseSsl reports whether the server offers zero-configuration TLS, which decides how the
// client connects and where it reads the password.
// see https://mariadb.org/mission-impossible-zero-configuration-ssl/
func logDatabaseSsl(conn mariadbConn, action string) {
	switch {
	case conn.Socket != "":
		return
	case conn.Ssl && conn.Password == "":
		log.Warnf("%s: server supports zero-configuration ssl, but it cannot be verified without a password", action)
	case conn.Ssl && !conn.SslVerify:
		log.Warnf("%s: server supports zero-configuration ssl, but %s cannot verify it", action, filepath.Base(conn.Bin))
	case conn.Ssl:
		log.Infof("%s: server supports zero-configuration ssl", action)
	default:
		log.Infof("%s: zero-configuration ssl not supported by the server", action)
	}
}

// mariadbClientVersion matches the server version a MariaDB client was built with in its --version output.
var mariadbClientVersion = regexp.MustCompile(`(\d+)\.(\d+)\.\d+-MariaDB`)

// clientVersions caches the output of client version checks, keyed by binary.
var clientVersions sync.Map

// clientVerifiesSsl reports whether bin is a MariaDB 11.4+ client, which verifies zero-configuration
// TLS certificates against the password. The result is cached for each binary.
// see https://mariadb.org/mission-impossible-zero-configuration-ssl/
func clientVerifiesSsl(bin string) bool {
	if v, ok := clientVersions.Load(bin); ok {
		return v.(bool)
	}

	ctx, cancel := context.WithTimeout(context.Background(), clientProbeTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin, "--no-defaults", "--version") // #nosec G204 configured client binary
	cmd.WaitDelay = clientProbeWaitDelay
	out, _ := cmd.Output()

	// A check that timed out is not cached, so the next command checks again.
	if ctx.Err() != nil {
		log.Warnf("database: failed to check the version of %s (%s)", filepath.Base(bin), ctx.Err())
		return false
	}

	result := false

	if m := mariadbClientVersion.FindStringSubmatch(string(out)); len(m) == 3 {
		major, _ := strconv.Atoi(m[1])
		minor, _ := strconv.Atoi(m[2])
		result = major > 11 || major == 11 && minor >= 4
	}

	clientVersions.Store(bin, result)

	return result
}
