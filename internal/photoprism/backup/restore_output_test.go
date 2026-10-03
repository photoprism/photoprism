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
			"ERROR 1062 (23000) at line 4: Duplicate entry 'val-a' for key 'PRIMARY'\n",
			"ERROR at line 5: Unknown command '\\!'.\n",
			"ERROR 1064 (42000) at line 6 in file: 'dump.sql': near 'val-b'\n")
		o.Close()
		assert.Equal(t, restoreFailures{Count: 3, Errors: []string{"error 1062 at line 4", "error at line 5", "error 1064 at line 6"}}, o.Failures())
	})
	t.Run("ErrorsKept", func(t *testing.T) {
		o := &restoreOutput{}
		for i := 1; i <= restoreErrorsKept+3; i++ {
			_, _ = fmt.Fprintf(o, "ERROR 1062 (23000) at line %d: Duplicate entry 'val-%d' for key 'PRIMARY'\n", i, i)
		}
		o.Close()
		assert.Equal(t, restoreErrorsKept+3, o.failed)
		assert.Len(t, o.errors, 5)
		assert.Equal(t, "error 1062 at line 1", o.errors[0])
	})
	t.Run("FirstLinesKept", func(t *testing.T) {
		// A failed restore returns these lines in its error, which the CLI prints to the operator's terminal;
		// they are not logged, as restores run only from the CLI before logs are recorded in the database.
		o := &restoreOutput{}
		_, _ = fmt.Fprint(o, "ERROR 2002 (HY000): Can't connect to server on 'mariadb' (115)\n\n")
		for i := 0; i < restoreLinesKept+10; i++ {
			_, _ = fmt.Fprintf(o, "message %d\n", i)
		}
		o.Close()
		lines := strings.Split(o.String(), "\n")
		assert.Len(t, lines, 5)
		assert.Equal(t, "ERROR 2002 (HY000): Can't connect to server on 'mariadb' (115)", lines[0])
		assert.Equal(t, "message 0", lines[1])
		assert.Zero(t, o.failed)
	})
	t.Run("OtherLinesNotCounted", func(t *testing.T) {
		o := &restoreOutput{}
		_, _ = fmt.Fprint(o, "Parse error near line 3: near \"x\": syntax error\n", "val-a' at line 1\n", "error at line 2\n")
		o.Close()
		assert.Zero(t, o.failed)
	})
	t.Run("Warnings", func(t *testing.T) {
		// Warnings are kept only before any other output, as the client writes them when connecting.
		o := &restoreOutput{}
		_, _ = fmt.Fprint(o, "WARNING: option --ssl-verify-server-cert is disabled\n", "WARNING: other\n",
			"--------------\n", "WARNING: val-a\n")
		o.Close()
		assert.Equal(t, "WARNING: option --ssl-verify-server-cert is disabled\nWARNING: other", o.Warnings())
	})
	t.Run("LongLineBounded", func(t *testing.T) {
		// A statement of several megabytes is never held in memory as a whole.
		o := &restoreOutput{}
		_, _ = o.Write([]byte("INSERT INTO t VALUES "))
		chunk := []byte(strings.Repeat("x", 64<<10))
		for i := 0; i < 64; i++ {
			_, _ = o.Write(chunk)
			assert.LessOrEqual(t, len(o.line), restoreLineBytes)
		}
		_, _ = o.Write([]byte("\nERROR 1062 (23000) at line 2: Duplicate entry\n"))
		o.Close()
		assert.Equal(t, 1, o.failed)
		assert.Len(t, strings.Split(o.String(), "\n")[0], 1024)
	})
	t.Run("LastLineWithoutBreak", func(t *testing.T) {
		o := &restoreOutput{}
		_, _ = o.Write([]byte("ERROR 2026 (HY000): TLS/SSL error\r"))
		o.Close()
		assert.Equal(t, "ERROR 2026 (HY000): TLS/SSL error", o.String())
		assert.Zero(t, o.failed)
	})
}
