package backup

import (
	"os"
	"os/exec"
	"strings"

	"github.com/photoprism/photoprism/internal/config"
)

// mariadbPasswordEnv is the environment variable the MariaDB and MySQL clients read a password from.
const mariadbPasswordEnv = "MYSQL_PWD" // #nosec G101 environment variable name, not a credential

// mariadbConn holds the values a MariaDB client command connects with.
type mariadbConn struct {
	Bin      string
	Socket   string
	Host     string
	Port     string
	User     string
	Name     string
	Password string
	Ssl      bool
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
// appended. The password is passed in the environment, where the client reads it without it becoming
// part of the command, and --no-defaults comes first, which the client requires and which keeps an
// option file from supplying a credential of its own.
func (conn mariadbConn) Cmd(args ...string) *exec.Cmd {
	a := make([]string, 0, len(args)+10)
	a = append(a, "--no-defaults")

	if conn.Socket != "" {
		a = append(a, "--protocol", "socket", "-S", conn.Socket)
	} else {
		a = append(a, "--protocol", "tcp")

		if !conn.Ssl {
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

// logDatabaseSsl reports whether the server offers zero-configuration TLS, which decides how the
// client connects and where it reads the password.
// see https://mariadb.org/mission-impossible-zero-configuration-ssl/
func logDatabaseSsl(conn mariadbConn, action string) {
	switch {
	case conn.Socket != "":
		return
	case conn.Ssl:
		log.Infof("%s: server supports zero-configuration ssl", action)
	default:
		log.Infof("%s: zero-configuration ssl not supported by the server", action)
	}
}
