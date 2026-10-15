package photoprism

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestImport_StartUnportableNames checks that the import leaves files and folders whose names are not portable in place.
func TestImport_StartUnportableNames(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := config.TestConfig()
	src := t.TempDir()

	portable := []string{"ok.jpg", "03 - Março/#1 & Friends [2020] (copy) Tom's.jpg", "pair.jpg", "stack.jpg", "日本\u3000旅行/😀.jpg"}
	skipped := []string{"a|b.jpg", "x:y.jpg", "dir*/inner.jpg", "caf\xe9.jpg", "~home.jpg", "a$b.jpg", "pair.x|y.xmp", "stack.x|y.gif"}

	for i, name := range portable {
		writeTestJpeg(t, filepath.Join(src, name), 200+i)
	}

	for i, name := range skipped {
		switch filepath.Ext(name) {
		case ".xmp":
			require.NoError(t, os.WriteFile(filepath.Join(src, name), []byte(testXmp), fs.ModeFile))
		case ".gif":
			writeTestGif(t, filepath.Join(src, name))
		default:
			writeTestJpeg(t, filepath.Join(src, name), 220+i)
		}
	}

	// An empty folder whose name is not portable stays as well.
	require.NoError(t, os.MkdirAll(filepath.Join(src, "x:y"), fs.ModeDir))

	hook := captureLog(t)
	convert := NewConvert(cfg)
	imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)

	imp.Start(ImportOptionsMove(src, ""))
	done := imp.Start(ImportOptionsMove(src, ""))

	for _, name := range []string{"a|b.jpg", "x:y.jpg", "caf\xe9.jpg", "~home.jpg"} {
		assert.False(t, done[filepath.Join(src, name)].Exists(), "%q", name)
	}

	for _, name := range portable {
		assert.False(t, fs.FileExists(filepath.Join(src, name)), "%q was not imported", name)
	}

	for _, name := range skipped {
		assert.True(t, fs.FileExists(filepath.Join(src, name)), "%q was imported", name)
	}

	assert.DirExists(t, filepath.Join(src, "x:y"))

	warnings := loggedMessages(hook, logrus.WarnLevel)

	for _, rel := range []string{"a|b.jpg", "x:y.jpg", "dir*", "caf\xe9.jpg", "~home.jpg", "a$b.jpg", "pair.x|y.xmp", "stack.x|y.gif"} {
		count := 0

		for _, msg := range warnings {
			if strings.HasPrefix(msg, "import: skipped "+clean.Log(rel)+" ") {
				count++
			}
		}

		assert.Equal(t, 1, count, "%q in %v", rel, warnings)
	}
}

// TestImport_StartUnportableLimit checks that names above the warning limit are counted in one warning per run.
func TestImport_StartUnportableLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	saved := unportablePaths
	t.Cleanup(func() { unportablePaths = saved })
	unportablePaths = &warnedPaths{limit: 1}

	cfg := config.TestConfig()
	src := t.TempDir()

	for i, name := range []string{"a|b.jpg", "c|d.jpg", "e$f.jpg"} {
		writeTestJpeg(t, filepath.Join(src, name), 260+i)
	}

	hook := captureLog(t)
	convert := NewConvert(cfg)
	imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)

	imp.Start(ImportOptionsMove(src, ""))
	imp.Start(ImportOptionsMove(src, ""))

	listed, counted := 0, 0

	for _, msg := range loggedMessages(hook, logrus.WarnLevel) {
		switch {
		case strings.HasPrefix(msg, "import: skipped") && strings.HasSuffix(msg, "rename it to import it"):
			listed++
		case strings.HasPrefix(msg, "import: skipped 2 more names"):
			counted++
		}
	}

	assert.Equal(t, 1, listed)
	assert.Equal(t, 2, counted)

	// Names the import workers count for the index are reported by the same run.
	warnUnportable("index", entity.RootOriginals, "import-limit/x|y.jpg", "index it")
	hook.Reset()
	imp.Start(ImportOptionsMove(t.TempDir(), ""))

	indexed := 0

	for _, msg := range loggedMessages(hook, logrus.WarnLevel) {
		if strings.HasPrefix(msg, "index: skipped 1 more name ") {
			indexed++
		}
	}

	assert.Equal(t, 1, indexed)
}

// TestImport_StartUnportableFolder checks the names of folders below the import folder and of folders symbolic links point to.
func TestImport_StartUnportableFolder(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := config.TestConfig()
	convert := NewConvert(cfg)
	imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)

	t.Run("Subfolder", func(t *testing.T) {
		// A folder below the import folder whose name is not portable is not imported, also if the import starts in it.
		dir := filepath.Join(cfg.ImportPath(), "unportable-"+rnd.Base36(8)+":x")
		t.Cleanup(func() { _ = os.RemoveAll(dir) })
		writeTestJpeg(t, filepath.Join(dir, "a.jpg"), 240)

		hook := captureLog(t)
		imp.Start(ImportOptionsMove(dir, ""))

		assert.True(t, fs.FileExists(filepath.Join(dir, "a.jpg")))
		assert.Contains(t, strings.Join(loggedMessages(hook, logrus.WarnLevel), "\n"),
			"import: skipped "+clean.Log(filepath.Base(dir))+" ")
	})
	t.Run("Links", func(t *testing.T) {
		src := t.TempDir()
		writeTestJpeg(t, filepath.Join(src, "real", "x.jpg"), 241)
		require.NoError(t, os.Symlink("real", filepath.Join(src, "A:link")))

		imp.Start(ImportOptionsMove(src, ""))

		assert.False(t, fs.FileExists(filepath.Join(src, "real", "x.jpg")), "real/x.jpg was not imported")
	})
}

// TestImport_DestinationFilenameUnportable checks that a file is not restored to a recorded name that is not portable.
func TestImport_DestinationFilenameUnportable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := config.TestConfig()
	dir := "unportable-" + rnd.Base36(8)
	root := filepath.Join(cfg.OriginalsPath(), dir)
	t.Cleanup(func() { _ = os.RemoveAll(root) })
	cleanupIndexed(t, dir)

	// The file is indexed directly under a name that is not portable, then removed.
	fileName := filepath.Join(root, "old|name.jpg")
	writeTestJpeg(t, fileName, 250)
	convert := NewConvert(cfg)
	ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())
	require.NoError(t, ind.FileName(fileName, NewIndexOptions(dir, false, false, false, false, true, cfg)).Err)
	require.True(t, indexedFile(t, dir+"/old|name.jpg"))

	src := filepath.Join(t.TempDir(), "copy.jpg")
	writeTestJpeg(t, src, 250)
	require.NoError(t, os.Remove(fileName))

	mf, err := NewMediaFile(src)
	require.NoError(t, err)

	imp := NewImport(cfg, ind, convert)
	dest, err := imp.DestinationFilename(mf, mf, "")
	require.NoError(t, err)
	assert.NotEqual(t, fileName, dest)
	assert.True(t, clean.PortableName(filepath.Base(dest)), dest)
}

// TestImport_StartUnportableDest checks that an imported file is indexed under the name its destination pattern gives it.
func TestImport_StartUnportableDest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := config.TestConfig()
	dest := cfg.Settings().Import.Dest
	t.Cleanup(func() { cfg.Settings().Import.Dest = dest })
	cfg.Settings().Import.Dest = "2006/01/2006-01-02T15:04:05_82F63B78.jpg"

	// The imported files are removed by the pattern of their names, also if a check below fails.
	t.Cleanup(func() {
		var files []entity.File
		_ = entity.UnscopedDb().Where("file_root = ? AND file_name LIKE ?", entity.RootOriginals, "%T%:%").Find(&files).Error
		for _, f := range files {
			_ = os.Remove(filepath.Join(cfg.OriginalsPath(), f.FileName))
			_ = entity.UnscopedDb().Delete(&entity.File{}, f.ID).Error
			_ = entity.UnscopedDb().Delete(&entity.Photo{}, f.PhotoID).Error
		}
	})

	src := t.TempDir()
	writeTestJpeg(t, filepath.Join(src, "colon.jpg"), 251)

	convert := NewConvert(cfg)
	imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)
	imp.Start(ImportOptionsMove(src, ""))

	require.False(t, fs.FileExists(filepath.Join(src, "colon.jpg")), "colon.jpg was not imported")

	var count int64
	require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).Where("file_root = ? AND file_name LIKE ?", entity.RootOriginals, "%T%:%").Count(&count).Error)
	assert.Positive(t, count)
}
