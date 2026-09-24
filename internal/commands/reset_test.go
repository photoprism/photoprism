package commands

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestResetCommand(t *testing.T) {
	// make sure that database is in a good state for later tests as this test empties it
	defer resetConfigAndDB()

	t.Run("ResetIndex", func(t *testing.T) {
		c := resetConfigAndOpenDB()
		count := int64(0)
		if err := c.Db().Model(&entity.Photo{}).Count(&count).Error; err != nil {
			assert.NoError(t, err)
			return
		}
		assert.Greater(t, count, int64(0))

		dbDrv := os.Getenv("PHOTOPRISM_TEST_DRIVER")
		dbDSN := os.Getenv("PHOTOPRISM_TEST_DSN")
		// Run command with test context.
		appArgs := []string{"photoprism",
			"--database-driver", dbDrv,
			"--database-dsn", dbDSN}
		if dbDrv == "sqlite" {
			appArgs = []string{"photoprism",
				"--database-driver", dbDrv,
				"--database-dsn", c.DatabaseDSN()}
		}
		cmdArgs := []string{"reset", "--index", "--yes"}

		ctx := NewTestContextWithParse(appArgs, cmdArgs)

		// Setup and capture SQL Logging output
		buffer := bytes.Buffer{}
		log.SetOutput(&buffer)

		output, err := RunWithProvidedTestContext(ctx, ResetCommand, cmdArgs)
		// Reset logger
		log.SetOutput(os.Stdout)

		// Check command output for plausibility.
		// t.Logf("buffer = %s", buffer.String())
		assert.NoError(t, err)
		assert.Empty(t, output)
		assert.Contains(t, buffer.String(), "dropping existing tables")
		assert.Contains(t, buffer.String(), "restoring default schema")

		c = reopenConnection()
		if err := c.Db().Model(&entity.Photo{}).Count(&count).Error; err != nil {
			assert.NoError(t, err)
			return
		}
		assert.Equal(t, int64(0), count)
	})
}

// writeResetTestFiles creates each named file below dir, including its parent folders.
func writeResetTestFiles(t *testing.T, dir string, names ...string) {
	t.Helper()

	for _, name := range names {
		fileName := filepath.Join(dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(fileName), fs.ModeDir))
		require.NoError(t, os.WriteFile(fileName, []byte("{}"), fs.ModeFile))
	}
}

// captureResetLog redirects the log output to a buffer until the test ends.
func captureResetLog(t *testing.T) *bytes.Buffer {
	t.Helper()

	var out io.Writer = os.Stderr

	if l, ok := log.(*logrus.Logger); ok {
		out = l.Out
	}

	buffer := &bytes.Buffer{}
	log.SetOutput(buffer)
	t.Cleanup(func() { log.SetOutput(out) })

	return buffer
}

// skipIfRoot skips a test that relies on permissions root would override.
func skipIfRoot(t *testing.T) {
	t.Helper()

	if os.Geteuid() == 0 {
		t.Skip("permissions are not enforced for root")
	}
}

// chmodResetTestDir changes the mode of a test folder and restores it when the test ends.
func chmodResetTestDir(t *testing.T, dir string, mode os.FileMode) {
	t.Helper()

	require.NoError(t, os.Chmod(dir, mode))        // #nosec G302 directory mode
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) }) // #nosec G302 directory mode
}

// resetTestFiles returns the files below dir as slash-separated relative names.
func resetTestFiles(t *testing.T, dir string) (names []string) {
	t.Helper()

	require.NoError(t, filepath.WalkDir(dir, func(name string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		rel, relErr := filepath.Rel(dir, name)
		names = append(names, filepath.ToSlash(rel))

		return relErr
	}))

	return names
}

func TestRemoveFilesWithExt(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		dir := t.TempDir()
		writeResetTestFiles(t, dir, "a.json", "x/b.json", "x/y/c.json", "x/y/z/d.json", "a.yml", "x/y/e.jpg", "x/y/f.json.bak")

		removed, failed, err := removeFilesWithExt(dir, fs.ExtJson)

		require.NoError(t, err)
		assert.Equal(t, 4, removed)
		assert.Equal(t, 0, failed)
		assert.ElementsMatch(t, []string{"a.yml", "x/y/e.jpg", "x/y/f.json.bak"}, resetTestFiles(t, dir))
	})
	t.Run("KeepsSymlinks", func(t *testing.T) {
		dir := t.TempDir()
		outside := t.TempDir()
		writeResetTestFiles(t, outside, "target.json", "nested/other.json")
		require.NoError(t, os.Symlink(filepath.Join(outside, "target.json"), filepath.Join(dir, "link.json")))
		require.NoError(t, os.Symlink(filepath.Join(outside, "nested"), filepath.Join(dir, "nested")))

		removed, failed, err := removeFilesWithExt(dir, fs.ExtJson)

		require.NoError(t, err)
		assert.Equal(t, 0, removed)
		assert.Equal(t, 0, failed)
		assert.FileExists(t, filepath.Join(dir, "link.json"))
		assert.ElementsMatch(t, []string{"target.json", "nested/other.json"}, resetTestFiles(t, outside))
	})
	t.Run("UnreadableFolder", func(t *testing.T) {
		skipIfRoot(t)

		dir := t.TempDir()
		writeResetTestFiles(t, dir, "a.json", "locked/b.json", "locked/sub/c.json", "x/y/d.json")
		chmodResetTestDir(t, filepath.Join(dir, "locked"), 0o300)

		removed, failed, err := removeFilesWithExt(dir, fs.ExtJson)

		require.Error(t, err)
		assert.ErrorIs(t, err, os.ErrPermission)
		assert.Equal(t, 2, removed)
		assert.Equal(t, 0, failed)
	})
	t.Run("RemoveFailed", func(t *testing.T) {
		skipIfRoot(t)

		dir := t.TempDir()
		writeResetTestFiles(t, dir, "a.json", "locked/b.json")
		chmodResetTestDir(t, filepath.Join(dir, "locked"), 0o555)

		removed, failed, err := removeFilesWithExt(dir, fs.ExtJson)

		require.NoError(t, err)
		assert.Equal(t, 1, removed)
		assert.Equal(t, 1, failed)
		assert.FileExists(t, filepath.Join(dir, "locked", "b.json"))
	})
	t.Run("MissingDir", func(t *testing.T) {
		removed, failed, err := removeFilesWithExt(filepath.Join(t.TempDir(), "missing"), fs.ExtJson)

		require.NoError(t, err)
		assert.Equal(t, 0, removed)
		assert.Equal(t, 0, failed)
	})
	t.Run("NotADirectory", func(t *testing.T) {
		dir := t.TempDir()
		writeResetTestFiles(t, dir, "file.json")

		removed, failed, err := removeFilesWithExt(filepath.Join(dir, "file.json", "sub"), fs.ExtJson)

		assert.Error(t, err)
		assert.Equal(t, 0, removed)
		assert.Equal(t, 0, failed)
		assert.FileExists(t, filepath.Join(dir, "file.json"))
	})
}

func TestResetFiles(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		dir := t.TempDir()
		writeResetTestFiles(t, dir, "a.yml", "x/y/b.yml")

		buffer := captureResetLog(t)

		resetFiles(dir, fs.ExtYml, "*.yml test files")

		assert.Empty(t, resetTestFiles(t, dir))
		assert.Contains(t, buffer.String(), "removed 2 *.yml test files")
	})
	t.Run("NoneFound", func(t *testing.T) {
		buffer := captureResetLog(t)

		resetFiles(t.TempDir(), fs.ExtYml, "*.yml test files")

		assert.Contains(t, buffer.String(), "found no *.yml test files")
	})
	t.Run("RemoveFailed", func(t *testing.T) {
		skipIfRoot(t)

		dir := t.TempDir()
		writeResetTestFiles(t, dir, "locked/a.yml")
		chmodResetTestDir(t, filepath.Join(dir, "locked"), 0o555)

		buffer := captureResetLog(t)

		resetFiles(dir, fs.ExtYml, "*.yml test files")

		assert.Contains(t, buffer.String(), "failed to remove 1 *.yml test files")
		assert.NotContains(t, buffer.String(), "found no")
	})
	t.Run("UnreadableFolder", func(t *testing.T) {
		skipIfRoot(t)

		dir := t.TempDir()
		writeResetTestFiles(t, dir, "locked/a.yml")
		chmodResetTestDir(t, filepath.Join(dir, "locked"), 0o300)

		buffer := captureResetLog(t)

		resetFiles(dir, fs.ExtYml, "*.yml test files")

		assert.Contains(t, buffer.String(), "permission denied")
		assert.NotContains(t, buffer.String(), "found no")
	})
}

func TestResetSidecarJson(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c := config.NewMinimalTestConfig(t.TempDir())
		c.Options().SidecarPath = t.TempDir()
		writeResetTestFiles(t, c.SidecarPath(), "a.json", "x/b.json", "x/y/c.json", "x/y/z/d.json", "x/y/z/e.yml")

		resetSidecarJson(c)

		assert.Equal(t, []string{"x/y/z/e.yml"}, resetTestFiles(t, c.SidecarPath()))
	})
	t.Run("RelativePath", func(t *testing.T) {
		dir := t.TempDir()
		writeResetTestFiles(t, dir, ".photoprism/x/a.json")
		t.Chdir(dir)

		c := config.NewMinimalTestConfig(t.TempDir())
		c.Options().SidecarPath = ".photoprism"

		buffer := captureResetLog(t)

		resetSidecarJson(c)

		assert.Equal(t, []string{".photoprism/x/a.json"}, resetTestFiles(t, dir))
		assert.Contains(t, buffer.String(), "level=warning")
		assert.Contains(t, buffer.String(), "removed 0 *.json sidecar files, because the sidecar path .photoprism is relative")
	})
}

func TestResetSidecarFiles(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c := config.NewMinimalTestConfig(t.TempDir())
		c.Options().SidecarPath = t.TempDir()
		writeResetTestFiles(t, c.SidecarPath(), "a.json", "x/y/b.json", "x/y/c.yml")

		resetSidecarFiles(c, fs.ExtJson, "*.json sidecar files")

		assert.Equal(t, []string{"x/y/c.yml"}, resetTestFiles(t, c.SidecarPath()))
	})
	t.Run("ContainsOriginals", func(t *testing.T) {
		c := config.NewMinimalTestConfig(t.TempDir())
		c.Options().SidecarPath = filepath.Dir(c.OriginalsPath())
		writeResetTestFiles(t, c.OriginalsPath(), "x/y/takeout.json")

		buffer := captureResetLog(t)

		resetSidecarFiles(c, fs.ExtJson, "*.json sidecar files")

		assert.FileExists(t, filepath.Join(c.OriginalsPath(), "x", "y", "takeout.json"))
		assert.Contains(t, buffer.String(), "level=warning")
		assert.Contains(t, buffer.String(), "removed 0 *.json sidecar files, because the sidecar path")
	})
}

// newResetOverlapConfig returns a test config whose originals, import, and backup folders are
// outside the storage folder, as in the Docker layout, so each path is matched on its own.
func newResetOverlapConfig(t *testing.T) *config.Config {
	t.Helper()

	c := config.NewMinimalTestConfig(t.TempDir())
	c.Options().OriginalsPath = t.TempDir()
	c.Options().ImportPath = t.TempDir()
	c.Options().BackupPath = t.TempDir()

	return c
}

func TestResetSidecarOverlap(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		c := config.NewMinimalTestConfig(t.TempDir())

		assert.Equal(t, "", resetSidecarOverlap(c))
	})
	t.Run("SeparateFolder", func(t *testing.T) {
		c := newResetOverlapConfig(t)
		c.Options().SidecarPath = t.TempDir()

		assert.Equal(t, "", resetSidecarOverlap(c))
	})
	t.Run("EqualsOriginals", func(t *testing.T) {
		c := newResetOverlapConfig(t)
		c.Options().SidecarPath = c.OriginalsPath()

		assert.Equal(t, c.OriginalsPath(), resetSidecarOverlap(c))
	})
	t.Run("EqualsImport", func(t *testing.T) {
		c := newResetOverlapConfig(t)
		c.Options().SidecarPath = c.ImportPath()

		assert.Equal(t, c.ImportPath(), resetSidecarOverlap(c))
	})
	t.Run("EqualsStorage", func(t *testing.T) {
		c := newResetOverlapConfig(t)
		c.Options().SidecarPath = c.StoragePath()

		assert.Equal(t, c.StoragePath(), resetSidecarOverlap(c))
	})
	t.Run("ContainsConfig", func(t *testing.T) {
		c := newResetOverlapConfig(t)
		c.Options().StoragePath = t.TempDir()
		c.Options().ConfigPath = filepath.Join(t.TempDir(), "x", "config")
		c.Options().SidecarPath = filepath.Dir(c.ConfigPath())

		assert.Equal(t, c.ConfigPath(), resetSidecarOverlap(c))
	})
	t.Run("EqualsBackup", func(t *testing.T) {
		c := newResetOverlapConfig(t)
		c.Options().SidecarPath = c.BackupBasePath()

		assert.Equal(t, c.BackupBasePath(), resetSidecarOverlap(c))
	})
	t.Run("SymlinkToOriginals", func(t *testing.T) {
		c := newResetOverlapConfig(t)
		link := filepath.Join(t.TempDir(), "sidecar")
		require.NoError(t, os.Symlink(c.OriginalsPath(), link))
		c.Options().SidecarPath = link

		assert.Equal(t, c.OriginalsPath(), resetSidecarOverlap(c))
	})
	t.Run("InsideOriginals", func(t *testing.T) {
		c := newResetOverlapConfig(t)
		c.Options().SidecarPath = filepath.Join(c.OriginalsPath(), ".sidecar")

		assert.Equal(t, "", resetSidecarOverlap(c))
	})
}

func TestResolvedResetPath(t *testing.T) {
	t.Run("Symlink", func(t *testing.T) {
		dir, err := filepath.EvalSymlinks(t.TempDir())
		require.NoError(t, err)
		link := filepath.Join(t.TempDir(), "link")
		require.NoError(t, os.Symlink(dir, link))

		assert.Equal(t, dir, resolvedResetPath(link))
	})
	t.Run("Missing", func(t *testing.T) {
		assert.Equal(t, "/missing/dir", resolvedResetPath("/missing/dir/"))
	})
}

func TestResetSidecarYaml(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c := config.NewMinimalTestConfig(t.TempDir())
		c.Options().SidecarPath = t.TempDir()
		writeResetTestFiles(t, c.SidecarPath(), "a.yml", "x/b.yml", "x/y/c.yml", "x/y/z/d.yml", "x/y/z/e.json")

		resetSidecarYaml(c)

		assert.Equal(t, []string{"x/y/z/e.json"}, resetTestFiles(t, c.SidecarPath()))
	})
}

func TestResetAlbumYaml(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c := config.NewMinimalTestConfig(t.TempDir())
		dir := c.BackupAlbumsPath()
		writeResetTestFiles(t, dir, "album/a.yml", "folder/x/y/b.yml", "moment/c.txt")

		resetAlbumYaml(c)

		assert.Equal(t, []string{"moment/c.txt"}, resetTestFiles(t, dir))
	})
}
