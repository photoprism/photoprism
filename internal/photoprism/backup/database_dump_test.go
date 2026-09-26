package backup

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// seedDumps creates the named files with their names as content and returns the directory.
func seedDumps(t *testing.T, names ...string) string {
	t.Helper()

	dir := t.TempDir()

	for _, name := range names {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("good "+name), fs.ModeBackupFile))
	}

	return dir
}

// dirNames returns the sorted names of all entries in dir, including hidden ones.
func dirNames(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	names := make([]string, 0, len(entries))

	for _, e := range entries {
		names = append(names, e.Name())
	}

	sort.Strings(names)

	return names
}

func TestWriteDump(t *testing.T) {
	t.Run("FailedKeepsExisting", func(t *testing.T) {
		dir := seedDumps(t, "2026-09-23.sql", "2026-09-24.sql", "2026-09-25.sql")
		fileName := filepath.Join(dir, "2026-09-25.sql")
		cmd := exec.Command("sh", "-c", "printf partial; exit 23")

		err := writeDump(cmd, dir, fileName, "", true, 1)

		require.Error(t, err)
		data, readErr := os.ReadFile(fileName)
		require.NoError(t, readErr)
		assert.Equal(t, "good 2026-09-25.sql", string(data))
		assert.Equal(t, []string{"2026-09-23.sql", "2026-09-24.sql", "2026-09-25.sql"}, dirNames(t, dir))
	})
	t.Run("SuccessReplacesAndRotates", func(t *testing.T) {
		dir := seedDumps(t, "2026-09-23.sql", "2026-09-24.sql", "2026-09-25.sql")
		fileName := filepath.Join(dir, "2026-09-25.sql")
		require.NoError(t, os.Chmod(fileName, 0o644))
		cmd := exec.Command("sh", "-c", "printf 'complete dump'")

		err := writeDump(cmd, dir, fileName, "", true, 2)

		require.NoError(t, err)
		data, readErr := os.ReadFile(fileName)
		require.NoError(t, readErr)
		assert.Equal(t, "complete dump", string(data))
		info, statErr := os.Stat(fileName)
		require.NoError(t, statErr)
		assert.Equal(t, fs.ModeBackupFile, info.Mode().Perm())
		assert.Equal(t, []string{"2026-09-24.sql", "2026-09-25.sql"}, dirNames(t, dir))
	})
	t.Run("NewFile", func(t *testing.T) {
		dir := seedDumps(t)
		fileName := filepath.Join(dir, "2026-09-25.sql")
		cmd := exec.Command("sh", "-c", "printf 'complete dump'")

		require.NoError(t, writeDump(cmd, dir, fileName, "", false, 0))

		data, err := os.ReadFile(fileName)
		require.NoError(t, err)
		assert.Equal(t, "complete dump", string(data))
		assert.Equal(t, []string{"2026-09-25.sql"}, dirNames(t, dir))
	})
	t.Run("EmptyNotPublished", func(t *testing.T) {
		dir := seedDumps(t, "2026-09-24.sql", "2026-09-25.sql")
		fileName := filepath.Join(dir, "2026-09-25.sql")
		cmd := exec.Command("sh", "-c", "exit 0")

		err := writeDump(cmd, dir, fileName, "", true, 1)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty")
		data, readErr := os.ReadFile(fileName)
		require.NoError(t, readErr)
		assert.Equal(t, "good 2026-09-25.sql", string(data))
		assert.Equal(t, []string{"2026-09-24.sql", "2026-09-25.sql"}, dirNames(t, dir))
	})
	t.Run("ExistingWithoutForce", func(t *testing.T) {
		dir := seedDumps(t, "2026-09-25.sql")
		fileName := filepath.Join(dir, "2026-09-25.sql")
		cmd := exec.Command("sh", "-c", "printf 'complete dump'")

		err := writeDump(cmd, dir, fileName, "", false, 0)

		require.Error(t, err)
		data, readErr := os.ReadFile(fileName)
		require.NoError(t, readErr)
		assert.Equal(t, "good 2026-09-25.sql", string(data))
		assert.Equal(t, []string{"2026-09-25.sql"}, dirNames(t, dir))
	})
	t.Run("StageFile", func(t *testing.T) {
		dir := seedDumps(t)
		fileName := filepath.Join(dir, "2026-09-25.sql")

		// The command reports the directory entries and their modes while it writes the stage.
		cmd := exec.Command("sh", "-c", `cd "$1" && for f in .* *; do [ -f "$f" ] && stat -c '%a %n' "$f"; done; true`, "sh", dir)

		require.NoError(t, writeDump(cmd, dir, fileName, "", false, 0))

		data, err := os.ReadFile(fileName)
		require.NoError(t, err)
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		require.Len(t, lines, 1)
		mode, stageName, found := strings.Cut(lines[0], " ")
		require.True(t, found)
		assert.Equal(t, "600", mode)
		assert.True(t, strings.HasPrefix(stageName, ".2026-09-25.sql."), stageName)
		matched, matchErr := filepath.Match(SqlBackupFileNamePattern, stageName)
		require.NoError(t, matchErr)
		assert.False(t, matched, stageName)
		assert.Equal(t, []string{"2026-09-25.sql"}, dirNames(t, dir))
	})
	t.Run("FirstBackupWithRetain", func(t *testing.T) {
		dir := seedDumps(t)
		fileName := filepath.Join(dir, "2026-09-25.sql")
		cmd := exec.Command("sh", "-c", "printf 'complete dump'")

		require.NoError(t, writeDump(cmd, dir, fileName, "", false, 1))
		assert.Equal(t, []string{"2026-09-25.sql"}, dirNames(t, dir))
	})
	t.Run("PublishFailsKeepsOlder", func(t *testing.T) {
		dir := seedDumps(t, "2026-09-24.sql", "2026-09-25.sql")
		fileName := filepath.Join(dir, "2026-09-25.sql")
		cmd := exec.Command("sh", "-c", "printf 'complete dump'")

		require.Error(t, writeDump(cmd, dir, fileName, "", false, 1))
		assert.Equal(t, []string{"2026-09-24.sql", "2026-09-25.sql"}, dirNames(t, dir))
	})
	t.Run("ForcedPublishFailsRemovesStage", func(t *testing.T) {
		dir := seedDumps(t)
		fileName := filepath.Join(dir, "2026-09-25.sql")

		// The command occupies the destination with a non-empty directory, which a rename cannot replace.
		cmd := exec.Command("sh", "-c", `printf data; mkdir "$1" && touch "$1/x"`, "sh", fileName)

		require.Error(t, writeDump(cmd, dir, fileName, "", true, 0))
		assert.Equal(t, []string{"2026-09-25.sql"}, dirNames(t, dir))
		assert.DirExists(t, fileName)
	})
	t.Run("WriteError", func(t *testing.T) {
		dir := seedDumps(t, "2026-09-24.sql", "2026-09-25.sql")
		fileName := filepath.Join(dir, "2026-09-25.sql")

		// The command ignores failed writes and exits 0, as the sqlite3 shell does.
		cmd := exec.Command("sh", "-c", "head -c 65536 /dev/zero; exit 0")

		var limit syscall.Rlimit
		require.NoError(t, syscall.Getrlimit(syscall.RLIMIT_FSIZE, &limit))
		require.NoError(t, syscall.Setrlimit(syscall.RLIMIT_FSIZE, &syscall.Rlimit{Cur: 4096, Max: limit.Max}))
		t.Cleanup(func() { _ = syscall.Setrlimit(syscall.RLIMIT_FSIZE, &limit) })
		err := writeDump(cmd, dir, fileName, "", true, 1)
		require.NoError(t, syscall.Setrlimit(syscall.RLIMIT_FSIZE, &limit))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "file too large")
		data, readErr := os.ReadFile(fileName)
		require.NoError(t, readErr)
		assert.Equal(t, "good 2026-09-25.sql", string(data))
		assert.Equal(t, []string{"2026-09-24.sql", "2026-09-25.sql"}, dirNames(t, dir))
	})
	t.Run("SymlinkRefused", func(t *testing.T) {
		dir := seedDumps(t, "target.sql")
		fileName := filepath.Join(dir, "2026-09-25.sql")
		require.NoError(t, os.Symlink("target.sql", fileName))
		marker := filepath.Join(dir, "ran")
		cmd := exec.Command("sh", "-c", `touch "$1"; printf 'complete dump'`, "sh", marker)

		err := writeDump(cmd, dir, fileName, "", true, 0)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "symbolic link")
		assert.NoFileExists(t, marker)
		info, lstatErr := os.Lstat(fileName)
		require.NoError(t, lstatErr)
		assert.Equal(t, os.ModeSymlink, info.Mode().Type())
		data, readErr := os.ReadFile(filepath.Join(dir, "target.sql"))
		require.NoError(t, readErr)
		assert.Equal(t, "good target.sql", string(data))
		assert.Equal(t, []string{"2026-09-25.sql", "target.sql"}, dirNames(t, dir))
	})
	t.Run("DanglingSymlinkRefused", func(t *testing.T) {
		dir := seedDumps(t)
		fileName := filepath.Join(dir, "2026-09-25.sql")
		require.NoError(t, os.Symlink("missing.sql", fileName))
		cmd := exec.Command("sh", "-c", "printf 'complete dump'")

		err := writeDump(cmd, dir, fileName, "", false, 0)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "symbolic link")
		assert.NoFileExists(t, filepath.Join(dir, "missing.sql"))
		assert.Equal(t, []string{"2026-09-25.sql"}, dirNames(t, dir))
	})
	t.Run("Pipe", func(t *testing.T) {
		dir := seedDumps(t)
		fileName := filepath.Join(dir, "dump.pipe")
		require.NoError(t, syscall.Mkfifo(fileName, 0o600))

		// Opening the pipe read-write keeps the reader from blocking on a missing writer.
		reader, err := os.OpenFile(fileName, os.O_RDWR, 0)
		require.NoError(t, err)
		t.Cleanup(func() { _ = reader.Close() })

		done := make(chan error, 1)

		go func() {
			done <- writeDump(exec.Command("sh", "-c", "printf 'complete dump'"), dir, fileName, "", true, 0)
		}()

		select {
		case err = <-done:
			require.NoError(t, err)
		case <-time.After(10 * time.Second):
			t.Fatal("dump to pipe did not complete")
		}

		buf := make([]byte, 64)
		n, err := reader.Read(buf)
		require.NoError(t, err)
		assert.Equal(t, "complete dump", string(buf[:n]))
		info, err := os.Lstat(fileName)
		require.NoError(t, err)
		assert.Equal(t, os.ModeNamedPipe, info.Mode().Type())
		assert.Equal(t, []string{"dump.pipe"}, dirNames(t, dir))
	})
	t.Run("PipeWithDumpName", func(t *testing.T) {
		dir := seedDumps(t)
		fileName := filepath.Join(dir, "2026-09-25.sql")
		require.NoError(t, syscall.Mkfifo(fileName, 0o600))
		done := make(chan error, 1)

		go func() {
			done <- writeDump(exec.Command("sh", "-c", "printf 'complete dump'"), dir, fileName, "", true, 0)
		}()

		select {
		case err := <-done:
			require.Error(t, err)
			assert.Contains(t, err.Error(), "not a regular file")
		case <-time.After(10 * time.Second):
			t.Fatal("dump to pipe with a dump name was not refused")
		}

		assert.Equal(t, []string{"2026-09-25.sql"}, dirNames(t, dir))
	})
	t.Run("WriteErrorBeatsStderr", func(t *testing.T) {
		dir := seedDumps(t, "2026-09-25.sql")
		fileName := filepath.Join(dir, "2026-09-25.sql")

		// The command warns on stderr and fails once its output pipe is closed.
		cmd := exec.Command("sh", "-c", "echo 'Warning: slow' >&2; head -c 1048576 /dev/zero")

		var limit syscall.Rlimit
		require.NoError(t, syscall.Getrlimit(syscall.RLIMIT_FSIZE, &limit))
		require.NoError(t, syscall.Setrlimit(syscall.RLIMIT_FSIZE, &syscall.Rlimit{Cur: 4096, Max: limit.Max}))
		t.Cleanup(func() { _ = syscall.Setrlimit(syscall.RLIMIT_FSIZE, &limit) })
		err := writeDump(cmd, dir, fileName, "", true, 0)
		require.NoError(t, syscall.Setrlimit(syscall.RLIMIT_FSIZE, &limit))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "file too large")
		assert.NotContains(t, err.Error(), "Warning")
		assert.Equal(t, []string{"2026-09-25.sql"}, dirNames(t, dir))
	})
	t.Run("RemovesStaleStage", func(t *testing.T) {
		stale, fresh := ".2026-09-23.sql.abcd1234.tmp.sql", ".2026-09-24.sql.efgh5678.tmp.sql"
		dir := seedDumps(t, stale, fresh)
		old := time.Now().Add(-staleStageAge - time.Hour)
		require.NoError(t, os.Chtimes(filepath.Join(dir, stale), old, old))
		fileName := filepath.Join(dir, "2026-09-25.sql")
		cmd := exec.Command("sh", "-c", "printf 'complete dump'")

		require.NoError(t, writeDump(cmd, dir, fileName, "", false, 0))
		assert.Equal(t, []string{fresh, "2026-09-25.sql"}, dirNames(t, dir))
	})
	t.Run("MissingDir", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "missing")
		cmd := exec.Command("sh", "-c", "printf 'complete dump'")

		err := writeDump(cmd, dir, filepath.Join(dir, "2026-09-25.sql"), "", false, 0)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create")
	})
}

func TestWriteDumpTo(t *testing.T) {
	t.Run("Device", func(t *testing.T) {
		require.NoError(t, writeDumpTo(exec.Command("sh", "-c", "printf 'complete dump'"), os.DevNull, "", os.Geteuid()))
	})
	t.Run("Symlink", func(t *testing.T) {
		dir := seedDumps(t, "target.sql")
		fileName := filepath.Join(dir, "link.sql")
		require.NoError(t, os.Symlink("target.sql", fileName))

		err := writeDumpTo(exec.Command("sh", "-c", "printf 'complete dump'"), fileName, "", os.Geteuid())

		require.Error(t, err)
		data, readErr := os.ReadFile(filepath.Join(dir, "target.sql"))
		require.NoError(t, readErr)
		assert.Equal(t, "good target.sql", string(data))
	})
	t.Run("RegularFile", func(t *testing.T) {
		dir := seedDumps(t, "manual.sql")
		fileName := filepath.Join(dir, "manual.sql")

		err := writeDumpTo(exec.Command("sh", "-c", "printf 'complete dump'"), fileName, "", os.Geteuid())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "regular file")
		data, readErr := os.ReadFile(fileName)
		require.NoError(t, readErr)
		assert.Equal(t, "good manual.sql", string(data))
	})
	t.Run("OtherOwner", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "dump.pipe")
		require.NoError(t, syscall.Mkfifo(fileName, 0o600))
		marker := filepath.Join(t.TempDir(), "ran")
		cmd := exec.Command("sh", "-c", `touch "$1"; printf 'complete dump'`, "sh", marker)
		done := make(chan error, 1)

		// Without a reader, the pipe is refused before the open, which would block.
		go func() { done <- writeDumpTo(cmd, fileName, "", os.Geteuid()+1) }()

		select {
		case err := <-done:
			require.Error(t, err)
			assert.Contains(t, err.Error(), "not owned")
		case <-time.After(10 * time.Second):
			t.Fatal("pipe of another owner was not refused before the open")
		}

		assert.NoFileExists(t, marker)
	})
	t.Run("CommandFails", func(t *testing.T) {
		require.Error(t, writeDumpTo(exec.Command("sh", "-c", "exit 23"), os.DevNull, "", os.Geteuid()))
	})
	t.Run("Missing", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "missing.sql")

		require.Error(t, writeDumpTo(exec.Command("sh", "-c", "printf 'complete dump'"), fileName, "", os.Geteuid()))
		assert.NoFileExists(t, fileName)
	})
}

func TestDumpTargetOwned(t *testing.T) {
	uid := os.Geteuid()

	t.Run("OwnPipe", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "dump.pipe")
		require.NoError(t, syscall.Mkfifo(fileName, 0o600))
		info, err := os.Lstat(fileName)
		require.NoError(t, err)

		assert.True(t, dumpTargetOwned(info, uid))
		assert.False(t, dumpTargetOwned(info, uid+1))
	})
	t.Run("RootDevice", func(t *testing.T) {
		info, err := os.Stat(os.DevNull)
		require.NoError(t, err)

		assert.True(t, dumpTargetOwned(info, uid+1))
	})
	t.Run("RootPipe", func(t *testing.T) {
		info := fakeFileInfo{mode: os.ModeNamedPipe, sys: &syscall.Stat_t{Uid: 0}}

		assert.False(t, dumpTargetOwned(info, 1000))
		assert.True(t, dumpTargetOwned(info, 0))
	})
	t.Run("OtherDevice", func(t *testing.T) {
		info := fakeFileInfo{mode: os.ModeDevice | os.ModeCharDevice, sys: &syscall.Stat_t{Uid: 1001}}

		assert.False(t, dumpTargetOwned(info, 1000))
		assert.True(t, dumpTargetOwned(info, 1001))
	})
	t.Run("NoStat", func(t *testing.T) {
		assert.False(t, dumpTargetOwned(fakeFileInfo{mode: os.ModeNamedPipe}, uid))
	})
}

// fakeFileInfo is a FileInfo with a given mode and system-specific attributes.
type fakeFileInfo struct {
	os.FileInfo
	mode os.FileMode
	sys  any
}

// Sys returns the system-specific attributes.
func (f fakeFileInfo) Sys() any { return f.sys }

// Mode returns the file mode.
func (f fakeFileInfo) Mode() os.FileMode { return f.mode }

func TestIsDumpName(t *testing.T) {
	assert.True(t, isDumpName("2026-09-25.sql"))
	assert.False(t, isDumpName(".2026-09-25.sql.abcd1234.tmp.sql"))
	assert.False(t, isDumpName("manual.sql"))
	assert.False(t, isDumpName("2026-09-25.sql.gz"))
}

func TestRemoveStaleStages(t *testing.T) {
	t.Run("AgeBoundary", func(t *testing.T) {
		stale, fresh := ".2026-09-23.sql.abcd1234.tmp.sql", ".2026-09-24.sql.efgh5678.tmp.sql"
		other, dump := ".manual.sql.ijkl9012.tmp.sql", "2026-09-20.sql"
		dir := seedDumps(t, stale, fresh, other, dump)
		now := time.Now()

		for name, age := range map[string]time.Duration{stale: 25 * time.Hour, fresh: 23 * time.Hour, other: 48 * time.Hour, dump: 48 * time.Hour} {
			require.NoError(t, os.Chtimes(filepath.Join(dir, name), now.Add(-age), now.Add(-age)))
		}

		removeStaleStages(dir, 24*time.Hour)

		assert.Equal(t, []string{fresh, other, dump}, dirNames(t, dir))
	})
	t.Run("RelativeDir", func(t *testing.T) {
		stale := ".2026-09-23.sql.abcd1234.tmp.sql"
		dir := seedDumps(t, stale)
		old := time.Now().Add(-48 * time.Hour)
		require.NoError(t, os.Chtimes(filepath.Join(dir, stale), old, old))
		t.Chdir(dir)

		removeStaleStages(".", 24*time.Hour)

		assert.Empty(t, dirNames(t, dir))
	})
	t.Run("MissingDir", func(t *testing.T) {
		removeStaleStages(filepath.Join(t.TempDir(), "missing"), time.Hour)
	})
}

func TestDumpWriter(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var out strings.Builder
		w := &dumpWriter{w: &out}

		n, err := w.Write([]byte("dump"))

		require.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, "dump", out.String())
		assert.NoError(t, w.err)
	})
	t.Run("KeepsFirstError", func(t *testing.T) {
		f, err := os.Create(filepath.Join(t.TempDir(), "closed.sql"))
		require.NoError(t, err)
		require.NoError(t, f.Close())
		w := &dumpWriter{w: f}

		_, firstErr := w.Write([]byte("a"))
		_, _ = w.Write([]byte("b"))

		require.Error(t, firstErr)
		assert.Same(t, firstErr, w.err)
	})
}

func TestRunDump(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var out strings.Builder

		require.NoError(t, runDump(exec.Command("sh", "-c", "printf ok"), &out, ""))
		assert.Equal(t, "ok", out.String())
	})
	t.Run("ExitStatus", func(t *testing.T) {
		err := runDump(exec.Command("sh", "-c", "exit 23"), io.Discard, "")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "23")
	})
	t.Run("StderrMasksPassword", func(t *testing.T) {
		err := runDump(exec.Command("sh", "-c", "echo 'access denied for s3cr3tpass' >&2; exit 1"), io.Discard, "s3cr3tpass")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "access denied")
		assert.NotContains(t, err.Error(), "s3cr3tpass")
	})
	t.Run("StderrWithoutWarnings", func(t *testing.T) {
		hook := captureLog(t)
		err := runDump(exec.Command("sh", "-c", "echo 'WARNING: insecure' >&2; echo 'Got error: 2005' >&2; echo 'when connecting' >&2; exit 2"), io.Discard, "")

		require.Error(t, err)
		assert.Equal(t, "Got error: 2005; when connecting", err.Error())
		assert.Contains(t, logMessages(hook), "backup: insecure")
	})
}

func TestRunRestore(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// The client reads the dump from its input.
		out := filepath.Join(t.TempDir(), "restored.sql")
		require.NoError(t, runRestore(exec.Command("sh", "-c", "cat > '"+out+"'"), strings.NewReader("SELECT 1;\n"), ""))
		data, err := os.ReadFile(out)
		require.NoError(t, err)
		assert.Equal(t, "SELECT 1;\n", string(data))
	})
	t.Run("ExitStatus", func(t *testing.T) {
		err := runRestore(exec.Command("sh", "-c", "exit 23"), strings.NewReader(""), "")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "23")
	})
	t.Run("StderrWithoutWarnings", func(t *testing.T) {
		hook := captureLog(t)
		script := "echo 'WARNING: insecure s3cr3tpass' >&2; echo 'ERROR 2026 (HY000): TLS/SSL error' >&2; echo 'for s3cr3tpass' >&2; exit 1"
		err := runRestore(exec.Command("sh", "-c", script), strings.NewReader(""), "s3cr3tpass")

		require.Error(t, err)
		assert.Equal(t, "ERROR 2026 (HY000): TLS/SSL error; for "+txt.Masked, err.Error())
		assert.Contains(t, logMessages(hook), "restore: insecure "+txt.Masked)
	})
	t.Run("PipedFromFile", func(t *testing.T) {
		// A file or terminal is passed through a pipe, so the client never reads it interactively.
		dir := t.TempDir()
		in, out := filepath.Join(dir, "dump.sql"), filepath.Join(dir, "stdin")
		require.NoError(t, os.WriteFile(in, []byte("SELECT 1;\n"), 0o600))
		f, err := os.Open(in)
		require.NoError(t, err)
		defer f.Close()

		require.NoError(t, runRestore(exec.Command("sh", "-c", "if [ -p /dev/stdin ]; then echo pipe; else echo other; fi > '"+out+"'; cat >/dev/null"), f, ""))
		data, err := os.ReadFile(out)
		require.NoError(t, err)
		assert.Equal(t, "pipe\n", string(data))
	})
	t.Run("EarlyExit", func(t *testing.T) {
		// A client that fails before reading its input reports its own error.
		err := runRestore(exec.Command("sh", "-c", "echo 'ERROR 2026 (HY000): TLS/SSL error' >&2; exit 1"), strings.NewReader(strings.Repeat("x", 1<<20)), "")

		require.Error(t, err)
		assert.Equal(t, "ERROR 2026 (HY000): TLS/SSL error", err.Error())
	})
	t.Run("EarlyExitWhileInputBlocks", func(t *testing.T) {
		// The result does not wait for input that never arrives, such as a terminal nobody types into.
		pr, pw := io.Pipe()
		t.Cleanup(func() { _ = pw.Close() })

		done := make(chan error, 1)
		go func() {
			done <- runRestore(exec.Command("sh", "-c", "echo 'ERROR 2026 (HY000): TLS/SSL error' >&2; exit 1"), pr, "")
		}()

		select {
		case err := <-done:
			assert.EqualError(t, err, "ERROR 2026 (HY000): TLS/SSL error")
		case <-time.After(10 * time.Second):
			t.Fatal("restore did not return after the client exited")
		}
	})
}

// logMessages returns the messages of the captured log entries.
func logMessages(hook *test.Hook) []string {
	var result []string

	for _, entry := range hook.AllEntries() {
		result = append(result, entry.Message)
	}

	return result
}

func TestRotateDumps(t *testing.T) {
	t.Run("Retain", func(t *testing.T) {
		dir := seedDumps(t, "2026-09-23.sql", "2026-09-24.sql", "2026-09-25.sql", ".2026-09-22.sql.abcd1234.tmp.sql", "manual.sql")

		require.NoError(t, rotateDumps(dir, 1))
		assert.Equal(t, []string{".2026-09-22.sql.abcd1234.tmp.sql", "2026-09-25.sql", "manual.sql"}, dirNames(t, dir))
	})
	t.Run("RelativeDir", func(t *testing.T) {
		dir := seedDumps(t, "2026-09-23.sql", "2026-09-24.sql", "2026-09-25.sql")
		t.Chdir(dir)

		require.NoError(t, rotateDumps(".", 1))
		assert.Equal(t, []string{"2026-09-25.sql"}, dirNames(t, dir))
	})
	t.Run("Disabled", func(t *testing.T) {
		dir := seedDumps(t, "2026-09-24.sql", "2026-09-25.sql")

		require.NoError(t, rotateDumps(dir, 0))
		require.NoError(t, rotateDumps("", 1))
		assert.Equal(t, []string{"2026-09-24.sql", "2026-09-25.sql"}, dirNames(t, dir))
	})
	t.Run("NoDumps", func(t *testing.T) {
		dir := seedDumps(t, ".2026-09-25.sql.abcd1234.tmp.sql")

		err := rotateDumps(dir, 1)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "found no database backup files")
	})
}

func TestRestoreDatabase_StageFiles(t *testing.T) {
	t.Run("OnlyStage", func(t *testing.T) {
		dir := seedDumps(t, ".2026-09-25.sql.abcd1234.tmp.sql")

		err := RestoreDatabase(dir, "", false, false)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find a backup")
	})
	t.Run("EmptyDumpAndStage", func(t *testing.T) {
		dir := seedDumps(t, ".2026-09-25.sql.abcd1234.tmp.sql")
		require.NoError(t, os.WriteFile(filepath.Join(dir, "2026-09-24.sql"), nil, fs.ModeBackupFile))

		err := RestoreDatabase(dir, "", false, false)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open 2026-09-24.sql")
	})
	t.Run("RelativeDir", func(t *testing.T) {
		dir := seedDumps(t, ".2026-09-25.sql.abcd1234.tmp.sql")
		require.NoError(t, os.WriteFile(filepath.Join(dir, "2026-09-24.sql"), nil, fs.ModeBackupFile))
		t.Chdir(dir)

		// The empty dump is selected and refused before anything is restored.
		err := RestoreDatabase(".", "", false, false)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open 2026-09-24.sql")
	})
}
