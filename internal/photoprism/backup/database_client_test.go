package backup

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
)

// mariadbTestConn returns connection values for a TCP client command.
func mariadbTestConn() mariadbConn {
	return mariadbConn{
		Bin:      "/usr/bin/mariadb-dump",
		Host:     "mariadb",
		Port:     "4001",
		User:     "photoprism",
		Name:     "photoprism",
		Password: "Sup3r$ecret",
	}
}

func TestMariadbConn_Cmd(t *testing.T) {
	t.Run("Tcp", func(t *testing.T) {
		cmd := mariadbTestConn().Cmd()
		assert.Equal(t, []string{"/usr/bin/mariadb-dump", "--no-defaults", "--protocol", "tcp", "--skip-ssl",
			"-h", "mariadb", "-P", "4001", "-u", "photoprism", "photoprism"}, cmd.Args)
		assert.Contains(t, cmd.Env, "MYSQL_PWD=Sup3r$ecret")
	})
	t.Run("TcpSsl", func(t *testing.T) {
		// A server offering zero-configuration TLS keeps the connection encrypted without --skip-ssl.
		conn := mariadbTestConn()
		conn.Ssl = true
		cmd := conn.Cmd()
		assert.Equal(t, []string{"/usr/bin/mariadb-dump", "--no-defaults", "--protocol", "tcp",
			"-h", "mariadb", "-P", "4001", "-u", "photoprism", "photoprism"}, cmd.Args)
		assert.Contains(t, cmd.Env, "MYSQL_PWD=Sup3r$ecret")
	})
	t.Run("NoPasswordArgumentOnAnyBranch", func(t *testing.T) {
		for _, tc := range []struct {
			socket string
			ssl    bool
		}{
			{"", false}, {"", true}, {"/run/mysqld/mysqld.sock", false}, {"/run/mysqld/mysqld.sock", true},
		} {
			conn := mariadbTestConn()
			conn.Socket, conn.Ssl = tc.socket, tc.ssl
			if tc.socket != "" {
				conn.Host, conn.Port = "", ""
			}
			cmd := conn.Cmd("-f")
			for _, arg := range cmd.Args {
				assert.False(t, strings.HasPrefix(arg, "-p"), "unexpected argument %s", arg)
				assert.NotContains(t, arg, conn.Password)
			}
			assert.Contains(t, cmd.Env, "MYSQL_PWD="+conn.Password)
		}
	})
	t.Run("Socket", func(t *testing.T) {
		conn := mariadbTestConn()
		conn.Socket, conn.Host, conn.Port = "/run/mysqld/mysqld.sock", "", ""
		cmd := conn.Cmd()
		assert.Equal(t, []string{"/usr/bin/mariadb-dump", "--no-defaults", "--protocol", "socket",
			"-S", "/run/mysqld/mysqld.sock", "-u", "photoprism", "photoprism"}, cmd.Args)
		assert.Contains(t, cmd.Env, "MYSQL_PWD=Sup3r$ecret")
	})
	t.Run("SocketSsl", func(t *testing.T) {
		// A socket deployment on a server that offers zero-configuration TLS still connects by socket.
		conn := mariadbTestConn()
		conn.Socket, conn.Host, conn.Port = "/run/mysqld/mysqld.sock", "", ""
		conn.Ssl = true
		cmd := conn.Cmd()
		assert.Equal(t, []string{"/usr/bin/mariadb-dump", "--no-defaults", "--protocol", "socket",
			"-S", "/run/mysqld/mysqld.sock", "-u", "photoprism", "photoprism"}, cmd.Args)
		assert.Contains(t, cmd.Env, "MYSQL_PWD=Sup3r$ecret")
	})
	t.Run("EnvironmentExtended", func(t *testing.T) {
		// The client keeps the environment it was started with, whatever else that environment holds.
		t.Setenv("ZZ_BACKUP_TEST_VAR", "kept")
		cmd := mariadbTestConn().Cmd()
		assert.Subset(t, cmd.Env, []string{"ZZ_BACKUP_TEST_VAR=kept", "PATH=" + os.Getenv("PATH")})
		assert.Contains(t, cmd.Env, mariadbPasswordEnv+"=Sup3r$ecret")
	})
	t.Run("RestoreFlags", func(t *testing.T) {
		conn := mariadbTestConn()
		conn.Bin = "/usr/bin/mariadb"
		cmd := conn.Cmd("-f")
		assert.Equal(t, []string{"/usr/bin/mariadb", "--no-defaults", "--protocol", "tcp", "--skip-ssl",
			"-h", "mariadb", "-P", "4001", "-u", "photoprism", "-f", "photoprism"}, cmd.Args)
	})
	t.Run("NoPassword", func(t *testing.T) {
		// No credential is passed, in an argument or in the environment, and the option files that
		// could supply one are suppressed.
		for _, ssl := range []bool{false, true} {
			conn := mariadbTestConn()
			conn.Password, conn.Ssl = "", ssl
			cmd := conn.Cmd()
			for _, arg := range cmd.Args {
				assert.False(t, strings.HasPrefix(arg, "-p"), "unexpected argument %s with ssl %v", arg, ssl)
			}
			assert.Equal(t, "--no-defaults", cmd.Args[1])
			assert.NotContains(t, strings.Join(cmd.Env, " "), mariadbPasswordEnv)
		}
	})
	t.Run("FirstArgumentIsNoDefaults", func(t *testing.T) {
		// The client accepts --no-defaults as its first argument only.
		cmd := mariadbTestConn().Cmd()
		require.Greater(t, len(cmd.Args), 1)
		assert.Equal(t, "--no-defaults", cmd.Args[1])
	})
	t.Run("TraceOmitsThePassword", func(t *testing.T) {
		conn := mariadbTestConn()
		assert.NotContains(t, clean.Cmd(conn.Cmd(), conn.Password), conn.Password)
		conn.Ssl = true
		assert.NotContains(t, clean.Cmd(conn.Cmd(), conn.Password), conn.Password)
	})
	t.Run("InheritedPasswordRemoved", func(t *testing.T) {
		// A value this process inherited must not stand in for the configured one, in either direction.
		t.Setenv(mariadbPasswordEnv, "inherited-value")
		conn := mariadbTestConn()
		conn.Password = ""
		assert.NotContains(t, strings.Join(conn.Cmd().Env, " "), "inherited-value")
		conn.Password = "Sup3r$ecret"
		env := conn.Cmd().Env
		assert.NotContains(t, strings.Join(env, " "), "inherited-value")
		var found int
		for _, v := range env {
			if strings.HasPrefix(v, mariadbPasswordEnv+"=") {
				found++
				assert.Equal(t, mariadbPasswordEnv+"=Sup3r$ecret", v)
			}
		}
		assert.Equal(t, 1, found, "expected exactly one %s entry", mariadbPasswordEnv)
	})
}

func TestNewMariadbConn(t *testing.T) {
	c := get.Config()
	conn := newMariadbConn(c, "/usr/bin/mariadb-dump")

	assert.Equal(t, "/usr/bin/mariadb-dump", conn.Bin)
	assert.Equal(t, c.DatabaseUser(), conn.User)
	assert.Equal(t, c.DatabaseName(), conn.Name)
	assert.Equal(t, c.DatabasePassword(), conn.Password)
	assert.Equal(t, c.DatabaseSsl(), conn.Ssl)

	if c.DatabaseDriver() == dsn.DriverSQLite3 {
		// SQLite has no server, user or password, so no client command reads one.
		assert.Empty(t, conn.Password)
		assert.Empty(t, conn.Socket)
		return
	}

	if strings.HasPrefix(c.DatabaseServer(), "/") {
		assert.Equal(t, c.DatabaseServer(), conn.Socket)
		assert.Empty(t, conn.Host)
	} else {
		assert.Empty(t, conn.Socket)
		assert.Equal(t, c.DatabaseHost(), conn.Host)
		assert.Equal(t, c.DatabasePortString(), conn.Port)
	}
}

func TestMariadbTarget(t *testing.T) {
	for _, tc := range []struct{ name, server, host, port, socket, wantHost, wantPort string }{
		{"Socket", "/run/mysqld/mysqld.sock", "", "", "/run/mysqld/mysqld.sock", "", ""},
		{"Host", "mariadb:4001", "mariadb", "4001", "", "mariadb", "4001"},
		{"HostWithoutPort", "mariadb", "mariadb", "3306", "", "mariadb", "3306"},
		{"Empty", "", "", "", "", "", ""},
		{"RelativePath", "run/mysqld.sock", "run/mysqld.sock", "3306", "", "run/mysqld.sock", "3306"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			socket, host, port := mariadbTarget(tc.server, tc.host, tc.port)
			assert.Equal(t, tc.socket, socket)
			assert.Equal(t, tc.wantHost, host)
			assert.Equal(t, tc.wantPort, port)
		})
	}
}

func TestLogDatabaseSsl(t *testing.T) {
	t.Run("Ssl", func(t *testing.T) {
		hook := captureLog(t)
		conn := mariadbTestConn()
		conn.Ssl = true
		logDatabaseSsl(conn, "backup")
		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, "backup: server supports zero-configuration ssl", hook.LastEntry().Message)
	})
	t.Run("NoSsl", func(t *testing.T) {
		hook := captureLog(t)
		logDatabaseSsl(mariadbTestConn(), "restore")
		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, "restore: zero-configuration ssl not supported by the server", hook.LastEntry().Message)
	})
	t.Run("Socket", func(t *testing.T) {
		// A socket connection reports nothing, as no zero-configuration TLS decision applies.
		hook := captureLog(t)
		conn := mariadbTestConn()
		conn.Socket, conn.Ssl = "/run/mysqld/mysqld.sock", true
		logDatabaseSsl(conn, "backup")
		assert.Empty(t, hook.AllEntries())
	})
}

func TestMariadbConn_String(t *testing.T) {
	t.Run("Tcp", func(t *testing.T) {
		conn := mariadbTestConn()
		assert.Equal(t, "photoprism@mariadb:4001/photoprism", conn.String())
		assert.NotContains(t, conn.String(), conn.Password)
	})
	t.Run("Socket", func(t *testing.T) {
		conn := mariadbTestConn()
		conn.Socket, conn.Host, conn.Port = "/run/mysqld/mysqld.sock", "", ""
		assert.Equal(t, "photoprism@/run/mysqld/mysqld.sock/photoprism", conn.String())
		assert.NotContains(t, conn.String(), conn.Password)
	})
	t.Run("Formatted", func(t *testing.T) {
		// A struct rendered with a verb that would otherwise print every field.
		conn := mariadbTestConn()
		assert.NotContains(t, fmt.Sprintf("%+v", conn), conn.Password)
	})
}

// runWithConnEnv returns the environment a child process receives from the connection.
func runWithConnEnv(t *testing.T, conn mariadbConn) string {
	t.Helper()

	cmd := exec.Command("/usr/bin/env")
	cmd.Env = conn.Cmd().Env
	out, err := cmd.Output()
	require.NoError(t, err)

	return string(out)
}

func TestMariadbConn_ChildEnvironment(t *testing.T) {
	t.Setenv(mariadbPasswordEnv, "inherited-value")
	t.Run("Password", func(t *testing.T) {
		out := runWithConnEnv(t, mariadbTestConn())
		assert.Contains(t, out, mariadbPasswordEnv+"=Sup3r$ecret\n")
		assert.NotContains(t, out, "inherited-value")
	})
	t.Run("NoPassword", func(t *testing.T) {
		conn := mariadbTestConn()
		conn.Password = ""
		out := runWithConnEnv(t, conn)
		assert.NotContains(t, out, mariadbPasswordEnv+"=")
		assert.NotContains(t, out, "inherited-value")
		assert.Contains(t, out, "PATH=")
	})
}
