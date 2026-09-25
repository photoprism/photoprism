package backup

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/dsn"
)

// fakeClient writes a shell script that records its arguments, prints output, and exits with code.
func fakeClient(t *testing.T, output string, code int) (bin, runs string) {
	t.Helper()

	dir := t.TempDir()
	bin = filepath.Join(dir, "client")
	runs = filepath.Join(dir, "runs")
	outFile := filepath.Join(dir, "output")
	require.NoError(t, os.WriteFile(outFile, []byte(output+"\n"), 0o600))
	script := "#!/bin/sh\necho \"$*\" >> '" + runs + "'\ncat '" + outFile + "'\nexit " + strconv.Itoa(code) + "\n"
	require.NoError(t, os.WriteFile(bin, []byte(script), 0o700))

	return bin, runs
}

// runArgs returns the arguments of each run of a fake client.
func runArgs(t *testing.T, runs string) []string {
	t.Helper()

	data, err := os.ReadFile(runs)

	if os.IsNotExist(err) {
		return nil
	}

	require.NoError(t, err)

	return strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
}

func TestFakeClient(t *testing.T) {
	bin, runs := fakeClient(t, "mysql: [ERROR] unknown option '--sandbox'.", 7)

	out, err := exec.Command(bin, "--sandbox", "--version").CombinedOutput()

	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 7, exitErr.ExitCode())
	assert.Equal(t, "mysql: [ERROR] unknown option '--sandbox'.\n", string(out))
	assert.Equal(t, []string{"--sandbox --version"}, runArgs(t, runs))
}

func TestClientAccepts(t *testing.T) {
	t.Run("Accepted", func(t *testing.T) {
		bin, runs := fakeClient(t, "client 1.0", 0)

		for i := 0; i < 2; i++ {
			ok, err := clientAccepts(bin, "--sandbox", "--version")
			assert.True(t, ok)
			assert.NoError(t, err)
		}

		assert.Equal(t, []string{"--sandbox --version"}, runArgs(t, runs))
	})
	t.Run("UnknownOptionWithoutError", func(t *testing.T) {
		bin, _ := fakeClient(t, "client: unknown option --sandbox", 0)

		ok, err := clientAccepts(bin, "--sandbox", "--version")

		assert.False(t, ok)
		assert.NoError(t, err)
	})
	t.Run("ErrorExit", func(t *testing.T) {
		bin, runs := fakeClient(t, "client: Error: bad flag", 1)

		for i := 0; i < 2; i++ {
			ok, err := clientAccepts(bin, "-safe", "-version")
			assert.False(t, ok)
			assert.NoError(t, err)
		}

		assert.Len(t, runArgs(t, runs), 1)
	})
	t.Run("PerArguments", func(t *testing.T) {
		bin, runs := fakeClient(t, "client 1.0", 0)

		ok, _ := clientAccepts(bin, "-safe", "-version")
		assert.True(t, ok)
		ok, _ = clientAccepts(bin, "--sandbox", "--version")
		assert.True(t, ok)
		assert.Equal(t, []string{"-safe -version", "--sandbox --version"}, runArgs(t, runs))
	})
	t.Run("Missing", func(t *testing.T) {
		ok, err := clientAccepts(filepath.Join(t.TempDir(), "missing"), "-safe", "-version")

		assert.False(t, ok)
		assert.Error(t, err)
	})
	t.Run("Timeout", func(t *testing.T) {
		prevTimeout, prevDelay := clientProbeTimeout, clientProbeWaitDelay
		t.Cleanup(func() { clientProbeTimeout, clientProbeWaitDelay = prevTimeout, prevDelay })
		clientProbeTimeout, clientProbeWaitDelay = 200*time.Millisecond, 200*time.Millisecond

		// The child keeps the output open after the script itself is stopped.
		bin := filepath.Join(t.TempDir(), "client")
		require.NoError(t, os.WriteFile(bin, []byte("#!/bin/sh\nexec sleep 30\n"), 0o700))
		start := time.Now()

		ok, err := clientAccepts(bin, "-safe", "-version")

		assert.False(t, ok)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		assert.Less(t, time.Since(start), 10*time.Second)
	})
	t.Run("ChildKeepsOutput", func(t *testing.T) {
		prevTimeout, prevDelay := clientProbeTimeout, clientProbeWaitDelay
		t.Cleanup(func() { clientProbeTimeout, clientProbeWaitDelay = prevTimeout, prevDelay })
		clientProbeTimeout, clientProbeWaitDelay = 5*time.Second, 200*time.Millisecond

		// The client exits after printing its version, while a child it started keeps the output open.
		bin := filepath.Join(t.TempDir(), "client")
		require.NoError(t, os.WriteFile(bin, []byte("#!/bin/sh\necho client 1.0\nsleep 3 &\nexit 0\n"), 0o700))

		ok, err := clientAccepts(bin, "-safe", "-version")

		assert.True(t, ok)
		assert.NoError(t, err)
	})
}

func TestMariadbRestoreArgs(t *testing.T) {
	t.Run("Sandbox", func(t *testing.T) {
		bin, runs := fakeClient(t, "mariadb 11.8", 0)

		assert.Equal(t, []string{"-f", "--local-infile=0", "--binary-mode", "--sandbox"}, mariadbRestoreArgs(bin))
		assert.Equal(t, []string{"--no-defaults --sandbox --version"}, runArgs(t, runs))
	})
	t.Run("NoSandbox", func(t *testing.T) {
		bin, _ := fakeClient(t, "mysql: [ERROR] unknown option '--sandbox'.", 7)

		assert.Equal(t, []string{"-f", "--local-infile=0", "--binary-mode"}, mariadbRestoreArgs(bin))
	})
	t.Run("ProbeFailed", func(t *testing.T) {
		assert.Equal(t, []string{"-f", "--local-infile=0", "--binary-mode"}, mariadbRestoreArgs(filepath.Join(t.TempDir(), "missing")))
	})
}

func TestSqliteRestoreCmd(t *testing.T) {
	t.Run("Safe", func(t *testing.T) {
		bin, runs := fakeClient(t, "3.46.1", 0)

		assert.Equal(t, []string{bin, "-safe", "index.db"}, sqliteRestoreCmd(bin, "index.db").Args)
		assert.Equal(t, []string{"-safe -version"}, runArgs(t, runs))
	})
	t.Run("NoSafe", func(t *testing.T) {
		bin, _ := fakeClient(t, "sqlite3: Error: unknown option: -safe", 1)

		assert.Equal(t, []string{bin, "index.db"}, sqliteRestoreCmd(bin, "index.db").Args)
	})
	t.Run("MissingClient", func(t *testing.T) {
		cmd := sqliteRestoreCmd(filepath.Base(t.TempDir())+"-missing-client", "index.db")

		assert.Error(t, cmd.Err)
	})
	t.Run("RoundTrip", func(t *testing.T) {
		if get.Config().DatabaseDriver() != dsn.DriverSQLite3 {
			t.Skip("the test database is not a SQLite file")
		}

		bin, srcFile := get.Config().SqliteBin(), get.Config().DatabaseFile()

		if ok, _ := clientAccepts(bin, "-safe", "-version"); bin == "" || !ok {
			t.Skip("sqlite3 with safe mode is not available")
		}

		dump, err := exec.Command(bin, srcFile, ".dump").Output()
		require.NoError(t, err)
		require.Contains(t, string(dump), "CREATE TABLE")

		dbFile := filepath.Join(t.TempDir(), "restored.db")
		cmd := sqliteRestoreCmd(bin, dbFile)
		cmd.Stdin = strings.NewReader(string(dump))
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))

		restored, err := exec.Command(bin, dbFile, ".dump").Output()
		require.NoError(t, err)
		assert.Equal(t, string(dump), string(restored))
	})
	t.Run("SafeModeApplied", func(t *testing.T) {
		bin := get.Config().SqliteBin()

		if ok, _ := clientAccepts(bin, "-safe", "-version"); bin == "" || !ok {
			t.Skip("sqlite3 with safe mode is not available")
		}

		// Safe mode refuses the dot-command, so no marker file is created.
		dir := t.TempDir()
		marker := filepath.Join(dir, "marker")
		cmd := sqliteRestoreCmd(bin, filepath.Join(dir, "restored.db"))
		cmd.Stdin = strings.NewReader(".shell touch '" + marker + "'\nCREATE TABLE t(x);\n")

		_ = cmd.Run()

		assert.NoFileExists(t, marker)
	})
}

func TestRestoreCmd(t *testing.T) {
	c := get.Config()

	cmd, password, err := restoreCmd(c)

	require.NoError(t, err)

	switch c.DatabaseDriver() {
	case dsn.DriverSQLite3:
		assert.Equal(t, sqliteRestoreCmd(c.SqliteBin(), c.DatabaseFile()).Args, cmd.Args)
		assert.Empty(t, password)
	case dsn.DriverMySQL, dsn.DriverMariaDB:
		assert.Subset(t, cmd.Args, mariadbRestoreArgs(c.MariadbBin()))
		assert.Equal(t, "--no-defaults", cmd.Args[1])
		assert.Equal(t, c.DatabaseName(), cmd.Args[len(cmd.Args)-1])
	default:
		t.Skipf("unsupported test driver %s", c.DatabaseDriver())
	}
}

func TestRestoreCmd_MariaDB(t *testing.T) {
	c := config.NewMinimalTestConfig(t.TempDir())
	c.Options().DatabaseDriver = dsn.DriverMySQL

	if c.MariadbBin() == "" {
		t.Skip("mariadb client not found")
	}

	cmd, password, err := restoreCmd(c)

	require.NoError(t, err)
	assert.Equal(t, c.DatabasePassword(), password)
	assert.Equal(t, newMariadbConn(c, c.MariadbBin()).Cmd(mariadbRestoreArgs(c.MariadbBin())...).Args, cmd.Args)
	assert.Equal(t, "--no-defaults", cmd.Args[1])
	assert.Equal(t, c.DatabaseName(), cmd.Args[len(cmd.Args)-1])
	assert.Contains(t, cmd.Args, "--local-infile=0")
	assert.Contains(t, cmd.Args, "--binary-mode")
}
