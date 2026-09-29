package backup

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/dsn"
)

// TestLogRestoreResult checks the log line of a completed restore.
func TestLogRestoreResult(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		hook := captureLog(t)
		logRestoreResult(restoreFailures{})
		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, logrus.InfoLevel, hook.LastEntry().Level)
		assert.Equal(t, "restore: index database successfully restored", hook.LastEntry().Message)
	})
	t.Run("FailedStatements", func(t *testing.T) {
		hook := captureLog(t)
		logRestoreResult(restoreFailures{Count: 7, Errors: []string{"error 1062 at line 4", "error 1286 at line 5"}})
		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, logrus.WarnLevel, hook.LastEntry().Level)
		assert.Equal(t, "restore: index database restored, but 7 statements failed and some rows may be missing "+
			"(error 1062 at line 4, error 1286 at line 5)", hook.LastEntry().Message)
	})
}

// TestRunRestore_MariaDB restores dumps with the MariaDB client into a temporary database.
func TestRunRestore_MariaDB(t *testing.T) {
	if os.Getenv("PHOTOPRISM_TEST_DRIVER") != dsn.DriverMySQL {
		t.Skip("requires the MariaDB test database")
	}

	bin, err := exec.LookPath("mariadb")

	if err != nil {
		t.Skip("mariadb client not found")
	}

	d := dsn.Parse(os.Getenv("PHOTOPRISM_TEST_DSN"))
	conn := mariadbConn{Bin: bin, Host: d.Host(), Port: strconv.Itoa(d.Port()), User: d.User, Password: d.Password, Name: "mysql"}
	name := fmt.Sprintf("zz_restore_%d", time.Now().UnixNano())

	admin := func(t *testing.T, sql string) {
		t.Helper()
		out, cmdErr := conn.Cmd("-e", sql).CombinedOutput()
		require.NoError(t, cmdErr, string(out))
	}

	restore := func(t *testing.T, dump string) restoreFailures {
		t.Helper()
		admin(t, "DROP DATABASE IF EXISTS "+name+"; CREATE DATABASE "+name)
		target := conn
		target.Name = name
		failed, restoreErr := runRestore(target.Cmd(mariadbRestoreArgs(bin)...), strings.NewReader(dump), conn.Password)
		require.NoError(t, restoreErr)
		return failed
	}

	t.Cleanup(func() { _ = conn.Cmd("-e", "DROP DATABASE IF EXISTS "+name).Run() })

	t.Run("Clean", func(t *testing.T) {
		failed := restore(t, "CREATE TABLE t (id INT PRIMARY KEY);\nINSERT INTO t VALUES (1),(2);\n")
		assert.Zero(t, failed.Count)
	})
	t.Run("FailedStatements", func(t *testing.T) {
		failed := restore(t, "CREATE TABLE t (id INT PRIMARY KEY, v VARCHAR(20));\n"+
			"INSERT INTO t VALUES (1,'val-a'),(2,'val-b');\n"+
			"INSERT INTO t VALUES (3,'val-c'),(1,'val-d'),(4,'val-e');\n"+
			"CREATE TABLE u (id INT) ENGINE=NoSuchEngine;\n"+
			"INSERT INTO t VALUES (5,'val-f');\n")
		assert.Equal(t, 2, failed.Count)
		assert.Equal(t, []string{"error 1062 at line 3", "error 1286 at line 4"}, failed.Errors)
	})
	t.Run("LostConnection", func(t *testing.T) {
		// Statements after a lost connection are reported as failed, although the client exits 0.
		failed := restore(t, "CREATE TABLE t (id INT);\nINSERT INTO t VALUES (1);\nKILL CONNECTION_ID();\n"+
			"INSERT INTO t VALUES (2);\nINSERT INTO t VALUES (3);\n")
		assert.Equal(t, 3, failed.Count)
		assert.Equal(t, []string{"error 1927 at line 3", "error 2013 at line 4", "error 2006 at line 5"}, failed.Errors)
	})
	t.Run("LargeFailedStatement", func(t *testing.T) {
		failed := restore(t, "CREATE TABLE t (id INT PRIMARY KEY, v LONGTEXT);\n"+
			"INSERT INTO t VALUES (1,'a');\n"+
			"INSERT INTO t VALUES (2,'"+strings.Repeat("x", 4<<20)+"'),(1,'b');\n")
		assert.Equal(t, 1, failed.Count)
		assert.Equal(t, []string{"error 1062 at line 3"}, failed.Errors)
	})
}
