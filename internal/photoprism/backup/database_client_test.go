package backup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/txt"
)

// mariadbTestConn returns connection values for a TCP client command.
func mariadbTestConn() mariadbConn {
	return mariadbConn{
		Bin:       "/usr/bin/mariadb-dump",
		Host:      "mariadb",
		Port:      "4001",
		User:      "photoprism",
		Name:      "photoprism",
		Password:  "Sup3r$ecret",
		SslVerify: true,
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
		// A server offering zero-configuration TLS gets its certificate verified against the password.
		conn := mariadbTestConn()
		conn.Ssl = true
		cmd := conn.Cmd()
		assert.Equal(t, []string{"/usr/bin/mariadb-dump", "--no-defaults", "--protocol", "tcp", "--ssl-verify-server-cert",
			"-h", "mariadb", "-P", "4001", "-u", "photoprism", "photoprism"}, cmd.Args)
		assert.Contains(t, cmd.Env, "MYSQL_PWD=Sup3r$ecret")
	})
	t.Run("TcpSslClientCannotVerify", func(t *testing.T) {
		// A client that is not a MariaDB client keeps its own TLS defaults.
		conn := mariadbTestConn()
		conn.Ssl, conn.SslVerify = true, false
		assert.Equal(t, []string{"/usr/bin/mariadb-dump", "--no-defaults", "--protocol", "tcp",
			"-h", "mariadb", "-P", "4001", "-u", "photoprism", "photoprism"}, conn.Cmd().Args)
	})
	t.Run("TcpSslRequest", func(t *testing.T) {
		// A client that cannot verify the certificate uses TLS without verification where the server offers it.
		for _, ssl := range []bool{true, false} {
			conn := mariadbTestConn()
			conn.Ssl, conn.SslVerify, conn.SslRequest = ssl, false, true
			assert.Equal(t, []string{"/usr/bin/mariadb-dump", "--no-defaults", "--protocol", "tcp", "--ssl", "--skip-ssl-verify-server-cert",
				"-h", "mariadb", "-P", "4001", "-u", "photoprism", "photoprism"}, conn.Cmd().Args, "ssl %v", ssl)
		}
	})
	t.Run("TcpSslNoPassword", func(t *testing.T) {
		// Without a password, zero-configuration TLS has nothing to verify the certificate against.
		conn := mariadbTestConn()
		conn.Ssl, conn.Password = true, ""
		assert.Equal(t, []string{"/usr/bin/mariadb-dump", "--no-defaults", "--protocol", "tcp",
			"-h", "mariadb", "-P", "4001", "-u", "photoprism", "photoprism"}, conn.Cmd().Args)
	})
	t.Run("VerifyOnlyOnTcpSsl", func(t *testing.T) {
		for _, tc := range []struct {
			socket, password                 string
			ssl, sslVerify, sslRequest, want bool
		}{
			{"", "Sup3r$ecret", true, true, false, true},
			{"", "Sup3r$ecret", true, false, false, false},
			{"", "Sup3r$ecret", false, true, false, false},
			{"", "", true, true, false, false},
			{"/run/mysqld/mysqld.sock", "Sup3r$ecret", true, true, false, false},
			{"/run/mysqld/mysqld.sock", "", true, true, false, false},
			{"", "Sup3r$ecret", true, false, true, false},
		} {
			conn := mariadbTestConn()
			conn.Socket, conn.Password, conn.Ssl, conn.SslVerify, conn.SslRequest = tc.socket, tc.password, tc.ssl, tc.sslVerify, tc.sslRequest
			if tc.want {
				assert.Contains(t, conn.Cmd("-f").Args, "--ssl-verify-server-cert", "%+v", tc)
			} else {
				assert.NotContains(t, conn.Cmd("-f").Args, "--ssl-verify-server-cert", "%+v", tc)
			}
		}
	})
	t.Run("RequestOnlyOnTcp", func(t *testing.T) {
		for _, tc := range []struct {
			socket string
			ssl    bool
			want   bool
		}{
			{"", true, true},
			{"", false, true},
			{"/run/mysqld/mysqld.sock", true, false},
		} {
			conn := mariadbTestConn()
			conn.Socket, conn.Ssl, conn.SslVerify, conn.SslRequest = tc.socket, tc.ssl, false, true
			if tc.want {
				assert.Contains(t, conn.Cmd().Args, "--ssl", "%+v", tc)
			} else {
				assert.NotContains(t, conn.Cmd().Args, "--ssl", "%+v", tc)
			}
		}
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
	assert.Equal(t, c.DatabaseSsl() || c.DatabaseVersion() == "", conn.Ssl)

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
		assert.Equal(t, logrus.InfoLevel, hook.LastEntry().Level)
	})
	t.Run("UnknownServerVersion", func(t *testing.T) {
		hook := captureLog(t)
		conn := mariadbTestConn()
		conn.Ssl, conn.SslUnknown = true, true
		logDatabaseSsl(conn, "backup")
		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, "backup: server version unknown, expecting zero-configuration ssl", hook.LastEntry().Message)
	})
	t.Run("ClientCannotVerify", func(t *testing.T) {
		hook := captureLog(t)
		conn := mariadbTestConn()
		conn.Ssl, conn.SslVerify = true, false
		logDatabaseSsl(conn, "backup")
		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, "backup: server supports zero-configuration ssl, but mariadb-dump cannot verify it", hook.LastEntry().Message)
		assert.Equal(t, logrus.WarnLevel, hook.LastEntry().Level)
	})
	t.Run("NoPassword", func(t *testing.T) {
		// Without a password, the client has nothing to verify the certificate against.
		hook := captureLog(t)
		conn := mariadbTestConn()
		conn.Ssl, conn.SslVerify, conn.Password = true, false, ""
		logDatabaseSsl(conn, "backup")
		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, "backup: server supports zero-configuration ssl, but it cannot be verified without a password", hook.LastEntry().Message)
		assert.Equal(t, logrus.WarnLevel, hook.LastEntry().Level)
	})
	t.Run("NoSslRequest", func(t *testing.T) {
		hook := captureLog(t)
		conn := mariadbTestConn()
		conn.Ssl, conn.SslRequest = false, true
		logDatabaseSsl(conn, "backup")
		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, "backup: zero-configuration ssl not supported by the server, using unverified ssl if available", hook.LastEntry().Message)
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

func TestClientDiagnostics(t *testing.T) {
	t.Run("Warnings", func(t *testing.T) {
		hook := captureLog(t)
		lines := clientDiagnostics("WARNING: a\nWARNING: b Sup3r$ecret\nerror c\nWARNING: d\n", "Sup3r$ecret", "backup")
		assert.Equal(t, []string{"error c", "WARNING: d"}, lines)
		assert.Equal(t, []string{"backup: a", "backup: b " + txt.Masked}, logMessages(hook))
	})
	t.Run("Empty", func(t *testing.T) {
		hook := captureLog(t)
		assert.Empty(t, clientDiagnostics("\n \n", "", "restore"))
		assert.Empty(t, hook.AllEntries())
	})
}

func TestClientError(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		hook := captureLog(t)
		assert.NoError(t, clientError(" \n", "Sup3r$ecret", "backup"))
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("WarningLogged", func(t *testing.T) {
		hook := captureLog(t)
		stderr := "WARNING: option --ssl-verify-server-cert is disabled, because of an insecure passwordless login.\n" +
			"mariadb-dump: Got error: 2005: \"Unknown server host 'mariadb'\" when trying to connect\n"
		err := clientError(stderr, "Sup3r$ecret", "backup")
		require.Error(t, err)
		assert.Equal(t, "mariadb-dump: Got error: 2005: 'Unknown server host 'mariadb'' when trying to connect", err.Error())
		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, "backup: option --ssl-verify-server-cert is disabled, because of an insecure passwordless login.", hook.LastEntry().Message)
		assert.Equal(t, logrus.WarnLevel, hook.LastEntry().Level)
	})
	t.Run("WarningOnly", func(t *testing.T) {
		// A failure that reports nothing but a warning falls back to the exit status.
		hook := captureLog(t)
		assert.NoError(t, clientError("WARNING: something\n", "", "restore"))
		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, "restore: something", hook.LastEntry().Message)
	})
	t.Run("LeadingWarningsOnly", func(t *testing.T) {
		// Server text after the client's error line stays part of the error.
		hook := captureLog(t)
		err := clientError("ERROR 1045 (28000): Access denied\nWARNING: x\nWARNING: y\n", "", "backup")
		require.Error(t, err)
		assert.Equal(t, "ERROR 1045 (28000): Access denied; WARNING: x; WARNING: y", err.Error())
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("ControlCharacters", func(t *testing.T) {
		hook := captureLog(t)
		err := clientError("WARNING: a\x1b[31mb\rc\nERROR 2026 (HY000): d\x1b[0me\r\n", "", "backup")
		require.Error(t, err)
		assert.NotContains(t, err.Error(), "\x1b")
		assert.NotContains(t, err.Error(), "\r")
		require.Len(t, hook.AllEntries(), 1)
		assert.NotContains(t, hook.LastEntry().Message, "\x1b")
		assert.NotContains(t, hook.LastEntry().Message, "\r")
	})
	t.Run("LinesJoined", func(t *testing.T) {
		err := clientError("ERROR 1045 (28000): Access denied\r\nsecond line\n", "", "restore")
		require.Error(t, err)
		assert.Equal(t, "ERROR 1045 (28000): Access denied; second line", err.Error())
	})
	t.Run("PasswordMasked", func(t *testing.T) {
		hook := captureLog(t)
		err := clientError("WARNING: using Sup3r$ecret\nfailed for Sup3r$ecret\n", "Sup3r$ecret", "backup")
		require.Error(t, err)
		assert.NotContains(t, err.Error(), "Sup3r$ecret")
		require.Len(t, hook.AllEntries(), 1)
		assert.NotContains(t, hook.LastEntry().Message, "Sup3r$ecret")
	})
}

func TestMariadbConn_SetClientSsl(t *testing.T) {
	for _, tc := range []struct {
		name                  string
		ssl                   bool
		password              string
		major, minor          int
		kind                  clientKind
		sslVerify, sslRequest bool
		args                  []string
	}{
		{"ZeroConfClient", true, "Sup3r$ecret", 11, 8, clientMariadb, true, false, []string{"--ssl-verify-server-cert"}},
		{"ZeroConfClient114", true, "Sup3r$ecret", 11, 4, clientMariadb, true, false, []string{"--ssl-verify-server-cert"}},
		{"ZeroConfClientNoPassword", true, "", 11, 8, clientMariadb, false, false, nil},
		{"OlderClient113", true, "Sup3r$ecret", 11, 3, clientMariadb, false, true, []string{"--ssl", "--skip-ssl-verify-server-cert"}},
		{"OlderClient1011", true, "Sup3r$ecret", 10, 11, clientMariadb, false, true, []string{"--ssl", "--skip-ssl-verify-server-cert"}},
		{"OlderServer", false, "Sup3r$ecret", 11, 8, clientMariadb, false, true, []string{"--ssl", "--skip-ssl-verify-server-cert"}},
		{"OlderServerOlderClient", false, "Sup3r$ecret", 10, 11, clientMariadb, false, true, []string{"--ssl", "--skip-ssl-verify-server-cert"}},
		{"OtherClient", true, "Sup3r$ecret", 0, 0, clientOther, false, false, nil},
		{"OtherClientOlderServer", false, "Sup3r$ecret", 0, 0, clientOther, false, false, []string{"--skip-ssl"}},
		{"UnknownClient", true, "Sup3r$ecret", 0, 0, clientUnknown, true, false, []string{"--ssl-verify-server-cert"}},
		{"UnknownClientNoPassword", true, "", 0, 0, clientUnknown, false, false, nil},
		{"UnknownClientOlderServer", false, "Sup3r$ecret", 0, 0, clientUnknown, false, false, []string{"--skip-ssl"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			conn := mariadbTestConn()
			conn.Ssl, conn.Password, conn.SslVerify = tc.ssl, tc.password, false
			conn.setClientSsl(tc.major, tc.minor, tc.kind)
			assert.Equal(t, tc.sslVerify, conn.SslVerify)
			assert.Equal(t, tc.sslRequest, conn.SslRequest)

			args := conn.Cmd().Args
			assert.Equal(t, tc.args, sslArgs(args))
		})
	}
}

// sslArgs returns the TLS flags of a client command.
func sslArgs(args []string) (result []string) {
	for _, arg := range args {
		if strings.Contains(arg, "ssl") {
			result = append(result, arg)
		}
	}

	return result
}

func TestMariadbClientVersion(t *testing.T) {
	for _, tc := range []struct {
		name, output string
		major, minor int
		kind         clientKind
	}{
		{"MariaDB118", "mariadb-dump from 11.8.6-MariaDB, client 10.19 for debian-linux-gnu (x86_64)", 11, 8, clientMariadb},
		{"MariaDB114", "mariadb  Ver 15.1 Distrib 11.4.2-MariaDB, for debian-linux-gnu (x86_64)", 11, 4, clientMariadb},
		{"MariaDB123", "mariadb from 12.3.3-MariaDB, client 15.2 for Linux (x86_64)", 12, 3, clientMariadb},
		{"MariaDB113", "mariadb  Ver 15.1 Distrib 11.3.2-MariaDB, for debian-linux-gnu (x86_64)", 11, 3, clientMariadb},
		{"MariaDB1011", "mysqldump  Ver 10.19 Distrib 10.11.6-MariaDB, for debian-linux-gnu (x86_64)", 10, 11, clientMariadb},
		{"MySQL80", "mysqldump  Ver 8.0.36 for Linux on x86_64 (MySQL Community Server - GPL)", 0, 0, clientOther},
		{"Empty", "", 0, 0, clientOther},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bin, runs := fakeClient(t, tc.output, 0)
			for i := 0; i < 2; i++ {
				major, minor, kind := mariadbClientVersion(bin)
				assert.Equal(t, tc.major, major)
				assert.Equal(t, tc.minor, minor)
				assert.Equal(t, tc.kind, kind)
			}
			assert.Equal(t, []string{"--no-defaults --version"}, runArgs(t, runs), "expected one cached check")
		})
	}
	t.Run("MissingNotCached", func(t *testing.T) {
		bin := filepath.Join(t.TempDir(), "missing")
		_, _, kind := mariadbClientVersion(bin)
		assert.Equal(t, clientUnknown, kind)
		_, cached := clientVersions.Load(bin)
		assert.False(t, cached)
	})
	t.Run("FailedNotCached", func(t *testing.T) {
		// A check that fails is not taken for a client that is not a MariaDB client.
		bin, runs := fakeClient(t, "from 11.8.6-MariaDB", 1)
		_, _, kind := mariadbClientVersion(bin)
		assert.Equal(t, clientUnknown, kind)
		_, _, kind = mariadbClientVersion(bin)
		assert.Equal(t, clientUnknown, kind)
		assert.Len(t, runArgs(t, runs), 2)
	})
	t.Run("TimeoutNotCached", func(t *testing.T) {
		prevTimeout, prevDelay := clientProbeTimeout, clientProbeWaitDelay
		t.Cleanup(func() { clientProbeTimeout, clientProbeWaitDelay = prevTimeout, prevDelay })
		clientProbeTimeout, clientProbeWaitDelay = 200*time.Millisecond, 200*time.Millisecond

		bin := filepath.Join(t.TempDir(), "client")
		require.NoError(t, os.WriteFile(bin, []byte("#!/bin/sh\necho 'from 11.8.6-MariaDB'\nexec sleep 30\n"), 0o700))

		_, _, kind := mariadbClientVersion(bin)
		assert.Equal(t, clientUnknown, kind)
		_, cached := clientVersions.Load(bin)
		assert.False(t, cached)
	})
}

// TestNewMariadbConn_OlderServer verifies that a MariaDB client still requests TLS from a server that is known
// not to offer zero-configuration TLS.
func TestNewMariadbConn_OlderServer(t *testing.T) {
	c := get.Config()

	if c.DatabaseSsl() || c.DatabaseVersion() == "" {
		t.Skip("requires a test database whose server is known not to offer zero-configuration ssl")
	}

	bin, _ := fakeClient(t, "mariadb-dump from 11.8.6-MariaDB, client 10.19 for debian-linux-gnu (x86_64)", 0)
	conn := newMariadbConn(c, bin)

	if conn.Socket != "" {
		t.Skip("requires a tcp connection")
	}

	assert.False(t, conn.Ssl)
	assert.True(t, conn.SslRequest)
	assert.NotContains(t, conn.Cmd().Args, "--skip-ssl")
}

// TestNewMariadbConn_Tls verifies how the client version and an unknown server version decide the TLS flags.
func TestNewMariadbConn_Tls(t *testing.T) {
	// The configuration is not connected to a database, so the server version is unknown.
	c := config.NewMinimalTestConfig(t.TempDir())
	require.Empty(t, c.DatabaseVersion())
	c.Options().DatabaseDriver = dsn.DriverMySQL
	c.Options().DatabaseDSN = ""
	c.Options().DatabaseServer = "mariadb:4001"
	c.Options().DatabasePassword = "Sup3r$ecret"

	for _, tc := range []struct {
		name, output          string
		sslVerify, sslRequest bool
	}{
		{"MariaDB118", "mariadb-dump from 11.8.6-MariaDB, client 10.19 for debian-linux-gnu (x86_64)", true, false},
		{"MariaDB1011", "mysqldump  Ver 10.19 Distrib 10.11.6-MariaDB, for debian-linux-gnu (x86_64)", false, true},
		{"MySQL80", "mysqldump  Ver 8.0.36 for Linux on x86_64 (MySQL Community Server - GPL)", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bin, _ := fakeClient(t, tc.output, 0)
			conn := newMariadbConn(c, bin)
			assert.True(t, conn.Ssl, "TLS must not be skipped for an unknown server version")
			assert.True(t, conn.SslUnknown)
			assert.Equal(t, tc.sslVerify, conn.SslVerify)
			assert.Equal(t, tc.sslRequest, conn.SslRequest)
			assert.NotContains(t, conn.Cmd().Args, "--skip-ssl")
		})
	}
	t.Run("NoPassword", func(t *testing.T) {
		bin, _ := fakeClient(t, "mariadb-dump from 11.8.6-MariaDB, client 10.19 for debian-linux-gnu (x86_64)", 0)
		c.Options().DatabasePassword = ""
		t.Cleanup(func() { c.Options().DatabasePassword = "Sup3r$ecret" })
		conn := newMariadbConn(c, bin)
		assert.False(t, conn.SslVerify)
		assert.False(t, conn.SslRequest)
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
