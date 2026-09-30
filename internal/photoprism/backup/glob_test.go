package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestGlobIn(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "sub")
	special := filepath.Join(root, "v1.0+[x]")
	require.NoError(t, os.MkdirAll(sub, fs.ModeDir))
	require.NoError(t, os.MkdirAll(special, fs.ModeDir))

	for _, dir := range []string{root, sub, special} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "2026-09-25.sql"), []byte("dump"), fs.ModeBackupFile))
	}

	t.Run("Absolute", func(t *testing.T) {
		files, err := globIn(sub, SqlBackupFileNamePattern)

		require.NoError(t, err)
		assert.Equal(t, []string{filepath.Join(sub, "2026-09-25.sql")}, files)
	})
	t.Run("Metacharacters", func(t *testing.T) {
		files, err := globIn(special, SqlBackupFileNamePattern)

		require.NoError(t, err)
		assert.Equal(t, []string{filepath.Join(special, "2026-09-25.sql")}, files)
	})
	t.Run("Relative", func(t *testing.T) {
		t.Chdir(sub)

		for dir, want := range map[string]string{
			".":      filepath.Join(sub, "2026-09-25.sql"),
			"./":     filepath.Join(sub, "2026-09-25.sql"),
			"..":     filepath.Join(root, "2026-09-25.sql"),
			"../sub": filepath.Join(sub, "2026-09-25.sql"),
		} {
			files, err := globIn(dir, SqlBackupFileNamePattern)

			require.NoError(t, err, dir)
			assert.Equal(t, []string{want}, files, dir)
		}
	})
	t.Run("RelativeSubdir", func(t *testing.T) {
		t.Chdir(root)

		for _, dir := range []string{"sub", "./sub"} {
			files, err := globIn(dir, SqlBackupFileNamePattern)

			require.NoError(t, err, dir)
			assert.Equal(t, []string{filepath.Join(sub, "2026-09-25.sql")}, files, dir)
		}
	})
	t.Run("NonUTF8", func(t *testing.T) {
		dir := filepath.Join(root, "caf\xe9")
		decoy := filepath.Join(root, "caf\uFFFD")
		require.NoError(t, os.MkdirAll(dir, fs.ModeDir))
		require.NoError(t, os.MkdirAll(decoy, fs.ModeDir))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "2026-09-25.sql"), []byte("dump"), fs.ModeBackupFile))
		require.NoError(t, os.WriteFile(filepath.Join(decoy, "2020-01-01.sql"), []byte("decoy"), fs.ModeBackupFile))

		files, err := globIn(dir, SqlBackupFileNamePattern)

		require.NoError(t, err)
		assert.Equal(t, []string{filepath.Join(dir, "2026-09-25.sql")}, files)
	})
	t.Run("UnreadableAncestor", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("root can read any directory")
		}

		// Only the dir itself must be listed, so a parent that can be traversed but not read is fine.
		parent := filepath.Join(root, "p")
		dir := filepath.Join(parent, "c.d+", "backup")
		require.NoError(t, os.MkdirAll(dir, fs.ModeDir))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "2026-09-25.sql"), []byte("dump"), fs.ModeBackupFile))
		require.NoError(t, os.Chmod(parent, 0o311))
		t.Cleanup(func() { _ = os.Chmod(parent, 0o755) })

		files, err := globIn(dir, SqlBackupFileNamePattern)

		require.NoError(t, err)
		assert.Equal(t, []string{filepath.Join(dir, "2026-09-25.sql")}, files)
	})
	t.Run("Missing", func(t *testing.T) {
		files, err := globIn(filepath.Join(root, "missing"), SqlBackupFileNamePattern)

		require.NoError(t, err)
		assert.Empty(t, files)
	})
	t.Run("BadPattern", func(t *testing.T) {
		_, err := globIn(root, "[")

		assert.Error(t, err)
	})
}

func TestGlobEscape(t *testing.T) {
	assert.Equal(t, "/srv/backup/mysql", globEscape("/srv/backup/mysql"))
	assert.Equal(t, "./v1.0+{x}", globEscape("./v1.0+{x}"))
	assert.Equal(t, `/a\*b\?c\[d]\\e`, globEscape(`/a*b?c[d]\e`))
	assert.Equal(t, "", globEscape(""))
	assert.Equal(t, "caf\xe9", globEscape("caf\xe9"))
	assert.Equal(t, "a\xffb\\*", globEscape("a\xffb*"))

	for _, name := range []string{"a*", "b[", "c?", "d]", `e\x`, "f.g+h", "caf\xe9", "a\xffb*"} {
		matched, err := filepath.Match(globEscape(name), name)

		require.NoError(t, err, name)
		assert.True(t, matched, name)
	}
}

func TestDatabase_SymlinkedWorkingDir(t *testing.T) {
	root := t.TempDir()
	physical := filepath.Join(root, "data", "deep")
	logical := filepath.Join(root, "logical")
	require.NoError(t, os.MkdirAll(filepath.Join(physical, "cwd"), fs.ModeDir))
	require.NoError(t, os.MkdirAll(filepath.Join(physical, "x"), fs.ModeDir))
	require.NoError(t, os.MkdirAll(filepath.Join(logical, "x"), fs.ModeDir))
	require.NoError(t, os.Symlink(filepath.Join(physical, "cwd"), filepath.Join(logical, "cwdlink")))

	for _, name := range []string{"2026-09-20.sql", "2026-09-21.sql"} {
		require.NoError(t, os.WriteFile(filepath.Join(physical, "x", name), []byte("physical"), fs.ModeBackupFile))
		require.NoError(t, os.WriteFile(filepath.Join(logical, "x", name), []byte("logical"), fs.ModeBackupFile))
	}

	// The dump reads the test database through its relative name, so it is linked into the new working dir.
	dbFile, err := filepath.Abs(get.Config().DatabaseFile())
	require.NoError(t, err)
	require.NoError(t, os.Symlink(dbFile, filepath.Join(physical, "cwd", filepath.Base(dbFile))))

	// The working dir is entered through the link, so "../x" names a different dir than the kernel resolves.
	t.Chdir(filepath.Join(logical, "cwdlink"))

	require.NoError(t, Database("../x", "", false, true, 1))

	today := time.Now().UTC().Format("2006-01-02") + ".sql"
	assert.Equal(t, []string{today}, dirNames(t, filepath.Join(logical, "x")))
	assert.Equal(t, []string{"2026-09-20.sql", "2026-09-21.sql"}, dirNames(t, filepath.Join(physical, "x")))
}
