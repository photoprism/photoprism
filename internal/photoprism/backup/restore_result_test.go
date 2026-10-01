package backup

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"testing/iotest"
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
		assert.Equal(t, "restore: index database restored, but 7 statements failed, so some rows may be missing "+
			"(error 1062 at line 4, error 1286 at line 5)", hook.LastEntry().Message)
	})
}

// TestRestoreFailures_Summary checks how failed statements are described.
func TestRestoreFailures_Summary(t *testing.T) {
	assert.Equal(t, "1 statement failed", restoreFailures{Count: 1}.Summary())
	assert.Equal(t, "3 statements failed", restoreFailures{Count: 3}.Summary())
}

// TestPreparedRestore_RunLog checks the outcome a restore logs.
func TestPreparedRestore_RunLog(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		hook := captureLog(t)
		require.NoError(t, preparedRestore{cmd: exec.Command("sh", "-c", "cat >/dev/null")}.run(strings.NewReader("SELECT 1;\n")))
		assert.Contains(t, logMessages(hook), "restore: index database successfully restored")
	})
	t.Run("FailedStatements", func(t *testing.T) {
		hook := captureLog(t)
		script := `printf '%s\n' "ERROR 1062 (23000) at line 4: Duplicate entry 'val-a' for key 'PRIMARY'" >&2; cat >/dev/null`
		require.NoError(t, preparedRestore{cmd: exec.Command("sh", "-c", script)}.run(strings.NewReader("")))
		assert.Contains(t, logMessages(hook), "restore: index database restored, but 1 statement failed, so some rows may be missing (error 1062 at line 4)")
		assert.NotContains(t, logMessages(hook), "restore: index database successfully restored")
		for _, entry := range hook.AllEntries() {
			if entry.Level != logrus.TraceLevel {
				assert.NotContains(t, entry.Message, "val-a")
			}
		}
	})
	t.Run("WarningsBeforeOutputOnly", func(t *testing.T) {
		hook := captureLog(t)
		script := `printf '%s\n' "WARNING: insecure" "ERROR 1062 (23000) at line 4: Duplicate entry 'x" "WARNING: val-a' for key 'v'" >&2; cat >/dev/null`
		require.NoError(t, preparedRestore{cmd: exec.Command("sh", "-c", script)}.run(strings.NewReader("")))
		assert.Contains(t, logMessages(hook), "restore: insecure")
		for _, entry := range hook.AllEntries() {
			if entry.Level != logrus.TraceLevel {
				assert.NotContains(t, entry.Message, "val-a")
			}
		}
	})
	t.Run("FailedWithStatements", func(t *testing.T) {
		script := `printf '%s\n' "ERROR 1062 (23000) at line 4: Duplicate entry" "ERROR 2013 (HY000): Lost connection" >&2; cat >/dev/null; exit 1`
		err := preparedRestore{cmd: exec.Command("sh", "-c", script)}.run(strings.NewReader(""))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ERROR 2013 (HY000): Lost connection; 1 statement failed (error 1062 at line 4)")
	})
	t.Run("Failed", func(t *testing.T) {
		hook := captureLog(t)
		require.Error(t, preparedRestore{cmd: exec.Command("sh", "-c", "echo 'ERROR 2002 (HY000): Can not connect' >&2; exit 1")}.run(strings.NewReader("")))
		assert.Contains(t, logMessages(hook), "restore: failed to restore index database")
	})
}

// TestRunRestore_Sqlite restores dumps with the SQLite client into a temporary database file.
func TestRunRestore_Sqlite(t *testing.T) {
	bin, err := exec.LookPath("sqlite3")

	if err != nil {
		t.Skip("sqlite3 client not found")
	}

	restore := func(t *testing.T, dump string) (restoreFailures, error) {
		t.Helper()
		return runRestore(sqliteRestoreCmd(bin, filepath.Join(t.TempDir(), "index.db")), strings.NewReader(dump), "")
	}

	t.Run("Clean", func(t *testing.T) {
		failed, restoreErr := restore(t, "CREATE TABLE t (id INTEGER PRIMARY KEY);\nINSERT INTO t VALUES (1);\n")
		require.NoError(t, restoreErr)
		assert.Zero(t, failed.Count)
	})
	t.Run("FailedStatement", func(t *testing.T) {
		// The SQLite client exits with an error after any failed statement, which fails the restore.
		_, restoreErr := restore(t, "CREATE TABLE t (id INTEGER PRIMARY KEY, v TEXT);\nINSERT INTO t VALUES (1,'val-a');\n"+
			"INSERT INTO t VALUES (1,'val-b');\nINSERT INTO t VALUES (3,'val-d');\n")
		require.Error(t, restoreErr)
		assert.Contains(t, restoreErr.Error(), "UNIQUE constraint failed")
	})
	t.Run("Incomplete", func(t *testing.T) {
		_, restoreErr := restore(t, "PRAGMA foreign_keys=OFF;\nBEGIN TRANSACTION;\nCREATE TABLE t (id INTEGER);\n"+
			"INSERT INTO t VALUES (1);\nINSERT INTO t VALUES (2")
		require.Error(t, restoreErr)
		assert.Contains(t, restoreErr.Error(), "incomplete input")
	})
	t.Run("CannotOpen", func(t *testing.T) {
		cmd := sqliteRestoreCmd(bin, filepath.Join(t.TempDir(), "missing", "index.db"))
		_, restoreErr := runRestore(cmd, strings.NewReader("CREATE TABLE t (id INTEGER);\n"), "")
		require.Error(t, restoreErr)
		assert.Contains(t, restoreErr.Error(), "unable to open database")
	})
	t.Run("ReadOnly", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("root can write read-only files")
		}
		dbFile := filepath.Join(t.TempDir(), "index.db")
		_, restoreErr := runRestore(sqliteRestoreCmd(bin, dbFile), strings.NewReader("CREATE TABLE t (id INTEGER);\n"), "")
		require.NoError(t, restoreErr)
		require.NoError(t, os.Chmod(dbFile, 0o400))
		_, restoreErr = runRestore(sqliteRestoreCmd(bin, dbFile), strings.NewReader("INSERT INTO t VALUES (1);\nINSERT INTO t VALUES (2);\n"), "")
		require.Error(t, restoreErr)
		assert.Contains(t, restoreErr.Error(), "(8)")
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
	t.Run("DumpHeader", func(t *testing.T) {
		// A dump turns off unique and foreign key checks and inserts each table in one transaction, so the
		// restore keeps the rows of a table only if its input keeps unique checks on.
		dump := dumpHeader(string(uniqueChecksOff), "\n") +
			"CREATE TABLE t (id INT PRIMARY KEY, v VARCHAR(20), UNIQUE KEY (v)) ENGINE=InnoDB;\n" +
			"SET @OLD_AUTOCOMMIT=@@AUTOCOMMIT, @@AUTOCOMMIT=0;\n" +
			"/*!40000 ALTER TABLE `t` DISABLE KEYS */;\n" +
			"INSERT INTO t VALUES\n(1,'val-a'),\n(2,'val-b'),\n(3,'val-c');\n" +
			"INSERT INTO t VALUES\n(1,'val-a'),\n(2,'val-b'),\n(3,'val-c');\n" +
			"/*!40000 ALTER TABLE `t` ENABLE KEYS */;\n" +
			"COMMIT;\nSET AUTOCOMMIT=@OLD_AUTOCOMMIT;\n" +
			"/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;\n" +
			"/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;\n"
		admin(t, "DROP DATABASE IF EXISTS "+name+"; CREATE DATABASE "+name)
		target := conn
		target.Name = name
		failed, restoreErr := runRestore(target.Cmd(mariadbRestoreArgs(bin)...), restoreReader(dsn.DriverMariaDB, strings.NewReader(dump)), conn.Password)
		require.NoError(t, restoreErr)
		assert.Equal(t, 1, failed.Count)
		assert.Equal(t, []string{"error 1062 at line 14"}, failed.Errors)
		out, cmdErr := conn.Cmd("-N", "-e", "SELECT COUNT(*) FROM "+name+".t").CombinedOutput()
		require.NoError(t, cmdErr, string(out))
		assert.Equal(t, "3", strings.TrimSpace(string(out)))
	})
	t.Run("LostConnection", func(t *testing.T) {
		// Statements after a lost connection are reported as failed, although the client exits 0.
		failed := restore(t, "CREATE TABLE t (id INT);\nINSERT INTO t VALUES (1);\nKILL CONNECTION_ID();\n"+
			"INSERT INTO t VALUES (2);\nINSERT INTO t VALUES (3);\n")
		assert.Equal(t, 3, failed.Count)
		assert.Equal(t, []string{"error 1927 at line 3", "error 2013 at line 4", "error 2006 at line 5"}, failed.Errors)
	})
	t.Run("MultiLineSyntaxError", func(t *testing.T) {
		// A completed restore logs no part of a statement that failed, including a quoted value with line breaks.
		dump := "CREATE TABLE t (id INT PRIMARY KEY, v TEXT);\n" +
			"INSERT INTO t VALUES (1 x,'val-a\nWARNING: val-b');\nINSERT INTO t VALUES (2,'c');\n"
		hook := captureLog(t)
		failed := restore(t, dump)
		assert.Equal(t, 1, failed.Count)
		for _, entry := range hook.AllEntries() {
			if entry.Level != logrus.TraceLevel {
				assert.NotContains(t, entry.Message, "val-")
			}
		}
	})
	t.Run("UnknownCommand", func(t *testing.T) {
		// The sandbox refuses client commands, and the statement that follows fails.
		failed := restore(t, "CREATE TABLE t (id INT PRIMARY KEY);\n\\! echo val-a\nINSERT INTO t VALUES (1);\n"+
			"INSERT INTO t VALUES (2);\n")
		assert.Equal(t, 2, failed.Count)
		assert.Equal(t, []string{"error at line 2", "error 1064 at line 2"}, failed.Errors)
	})
	t.Run("LargeFailedStatement", func(t *testing.T) {
		failed := restore(t, "CREATE TABLE t (id INT PRIMARY KEY, v LONGTEXT);\n"+
			"INSERT INTO t VALUES (1,'a');\n"+
			"INSERT INTO t VALUES (2,'"+strings.Repeat("x", 4<<20)+"'),(1,'b');\n")
		assert.Equal(t, 1, failed.Count)
		assert.Equal(t, []string{"error 1062 at line 3"}, failed.Errors)
	})
}

// TestRestoreInput checks that only errors reading the backup are recorded.
func TestRestoreInput(t *testing.T) {
	t.Run("ReadError", func(t *testing.T) {
		in := &restoreInput{r: io.MultiReader(strings.NewReader("x"), iotest.ErrReader(errors.New("read failed")))}
		_, err := io.Copy(io.Discard, in)
		require.Error(t, err)
		assert.EqualError(t, in.err, "read failed")
	})
	t.Run("EndOfInput", func(t *testing.T) {
		in := &restoreInput{r: strings.NewReader("SELECT 1;\n")}
		_, err := io.Copy(io.Discard, in)
		require.NoError(t, err)
		assert.NoError(t, in.err)
	})
}
