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
	t.Run("CaseVariant", func(t *testing.T) {
		// A case variant of the originals path is refused whatever the case mode, since it may name the same folder.
		variant := filepath.Join(filepath.Dir(originals), "ORIGINALS", "sub")
		mode := fs.GetCaseMode()
		t.Cleanup(func() { fs.RestoreCaseMode(mode) })

		for _, ignore := range []bool{false, true} {
			fs.SetCaseScopes(fs.CaseScope{Dir: originals, Ignore: ignore})
			assert.ErrorIs(t, imp.CheckPath(variant), ErrImportPath)
			assert.ErrorIs(t, imp.CheckPath(filepath.Dir(strings.ToUpper(base))), ErrImportPath)
			assert.True(t, imp.linksToOriginals(variant))
		}

		// A storage folder whose name has capital letters is still found on a case-sensitive file system.
		storage := filepath.Join(originals, "Storage")
		require.NoError(t, fs.MkdirAll(filepath.Join(storage, "users", "upload")))
		require.NoError(t, os.WriteFile(filepath.Join(storage, fs.PPStorageFilename), []byte("1"), fs.ModeFile))
		t.Cleanup(func() { _ = os.RemoveAll(storage) })
		assert.NoError(t, imp.CheckPath(filepath.Join(storage, "users", "upload")))
		assert.NoError(t, imp.CheckPath(filepath.Join(filepath.Dir(originals), "ORIGINALS", "Storage", "users", "upload")))
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

func TestImport_Check(t *testing.T) {
	cfg := config.NewMinimalTestConfig(t.TempDir())
	originals := t.TempDir()
	cfg.Options().OriginalsPath = originals
	imp := NewImport(cfg, nil, nil)

	assert.NoError(t, imp.Check(ImportOptionsCopy(t.TempDir(), "")))
	assert.Equal(t, ErrImportPath, imp.Check(ImportOptionsCopy(filepath.Join(originals, "sub"), "")))
	assert.Equal(t, ErrImportDest, imp.Check(ImportOptionsMove(t.TempDir(), "a:b")))
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
	assert.False(t, imp.linksToOriginals(""))
}

func TestOverlappingPaths(t *testing.T) {
	dir := t.TempDir()
	assert.True(t, overlappingPaths(dir, filepath.Join(dir, "a")))
	assert.True(t, overlappingPaths(filepath.Join(dir, "a"), dir))
	assert.True(t, overlappingPaths(dir, dir))
	assert.False(t, overlappingPaths(filepath.Join(dir, "a"), filepath.Join(dir, "b")))

	// Letters are compared case-insensitively.
	assert.True(t, overlappingPaths(strings.ToUpper(dir), filepath.Join(dir, "a")))
	assert.True(t, overlappingPaths(filepath.Join(dir, "a"), strings.ToUpper(dir)))

	// An empty path is not resolved to the working directory.
	t.Chdir(dir)
	assert.False(t, overlappingPaths("", dir))
	assert.False(t, overlappingPaths(filepath.Join(dir, "a"), ""))
}

func TestRelBelow(t *testing.T) {
	rel, ok := relBelow("/a/b", "/a/b/c/d", false)
	assert.True(t, ok)
	assert.Equal(t, filepath.Join("c", "d"), rel)
	rel, ok = relBelow("/a/b", "/a/b", false)
	assert.True(t, ok)
	assert.Equal(t, ".", rel)

	// With fold, letters are compared case-insensitively and the relative path keeps the case of folder.
	rel, ok = relBelow("/A/b", "/a/B/Cd/E", true)
	assert.True(t, ok)
	assert.Equal(t, filepath.Join("Cd", "E"), rel)
	rel, ok = relBelow("/A/b/", "/a/B", true)
	assert.True(t, ok)
	assert.Equal(t, ".", rel)
	rel, ok = relBelow("/", "/A/b", true)
	assert.True(t, ok)
	assert.Equal(t, filepath.Join("A", "b"), rel)

	for _, folder := range []string{"/a", "/a/bc", "/x", "/A/BC"} {
		_, ok = relBelow("/a/b", folder, false)
		assert.False(t, ok, folder)
		_, ok = relBelow("/a/b", folder, true)
		assert.False(t, ok, folder)
	}

	// Case variants may differ in length, e.g. the Kelvin sign, and names that are not valid UTF-8 fold ASCII only.
	rel, ok = relBelow("/data/\u212a", "/data/k/sub", true)
	assert.True(t, ok)
	assert.Equal(t, "sub", rel)
	rel, ok = relBelow("/CAF\xe9", "/caf\xe9/x", true)
	assert.True(t, ok)
	assert.Equal(t, "x", rel)
	_, ok = relBelow("/caf\xe9", "/caf\xe8/x", true)
	assert.False(t, ok)
}

func TestPathNames(t *testing.T) {
	assert.Equal(t, []string{""}, pathNames("/"))
	assert.Equal(t, []string{"", "a", "b"}, pathNames("/a/b/"))
	assert.Equal(t, []string{"a", "b"}, pathNames("a//b"))
	assert.Equal(t, []string{"."}, pathNames(""))
}

func TestSameNameFold(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.True(t, sameNameFold("Photos", "PHOTOS"))
		assert.True(t, sameNameFold("\u212a", "k"))
		assert.True(t, sameNameFold("Stra\u00dfe", "STRA\u1e9eE"))
		assert.False(t, sameNameFold("Photos", "Photo"))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.True(t, sameNameFold("CAF\xe9", "caf\xe9"))
		assert.False(t, sameNameFold("caf\xe9", "caf\xe8"))
		assert.False(t, sameNameFold("caf\xe9", "caf\xe9x"))
		assert.False(t, sameNameFold("caf\xe9", "café"))
	})
}

func TestLowerASCII(t *testing.T) {
	assert.Equal(t, byte('a'), lowerASCII('A'))
	assert.Equal(t, byte('z'), lowerASCII('Z'))
	assert.Equal(t, byte('a'), lowerASCII('a'))
	assert.Equal(t, byte('@'), lowerASCII('@'))
	assert.Equal(t, byte(0xc9), lowerASCII(0xc9))
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
	// The temporary folder may itself be below a link, e.g. /var on macOS.
	dir, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
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
