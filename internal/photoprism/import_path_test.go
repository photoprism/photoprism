package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestImport_CheckPath(t *testing.T) {
	cfg := config.NewMinimalTestConfig(t.TempDir())
	base := t.TempDir()
	originals := filepath.Join(base, "originals")
	require.NoError(t, fs.MkdirAll(filepath.Join(originals, "sub")))
	require.NoError(t, fs.MkdirAll(filepath.Join(base, "originals2")))
	require.NoError(t, os.Symlink(filepath.Join(originals, "sub"), filepath.Join(base, "link")))
	require.NoError(t, os.Symlink(originals, filepath.Join(base, "originals-link")))
	cfg.Options().OriginalsPath = originals
	imp := NewImport(cfg, nil, nil)

	t.Run("Outside", func(t *testing.T) {
		for _, dir := range []string{"", filepath.Join(base, "originals2"), filepath.Join(base, "import"),
			filepath.Join(originals, ".photoprism", "storage", "upload"), filepath.Join(originals, "sub", "@eaDir")} {
			assert.NoError(t, imp.CheckPath(dir), dir)
		}
	})
	t.Run("StorageFolder", func(t *testing.T) {
		// A storage folder in originals is marked with a file that walks skip.
		storage := filepath.Join(originals, "storage")
		require.NoError(t, fs.MkdirAll(filepath.Join(storage, "users", "upload")))
		assert.ErrorIs(t, imp.CheckPath(filepath.Join(storage, "users", "upload")), ErrImportPath)
		require.NoError(t, os.WriteFile(filepath.Join(storage, fs.PPStorageFilename), []byte("1"), fs.ModeFile))
		t.Cleanup(func() { _ = os.RemoveAll(storage) })
		assert.NoError(t, imp.CheckPath(filepath.Join(storage, "users", "upload")))
	})
	t.Run("Contains", func(t *testing.T) {
		for _, dir := range []string{base, filepath.Dir(base)} {
			assert.ErrorIs(t, imp.CheckPath(dir), ErrImportPath, dir)
		}
	})
	t.Run("HiddenOriginals", func(t *testing.T) {
		hidden := filepath.Join(base, "import", ".originals")
		require.NoError(t, fs.MkdirAll(hidden))
		cfg.Options().OriginalsPath = hidden
		t.Cleanup(func() { cfg.Options().OriginalsPath = originals })
		assert.NoError(t, imp.CheckPath(filepath.Join(base, "import")))
		assert.ErrorIs(t, imp.CheckPath(filepath.Join(hidden, "x")), ErrImportPath)
	})
	t.Run("Inside", func(t *testing.T) {
		for _, dir := range []string{originals, originals + "/", filepath.Join(originals, "sub"), originals + "/../originals/sub",
			filepath.Join(originals, "missing"), filepath.Join(base, "link"), filepath.Join(base, "originals-link", "sub")} {
			err := imp.CheckPath(dir)
			assert.ErrorIs(t, err, ErrImportPath, dir)
			assert.ErrorContains(t, err, "import and originals folders must not overlap (import ", dir)
			assert.ErrorContains(t, err, ", originals "+originals+")", dir)
		}
	})
	t.Run("OriginalsLink", func(t *testing.T) {
		cfg.Options().OriginalsPath = filepath.Join(base, "originals-link")
		t.Cleanup(func() { cfg.Options().OriginalsPath = originals })
		assert.ErrorIs(t, imp.CheckPath(filepath.Join(originals, "sub")), ErrImportPath)
		assert.NoError(t, imp.CheckPath(filepath.Join(base, "originals2")))
	})
}

func TestImportError(t *testing.T) {
	err := importError{"destination folder x is not supported", ErrImportDest}
	assert.EqualError(t, err, "destination folder x is not supported")
	assert.ErrorIs(t, err, ErrImportDest)
	assert.NotErrorIs(t, err, ErrImportPath)
}

func TestImport_linksToOriginals(t *testing.T) {
	cfg := config.NewMinimalTestConfig(t.TempDir())
	base := t.TempDir()
	originals := filepath.Join(base, "originals")
	require.NoError(t, fs.MkdirAll(filepath.Join(originals, ".photoprism")))
	cfg.Options().OriginalsPath = originals
	imp := NewImport(cfg, nil, nil)

	for _, target := range []string{originals, filepath.Join(originals, ".photoprism"), base} {
		assert.True(t, imp.linksToOriginals(target), target)
	}

	assert.False(t, imp.linksToOriginals(t.TempDir()))
}

func TestRelBelow(t *testing.T) {
	rel, ok := relBelow("/a/b", "/a/b/c/d")
	assert.True(t, ok)
	assert.Equal(t, filepath.Join("c", "d"), rel)
	rel, ok = relBelow("/a/b", "/a/b")
	assert.True(t, ok)
	assert.Equal(t, ".", rel)

	for _, folder := range []string{"/a", "/a/bc", "/x"} {
		_, ok = relBelow("/a/b", folder)
		assert.False(t, ok, folder)
	}
}

func TestWalkSkips(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, fs.MkdirAll(filepath.Join(dir, "storage", "users")))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "storage", fs.PPStorageFilename), []byte("1"), fs.ModeFile))

	assert.False(t, walkSkips(dir, "."))
	assert.False(t, walkSkips(dir, filepath.Join("2019", "x")))
	assert.True(t, walkSkips(dir, filepath.Join(".photoprism", "storage")))
	assert.True(t, walkSkips(dir, filepath.Join("2019", "@eaDir")))
	assert.True(t, walkSkips(dir, "storage"))
	assert.True(t, walkSkips(dir, filepath.Join("storage", "users", "upload")))
}

func TestResolvedPath(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Symlink(dir, filepath.Join(dir, "link")))
	assert.Equal(t, dir, resolvedPath(filepath.Join(dir, "link")))
	assert.Equal(t, filepath.Join(dir, "missing"), resolvedPath(dir+"/x/../missing/"))
}

// TestImport_RunPathInOriginals checks that moving and copying from a folder in originals is refused.
func TestImport_RunPathInOriginals(t *testing.T) {
	cfg := config.TestConfig()
	src := filepath.Join(cfg.OriginalsPath(), "import-in-originals-"+rnd.Base36(8))
	writeTestJpeg(t, filepath.Join(src, "ok.jpg"), 280)
	t.Cleanup(func() { _ = os.RemoveAll(src) })

	convert := NewConvert(cfg)
	imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)

	t.Run("Subfolder", func(t *testing.T) {
		for _, opt := range []ImportOptions{ImportOptionsMove(src, ""), ImportOptionsCopy(src, "")} {
			_, err := imp.Run(opt)
			assert.Equal(t, ErrImportPath, err)
			assert.FileExists(t, filepath.Join(src, "ok.jpg"))
		}
	})
	t.Run("LinkedFolder", func(t *testing.T) {
		// A linked folder that leads to originals is skipped, so a move import does not remove what it holds.
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		importDir := t.TempDir()
		require.NoError(t, os.Symlink(src, filepath.Join(importDir, "linked")))
		hook := captureLog(t)

		_, err := imp.Run(ImportOptionsMove(importDir, ""))
		require.NoError(t, err)
		assert.FileExists(t, filepath.Join(src, "ok.jpg"))
		assert.Equal(t, 1, countPrefix(loggedMessages(hook, logrus.WarnLevel), "import: skipped linked because it links to the originals folder"))
	})
	t.Run("ImportFolder", func(t *testing.T) {
		saved := cfg.Options().ImportPath
		t.Cleanup(func() { cfg.Options().ImportPath = saved })
		cfg.Options().ImportPath = src

		_, err := imp.Run(ImportOptionsMove(cfg.ImportPath(), ""))
		assert.Equal(t, ErrImportPath, err)
		assert.FileExists(t, filepath.Join(src, "ok.jpg"))
	})
}
