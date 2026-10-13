package backup

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// newPostgresTestConfig returns a config for a PostgreSQL database that is not connected.
func newPostgresTestConfig(t *testing.T) *config.Config {
	t.Helper()

	c := config.NewMinimalTestConfig(t.TempDir())
	c.Options().DatabaseDriver = dsn.DriverPostgres
	c.Options().DatabaseDSN = ""

	return c
}

func TestNewPostgresConn(t *testing.T) {
	t.Run("Server", func(t *testing.T) {
		c := newPostgresTestConfig(t)
		c.Options().DatabaseServer = "db.local:5433"
		c.Options().DatabaseName = "photos"
		c.Options().DatabaseUser = "app"
		c.Options().DatabasePassword = "secret"

		conn := newPostgresConn(c, "/usr/bin/pg_dump")
		assert.Equal(t, "/usr/bin/pg_dump", conn.Bin)
		assert.Equal(t, "db.local", conn.Host)
		assert.Equal(t, "5433", conn.Port)
		assert.Equal(t, "app", conn.User)
		assert.Equal(t, "photos", conn.Name)
		assert.Equal(t, "secret", conn.Password)
		assert.Equal(t, map[string]string{"sslmode": "disable", "connect_timeout": "15"}, conn.Params)
		assert.Equal(t, "app@db.local:5433/photos", conn.String())
	})
	t.Run("ServerWithoutPort", func(t *testing.T) {
		c := newPostgresTestConfig(t)
		c.Options().DatabaseServer = "db.local"

		// The DSN of the app names the default port.
		conn := newPostgresConn(c, "/usr/bin/pg_dump")
		assert.Equal(t, "db.local", conn.Host)
		assert.Equal(t, "5432", conn.Port)
	})
	t.Run("DSNWithoutUserAndName", func(t *testing.T) {
		c := newPostgresTestConfig(t)
		c.Options().DatabaseDSN = "postgres://db.local:5433"

		// The user, database and password are read from the environment, as the app does.
		conn := newPostgresConn(c, "/usr/bin/pg_dump")
		assert.Empty(t, conn.User)
		assert.Empty(t, conn.Name)
		assert.Empty(t, conn.Password)
	})
	t.Run("DSNWithoutPort", func(t *testing.T) {
		c := newPostgresTestConfig(t)
		c.Options().DatabaseDSN = "postgres://app:secret@db.local/photos"

		// The port is read from the environment, as the app does.
		conn := newPostgresConn(c, "/usr/bin/pg_dump")
		assert.Equal(t, "db.local", conn.Host)
		assert.Empty(t, conn.Port)
		assert.Equal(t, "app@db.local/photos", conn.String())
	})
	t.Run("SocketServer", func(t *testing.T) {
		c := newPostgresTestConfig(t)
		c.Options().DatabaseServer = "/var/run/postgresql:5433"

		conn := newPostgresConn(c, "/usr/bin/psql")
		assert.Equal(t, "/var/run/postgresql", conn.Host)
		assert.Equal(t, "5433", conn.Port)
	})
	t.Run("SocketURI", func(t *testing.T) {
		c := newPostgresTestConfig(t)
		c.Options().DatabaseDSN = "postgres://app@/photos?host=%2Fvar%2Frun%2Fpostgresql&port=5433&sslmode=require"

		conn := newPostgresConn(c, "/usr/bin/psql")
		assert.Equal(t, "/var/run/postgresql", conn.Host)
		assert.Equal(t, "5433", conn.Port)
		assert.Equal(t, "app", conn.User)
		assert.Equal(t, "photos", conn.Name)
		assert.Equal(t, map[string]string{"sslmode": "require"}, conn.Params)
	})
	t.Run("HostAndPortParams", func(t *testing.T) {
		c := newPostgresTestConfig(t)
		c.Options().DatabaseDSN = "postgres://app:secret@/photos?host=db.internal&port=5433"

		conn := newPostgresConn(c, "/usr/bin/psql")
		assert.Equal(t, "db.internal", conn.Host)
		assert.Equal(t, "5433", conn.Port)
	})
	t.Run("KeyValue", func(t *testing.T) {
		c := newPostgresTestConfig(t)
		c.Options().DatabaseDSN = `host=db.local user=app dbname=photos sslmode=verify-full sslrootcert='/etc/my ca.pem'`

		conn := newPostgresConn(c, "/usr/bin/psql")
		assert.Equal(t, "db.local", conn.Host)
		assert.Equal(t, "app", conn.User)
		assert.Equal(t, "photos", conn.Name)
		assert.Equal(t, map[string]string{"sslmode": "verify-full", "sslrootcert": "/etc/my ca.pem"}, conn.Params)
	})
	t.Run("SystemCertificates", func(t *testing.T) {
		t.Setenv("PGSSLROOTCERT", "")
		c := newPostgresTestConfig(t)
		c.Options().DatabaseDSN = "postgres://app:secret@db.local/photos?sslmode=verify-full"

		conn := newPostgresConn(c, "/usr/bin/psql")
		assert.Equal(t, "system", conn.Params["sslrootcert"])

		// Certificates named in the environment are used, as the app does.
		t.Setenv("PGSSLROOTCERT", "/private-ca.pem")
		conn = newPostgresConn(c, "/usr/bin/psql")
		assert.Empty(t, conn.Params["sslrootcert"])
	})
	t.Run("VerifyCA", func(t *testing.T) {
		t.Setenv("PGSSLROOTCERT", "")
		c := newPostgresTestConfig(t)
		c.Options().DatabaseDSN = "postgres://app:secret@db.local/photos?sslmode=verify-ca"

		// A client refuses the system certificates for this mode.
		conn := newPostgresConn(c, "/usr/bin/psql")
		assert.Empty(t, conn.Params["sslrootcert"])
	})
}

func TestPostgresParams(t *testing.T) {
	t.Run("Query", func(t *testing.T) {
		assert.Equal(t, map[string]string{"sslmode": "verify-full", "sslrootcert": "/a+b.pem", "channel_binding": "require"},
			postgresParams("sslmode=verify-full&sslrootcert=%2Fa%2Bb.pem&TimeZone=UTC&channel_binding=require", true))
	})
	t.Run("PassFile", func(t *testing.T) {
		assert.Equal(t, map[string]string{"passfile": "/run/pgpass"}, postgresParams("passfile=%2Frun%2Fpgpass", true))
	})
	t.Run("KeyValue", func(t *testing.T) {
		assert.Equal(t, map[string]string{"sslmode": "require", "connect_timeout": "5", "sslrootcert": "/a+b c.pem"},
			postgresParams(`sslmode='require' TimeZone=UTC connect_timeout=5 sslrootcert="/a+b c.pem"`, false))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Empty(t, postgresParams("", true))
		assert.Empty(t, postgresParams("", false))
	})
}

func TestSplitPostgresParams(t *testing.T) {
	assert.Equal(t, []string{"a=1", "b=x y", "c=it's"}, splitPostgresParams(`a=1  b="x y"	c="it's"`))
	assert.Empty(t, splitPostgresParams(" "))
}

func TestPostgresConn_Cmd(t *testing.T) {
	t.Setenv("PGPASSWORD", "inherited")
	t.Setenv("PGSERVICE", "inherited")
	t.Setenv("PGHOST", "inherited")
	t.Setenv("PGPORT", "4444")
	t.Setenv("PGSSLMODE", "verify-full")

	conn := postgresConn{Bin: "/usr/bin/psql", Host: "db.local", Port: "5432", User: "app", Name: "photos",
		Password: "secret", Params: map[string]string{"sslmode": "require"}}
	cmd := conn.Cmd(postgresRestoreArgs()...)

	assert.Equal(t, append([]string{"/usr/bin/psql"}, postgresRestoreArgs()...), cmd.Args)
	assert.NotContains(t, strings.Join(cmd.Args, " "), "secret")
	assert.True(t, slices.Contains(cmd.Env, "PGHOST=db.local"), "PGHOST=db.local")
	assert.True(t, slices.Contains(cmd.Env, "PGPORT=5432"), "PGPORT=5432")
	assert.True(t, slices.Contains(cmd.Env, "PGUSER=app"), "PGUSER=app")
	assert.True(t, slices.Contains(cmd.Env, "PGDATABASE=photos"), "PGDATABASE=photos")
	assert.True(t, slices.Contains(cmd.Env, "PGPASSWORD=secret"), "PGPASSWORD=secret")
	assert.False(t, slices.Contains(cmd.Env, "PGPASSWORD=inherited"), "PGPASSWORD=inherited")
	assert.False(t, slices.Contains(cmd.Env, "PGSERVICE=inherited"), "PGSERVICE=inherited")
	assert.False(t, slices.Contains(cmd.Env, "PGHOST=inherited"), "PGHOST=inherited")
	assert.False(t, slices.Contains(cmd.Env, "PGPORT=4444"), "PGPORT=4444")

	// The parameters of the connection come after the inherited ones, which take effect otherwise.
	assert.Greater(t, slices.Index(cmd.Env, "PGSSLMODE=require"), slices.Index(cmd.Env, "PGSSLMODE=verify-full"))

	t.Run("Inherited", func(t *testing.T) {
		conn.Host, conn.Password, conn.Port = "", "", ""
		env := conn.Cmd().Env

		assert.True(t, slices.Contains(env, "PGHOST=inherited"), "PGHOST=inherited")
		assert.True(t, slices.Contains(env, "PGPASSWORD=inherited"), "PGPASSWORD=inherited")
		assert.True(t, slices.Contains(env, "PGPORT=4444"), "PGPORT=4444")
		assert.False(t, slices.Contains(env, "PGSERVICE=inherited"), "PGSERVICE=inherited")
	})
}

func TestPostgresArgs(t *testing.T) {
	assert.Equal(t, []string{"--no-password", "--format=plain", "--clean", "--if-exists", "--no-owner", "--no-privileges"}, postgresDumpArgs())
	assert.Equal(t, []string{"--no-password", "--no-psqlrc", "--quiet", "--set=ON_ERROR_STOP=1", "--set=VERBOSITY=terse",
		"--set=SHOW_CONTEXT=never", "--single-transaction", "--output=" + os.DevNull, "--file=-"}, postgresRestoreArgs())
}

func TestPostgresRestrictSupported(t *testing.T) {
	// fakeClient returns a client that prints the specified version.
	fakeClient := func(t *testing.T, version string) string {
		bin := filepath.Join(t.TempDir(), "psql")
		require.NoError(t, os.WriteFile(bin, []byte("#!/bin/sh\necho 'psql (PostgreSQL) "+version+"'\n"), 0o700)) //nolint:gosec // test client
		return bin
	}

	for version, want := range map[string]bool{"18.0": true, "17.6": true, "17.5": false, "16.10": true, "16.9": false,
		"15.14": true, "14.19": true, "13.22": true, "13.21": false, "12.22": false, "19.1": true} {
		t.Run(version, func(t *testing.T) {
			ok, err := postgresRestrictSupported(fakeClient(t, version))
			require.NoError(t, err)
			assert.Equal(t, want, ok)
		})
	}
	t.Run("NoVersion", func(t *testing.T) {
		bin := filepath.Join(t.TempDir(), "psql")
		require.NoError(t, os.WriteFile(bin, []byte("#!/bin/sh\necho other\n"), 0o700)) //nolint:gosec // test client
		ok, err := postgresRestrictSupported(bin)
		require.NoError(t, err)
		assert.False(t, ok)
	})
	t.Run("NotFound", func(t *testing.T) {
		_, err := postgresRestrictSupported(filepath.Join(t.TempDir(), "missing"))
		assert.Error(t, err)
	})
}

func TestPostgresRestrictedReader(t *testing.T) {
	// read returns the input psql receives for a dump, without the line that restricts it.
	read := func(t *testing.T, dump string) string {
		b, err := io.ReadAll(newPostgresRestrictedReader(strings.NewReader(dump)))
		require.NoError(t, err)
		first, rest, _ := strings.Cut(string(b), "\n")
		require.Regexp(t, `^\\restrict [0-9A-Za-z]{32}$`, first)
		return rest
	}

	t.Run("Dump", func(t *testing.T) {
		assert.Equal(t, "--\n\nCREATE TABLE t (v text);\nCOPY public.t (v) FROM stdin;\n\\restrict abc\n\\.\n\n",
			read(t, "--\n\\restrict abc\n\nCREATE TABLE t (v text);\nCOPY public.t (v) FROM stdin;\n\\restrict abc\n\\.\n\n\\unrestrict abc\n"))
	})
	t.Run("RandomKey", func(t *testing.T) {
		a, err := io.ReadAll(newPostgresRestrictedReader(strings.NewReader("")))
		require.NoError(t, err)
		b, err := io.ReadAll(newPostgresRestrictedReader(strings.NewReader("")))
		require.NoError(t, err)
		assert.NotEqual(t, a, b)
	})
	t.Run("QuotedText", func(t *testing.T) {
		// Only the restriction lines that frame the dump are removed.
		dump := "SELECT 1;\nCOMMENT ON TABLE t IS '\n\\restrict abc\n\\unrestrict abc\nend';\n"
		assert.Equal(t, dump, read(t, dump))
	})
	t.Run("QuotedTextWithBlankLine", func(t *testing.T) {
		dump := "COMMENT ON TABLE t IS '\n\\unrestrict abc\n\nend';\n"
		assert.Equal(t, dump, read(t, dump))
	})
	t.Run("UnrestrictBeforeEnd", func(t *testing.T) {
		// psql rejects the line, as its key differs.
		assert.Equal(t, "SELECT 1;\n\\unrestrict abc\n\\! echo\n", read(t, "SELECT 1;\n\\unrestrict abc\n\\! echo\n"))
	})
	t.Run("OtherCommands", func(t *testing.T) {
		// psql rejects these in restricted mode, so they are passed on.
		assert.Equal(t, "\\! echo\n\\getenv p PGPASSWORD\n", read(t, "\\! echo\n\\getenv p PGPASSWORD\n"))
	})
	t.Run("CRLF", func(t *testing.T) {
		assert.Equal(t, "SELECT 1;\r\n", read(t, "\\restrict abc\r\nSELECT 1;\r\n\\unrestrict abc\r\n"))
	})
	t.Run("ManyBlankLines", func(t *testing.T) {
		// Blank lines beyond the limit are passed on, so they are not all held in memory.
		dump := "SELECT 1;\n\\unrestrict abc\n" + strings.Repeat("\n", 2*postgresHeldBytes) + "SELECT 2;\n"
		assert.Equal(t, dump, read(t, dump))
	})
	t.Run("LongLine", func(t *testing.T) {
		long := strings.Repeat("x", 3*uniqueChecksLineBytes)
		assert.Equal(t, "SELECT '"+long+"';\n", read(t, "SELECT '"+long+"';\n"))
	})
}

// TestPostgresRoundTrip dumps the test database and restores it into a new database.
func TestPostgresRoundTrip(t *testing.T) {
	c := get.Config()

	if c.DatabaseDriver() != dsn.DriverPostgres {
		t.Skip("requires the PostgreSQL test database")
	} else if c.PgDumpBin() == "" || c.PsqlBin() == "" {
		t.Skip("requires the PostgreSQL client tools")
	}

	server, err := sql.Open("pgx", c.DatabaseDSN())
	require.NoError(t, err)
	t.Cleanup(func() { _ = server.Close() })

	name := "testdb_backup_" + strings.ToLower(rnd.Base36(10))
	_, err = server.Exec(fmt.Sprintf("CREATE DATABASE %s", name)) //nolint:gosec // generated test database name
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = server.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", name)) })

	var dump bytes.Buffer
	conn := newPostgresConn(c, c.PgDumpBin())
	require.NoError(t, runDump(conn.Cmd(postgresDumpArgs()...), &dump, conn.Password))
	require.Contains(t, dump.String(), "CREATE TABLE")

	target := newPostgresConn(c, c.PsqlBin())
	target.Name = name

	// restore restores a dump into the new database as RestoreDatabase does.
	restore := func(dump io.Reader) error {
		_, restoreErr := runRestore(target.Cmd(postgresRestoreArgs()...), restoreReader(dsn.DriverPostgres, dump), target.Password)
		return restoreErr
	}

	// The dump is restored twice, as it replaces the objects that exist.
	require.NoError(t, restore(bytes.NewReader(dump.Bytes())))
	require.NoError(t, restore(bytes.NewReader(dump.Bytes())))

	restored, err := sql.Open("pgx", strings.Replace(c.DatabaseDSN(), "/"+c.DatabaseName()+"?", "/"+name+"?", 1))
	require.NoError(t, err)
	t.Cleanup(func() { _ = restored.Close() })

	var current string
	require.NoError(t, restored.QueryRow("SELECT current_database()").Scan(&current))
	require.Equal(t, name, current)

	var want, got int
	require.NoError(t, server.QueryRow("SELECT COUNT(*) FROM photos").Scan(&want))
	require.NoError(t, restored.QueryRow("SELECT COUNT(*) FROM photos").Scan(&got))
	assert.Equal(t, want, got)

	// The next ID follows the restored rows.
	var maxID, nextID int64
	require.NoError(t, restored.QueryRow("SELECT COALESCE(MAX(id), 0) FROM photos").Scan(&maxID))
	require.NoError(t, restored.QueryRow("SELECT nextval(pg_get_serial_sequence('photos', 'id'))").Scan(&nextID))
	assert.Greater(t, nextID, maxID)

	t.Run("StopsAtFirstError", func(t *testing.T) {
		require.Error(t, restore(strings.NewReader("DELETE FROM photos;\nSELECT no_such_column FROM photos;\n")))

		var count int
		require.NoError(t, restored.QueryRow("SELECT COUNT(*) FROM photos").Scan(&count))
		assert.Equal(t, want, count, "a failed restore changes nothing")
	})
	t.Run("RejectsClientCommands", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "executed")

		for _, script := range []string{"\\! touch " + file + "\n", "\\getenv p PGPASSWORD\n\\echo :p\n",
			"\\unrestrict abc\n\\! touch " + file + "\n", "\\copy photos (id) from program 'touch " + file + "'\n"} {
			err := restore(strings.NewReader(script))
			require.Error(t, err, script)
			assert.NotContains(t, err.Error(), target.Password)
			assert.NoFileExists(t, file, script)
		}
	})
	t.Run("ErrorsWithoutValues", func(t *testing.T) {
		err := restore(strings.NewReader("INSERT INTO no_such_table VALUES ('secret value');\n"))
		require.Error(t, err)
		assert.NotContains(t, err.Error(), "secret value")
	})
}

func TestWarnPostgresServerPrivileges(t *testing.T) {
	t.Run("NotConnected", func(t *testing.T) {
		hook := captureLog(t)
		warnPostgresServerPrivileges(nil)
		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, logrus.DebugLevel, hook.LastEntry().Level)
		assert.Contains(t, hook.LastEntry().Message, "as the database is not connected")
	})
	t.Run("QueryFailed", func(t *testing.T) {
		// A database without pg_roles cannot be checked.
		db, err := entity.OpenDb(dsn.DriverSQLite3, ":memory:")
		require.NoError(t, err)
		t.Cleanup(func() { _ = entity.CloseDb(db) })
		hook := captureLog(t)
		warnPostgresServerPrivileges(db)
		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, logrus.WarnLevel, hook.LastEntry().Level)
		assert.Contains(t, hook.LastEntry().Message, "restore: failed to check the role of the database user")
	})
	t.Run("PostgreSQL", func(t *testing.T) {
		c := get.Config()

		if c.DatabaseDriver() != dsn.DriverPostgres {
			t.Skip("requires the PostgreSQL test database")
		}

		admin := c.Db()
		var superuser bool
		require.NoError(t, admin.Raw("SELECT rolsuper FROM pg_roles WHERE rolname = current_user").Row().Scan(&superuser))

		if !superuser {
			t.Skip("requires a superuser to connect to the PostgreSQL test database")
		}

		hook := captureLog(t)
		warnPostgresServerPrivileges(admin)
		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, logrus.WarnLevel, hook.LastEntry().Level)
		assert.Contains(t, hook.LastEntry().Message, "has server-side privileges")

		// A role that only owns objects is not reported, one with access to server files is.
		role, password := "zz_restore_"+strings.ToLower(rnd.Base36(8)), rnd.Base62(16)
		require.NoError(t, admin.Exec(fmt.Sprintf("CREATE ROLE %s LOGIN PASSWORD '%s'", role, password)).Error)
		t.Cleanup(func() { _ = admin.Exec(fmt.Sprintf("DROP ROLE IF EXISTS %s", role)).Error })

		u, err := url.Parse(c.DatabaseDSN())
		require.NoError(t, err)
		u.User = url.UserPassword(role, password)

		db, err := entity.OpenDb(dsn.DriverPostgres, u.String())
		require.NoError(t, err)

		t.Cleanup(func() { _ = entity.CloseDb(db) })

		warnPostgresServerPrivileges(db)
		assert.Len(t, hook.AllEntries(), 1)

		require.NoError(t, admin.Exec(fmt.Sprintf("ALTER ROLE %s CREATEROLE", role)).Error)
		warnPostgresServerPrivileges(db)
		require.Len(t, hook.AllEntries(), 2)
		assert.Equal(t, fmt.Sprintf("restore: database user %s has server-side privileges, restoring with a role "+
			"that only owns the database is recommended", role), hook.LastEntry().Message)

		// A membership counts even if it neither inherits privileges nor allows SET ROLE.
		require.NoError(t, admin.Exec(fmt.Sprintf("ALTER ROLE %s NOCREATEROLE", role)).Error)
		warnPostgresServerPrivileges(db)
		require.Len(t, hook.AllEntries(), 2)
		require.NoError(t, admin.Exec(fmt.Sprintf("GRANT pg_read_server_files TO %s WITH INHERIT FALSE, SET FALSE", role)).Error)
		warnPostgresServerPrivileges(db)
		require.Len(t, hook.AllEntries(), 3)
		assert.Contains(t, hook.LastEntry().Message, "has server-side privileges")

		// A check that does not finish in time is reported and skipped.
		timeout := privilegeCheckTimeout
		privilegeCheckTimeout = time.Nanosecond
		t.Cleanup(func() { privilegeCheckTimeout = timeout })
		warnPostgresServerPrivileges(admin)
		assert.Equal(t, "restore: failed to check the role of the database user (context deadline exceeded)", hook.LastEntry().Message)
	})
}
