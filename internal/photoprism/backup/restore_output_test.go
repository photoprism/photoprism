package backup

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRestoreOutput checks how restore client output is collected.
func TestRestoreOutput(t *testing.T) {
	t.Run("FailedStatements", func(t *testing.T) {
		o := &restoreOutput{}
		_, _ = fmt.Fprint(o, "--------------\nINSERT INTO t VALUES (1,'val-a')\n--------------\n\n",
			"ERROR 1062 (23000) at line 4: Duplicate entry 'val-a' for key 'PRIMARY'\n")
		o.Close()
		assert.Equal(t, 1, o.failed)
		assert.Equal(t, []string{"error 1062 at line 4"}, o.errors)
		assert.Empty(t, o.String())
	})
	t.Run("MultiLineEcho", func(t *testing.T) {
		o := &restoreOutput{}
		_, _ = fmt.Fprint(o, "--------------\nCREATE TABLE `t` (\n  `v` varchar(20) DEFAULT 'val-b'\n) ENGINE=X\n--------------\n\n",
			"ERROR 1286 (42000) at line 9: Unknown storage engine 'X'\nWARNING: option ignored\n")
		o.Close()
		assert.Equal(t, 1, o.failed)
		assert.Empty(t, o.String(), "warnings after statement output are not kept")
	})
	t.Run("ErrorsKept", func(t *testing.T) {
		o := &restoreOutput{}
		for i := 1; i <= restoreErrorsKept+3; i++ {
			_, _ = fmt.Fprintf(o, "ERROR 1062 (23000) at line %d: Duplicate entry 'val-%d' for key 'PRIMARY'\n", i, i)
		}
		o.Close()
		assert.Equal(t, restoreErrorsKept+3, o.failed)
		assert.Len(t, o.errors, restoreErrorsKept)
		assert.Equal(t, "error 1062 at line 1", o.errors[0])
	})
	t.Run("LongLineBounded", func(t *testing.T) {
		// A statement of several megabytes is never held in memory as a whole.
		o := &restoreOutput{}
		_, _ = o.Write([]byte("--------------\nINSERT INTO t VALUES "))
		chunk := []byte(strings.Repeat("x", 64<<10))
		for i := 0; i < 64; i++ {
			_, _ = o.Write(chunk)
			assert.LessOrEqual(t, len(o.line), restoreLineBytes)
		}
		_, _ = o.Write([]byte("\n--------------\n\nERROR 1062 (23000) at line 2: Duplicate entry\n"))
		o.Close()
		assert.Equal(t, 1, o.failed)
		assert.Empty(t, o.String())
	})
	t.Run("OtherLinesKept", func(t *testing.T) {
		o := &restoreOutput{}
		for i := 0; i < restoreLinesKept+10; i++ {
			_, _ = fmt.Fprintf(o, "mariadb: message %d\n", i)
		}
		_, _ = o.Write([]byte(strings.Repeat("y", restoreLineBytes*2)))
		o.Close()
		assert.Len(t, strings.Split(o.String(), "\n"), restoreLinesKept)
		assert.Zero(t, o.failed)
	})
	t.Run("ErrorInsideEcho", func(t *testing.T) {
		// An unclosed echo block does not hide the failed statements that follow.
		o := &restoreOutput{}
		_, _ = fmt.Fprint(o, "--------------\nINSERT INTO t VALUES (1,'val-a')\n\n",
			"ERROR 1062 (23000) at line 4: Duplicate entry 'val-a' for key 'PRIMARY'\n")
		o.Close()
		assert.Equal(t, 1, o.failed)
		assert.Equal(t, []string{"error 1062 at line 4"}, o.errors)
		assert.Empty(t, o.String())
	})
	t.Run("CarriageReturn", func(t *testing.T) {
		o := &restoreOutput{}
		_, _ = fmt.Fprint(o, "WARNING: option ignored\r\n--------------\r\nINSERT INTO t VALUES (1,'val-a')\r\n--------------\r\n\r\n",
			"ERROR 1062 (23000) at line 4: Duplicate entry 'val-a' for key 'PRIMARY'\r\n")
		o.Close()
		assert.Equal(t, 1, o.failed)
		assert.Equal(t, "WARNING: option ignored", o.String())
	})
	t.Run("ClientError", func(t *testing.T) {
		// A client error is counted as a failed statement and also kept in full.
		o := &restoreOutput{}
		_, _ = fmt.Fprint(o, "ERROR 2013 (HY000) at line 812: Lost connection to server during query\n")
		o.Close()
		assert.Equal(t, 1, o.failed)
		assert.Equal(t, []string{"error 2013 at line 812"}, o.errors)
		assert.Equal(t, "ERROR 2013 (HY000) at line 812: Lost connection to server during query", o.String())
	})
	t.Run("KeptLineBounded", func(t *testing.T) {
		o := &restoreOutput{}
		_, _ = o.Write([]byte("mariadb: " + strings.Repeat("y", restoreLineBytes*2) + "\n"))
		o.Close()
		assert.Len(t, o.String(), restoreLineBytes)
	})
	t.Run("DelimiterInsideStatement", func(t *testing.T) {
		// A statement line that looks like a delimiter does not end the skipped block early.
		o := &restoreOutput{}
		_, _ = fmt.Fprint(o, "--------------\nINSERT INTO t VALUES ('val-a\n--------------\nWARNING: val-b')\n--------------\n\n",
			"ERROR 1062 (23000) at line 4: Duplicate entry 'val-a' for key 'PRIMARY'\n",
			"--------------\nINSERT INTO t VALUES ('val-c')\n--------------\n\n",
			"ERROR 1062 (23000) at line 5: Duplicate entry 'val-c' for key 'PRIMARY'\nWARNING: after\n")
		o.Close()
		assert.Equal(t, 2, o.failed)
		assert.Empty(t, o.String())
	})
	t.Run("Sqlite", func(t *testing.T) {
		// SQLite errors are counted without the statement text the client quotes.
		o := &restoreOutput{}
		_, _ = fmt.Fprint(o, "Parse error near line 3: near \"VALUE\": syntax error\n",
			"  INSERT INTO t VALUE('k2','val-c');\n",
			"                ^--- error here\n",
			"Runtime error near line 4: UNIQUE constraint failed: t.id (19)\n")
		o.Close()
		assert.Equal(t, 2, o.failed)
		assert.Equal(t, []string{"error at line 3", "error at line 4"}, o.errors)
		assert.Empty(t, o.String())
		assert.True(t, o.OnlyFailedStatements())
	})
	t.Run("WarningBeforeStatements", func(t *testing.T) {
		o := &restoreOutput{}
		_, _ = fmt.Fprint(o, "WARNING: option --ssl-verify-server-cert is disabled\n",
			"ERROR 1062 (23000) at line 4: Duplicate entry 'val-a' for key 'PRIMARY'\n", "WARNING: val-b\n")
		o.Close()
		assert.Equal(t, "WARNING: option --ssl-verify-server-cert is disabled", o.String())
		assert.True(t, o.OnlyFailedStatements())
	})
	t.Run("UnknownLinesDropped", func(t *testing.T) {
		// A continuation of a quoted error message or any other unknown line is never kept.
		o := &restoreOutput{}
		_, _ = fmt.Fprint(o, "ERROR 1064 (42000) at line 1: You have an error in your SQL syntax near 'val-a\n",
			"second-line val-b' at line 1\n", "val-c\n")
		o.Close()
		assert.Equal(t, 1, o.failed)
		assert.Empty(t, o.String())
		assert.True(t, o.OnlyFailedStatements())
	})
	t.Run("ErrorShapes", func(t *testing.T) {
		o := &restoreOutput{}
		_, _ = fmt.Fprint(o, "ERROR 1064 at line 2: syntax near 'val-a'\n",
			"ERROR 1064 (42000) at line 3 in file: 'dump.sql': near 'val-b'\n",
			"ERROR at line 4: val-c\n",
			"Error: near line 5: near \"val-d\": syntax error\n")
		o.Close()
		assert.Equal(t, 4, o.failed)
		assert.Equal(t, []string{"error 1064 at line 2", "error 1064 at line 3", "error at line 4", "error at line 5"}, o.errors)
		assert.Empty(t, o.String())
	})
	t.Run("ErrorWithoutLine", func(t *testing.T) {
		// Server errors without a line number keep code and state only, unless they concern the account.
		o := &restoreOutput{}
		_, _ = fmt.Fprint(o, "ERROR 1062 (23000): Duplicate entry 'val-a' for key 'PRIMARY'\n",
			"ERROR 1045 (28000): Access denied for user 'photoprism'@'10.0.0.5' (using password: YES)\n")
		o.Close()
		assert.Zero(t, o.failed)
		assert.Equal(t, "ERROR 1062 (23000)\nERROR 1045 (28000): Access denied for user 'photoprism'@'10.0.0.5' (using password: YES)", o.String())
		assert.False(t, o.OnlyFailedStatements())
	})
	t.Run("ClientPrefix", func(t *testing.T) {
		// Clients prefix their messages with the path they were started with.
		o := &restoreOutput{}
		_, _ = fmt.Fprint(o, "/usr/bin/mariadb: unknown option '--bogus-option'\n",
			"/usr/bin/sqlite3: Error: unknown option: -bogus\n",
			"Error: unable to open database \"/nonexistent/x.db\": unable to open database file\n")
		o.Close()
		assert.Equal(t, "/usr/bin/mariadb: unknown option '--bogus-option'\n/usr/bin/sqlite3: Error: unknown option: -bogus\n"+
			"Error: unable to open database \"/nonexistent/x.db\": unable to open database file", o.String())
		assert.Equal(t, 3, o.problems)
	})
	t.Run("ClientPrefixWarning", func(t *testing.T) {
		// A client warning is kept before the first statement output and is not a problem.
		o := &restoreOutput{}
		_, _ = fmt.Fprint(o, "mysql: [Warning] Using a password on the command line interface can be insecure.\n",
			"ERROR 1062 (23000) at line 4: Duplicate entry 'val-a' for key 'PRIMARY'\n",
			"/usr/bin/mysql: [Warning] val-b\n")
		o.Close()
		assert.Equal(t, "mysql: [Warning] Using a password on the command line interface can be insecure.", o.String())
		assert.True(t, o.OnlyFailedStatements())
	})
	t.Run("SqliteFatalCode", func(t *testing.T) {
		// A result code that fails every statement, such as a read-only database, is a problem.
		o := &restoreOutput{}
		_, _ = fmt.Fprint(o, "Runtime error near line 1: attempt to write a readonly database (8)\n",
			"Runtime error near line 2: UNIQUE constraint failed: t.id (19)\n")
		o.Close()
		assert.Equal(t, 2, o.failed)
		assert.Equal(t, "Runtime error near line 1 (8)", o.String())
		assert.False(t, o.OnlyFailedStatements())
		o = &restoreOutput{}
		_, _ = fmt.Fprint(o, "Error: near line 3: database or disk is full (13)\n")
		o.Close()
		assert.Equal(t, "Error near line 3 (13)", o.String())
	})
	t.Run("LastLineWithoutBreak", func(t *testing.T) {
		o := &restoreOutput{}
		_, _ = o.Write([]byte("ERROR 2026 (HY000): TLS/SSL error"))
		o.Close()
		assert.Equal(t, "ERROR 2026 (HY000): TLS/SSL error", o.String())
		assert.Zero(t, o.failed)
	})
}
