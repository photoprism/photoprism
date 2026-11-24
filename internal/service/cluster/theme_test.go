package cluster

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// themeZipEntry describes an entry of a test theme archive.
type themeZipEntry struct {
	name, content string
	mode          os.FileMode
}

// writeThemeZip creates a theme archive with the specified entries, in order.
func writeThemeZip(t *testing.T, entries ...themeZipEntry) string {
	t.Helper()

	zipPath := filepath.Join(t.TempDir(), "theme.zip")
	file, err := os.Create(zipPath) //nolint:gosec // test temp directory
	require.NoError(t, err)

	zw := zip.NewWriter(file)

	for _, e := range entries {
		header := &zip.FileHeader{Name: e.name, Method: zip.Store}

		if e.mode != 0 {
			header.SetMode(e.mode)
		}

		w, createErr := zw.CreateHeader(header)
		require.NoError(t, createErr)
		_, err = w.Write([]byte(e.content))
		require.NoError(t, err)
	}

	require.NoError(t, zw.Close())
	require.NoError(t, file.Close())

	return zipPath
}

// themeFile returns a regular theme archive entry.
func themeFile(name, content string) themeZipEntry {
	return themeZipEntry{name: name, content: content}
}

// useThemeLimits sets the theme archive bounds for the duration of a test.
func useThemeLimits(t *testing.T, fileLimit, totalLimit int64) {
	t.Helper()

	prevFile, prevTotal := themeFileSizeLimit, themeTotalSizeLimit
	t.Cleanup(func() { themeFileSizeLimit, themeTotalSizeLimit = prevFile, prevTotal })
	themeFileSizeLimit, themeTotalSizeLimit = fileLimit, totalLimit
}

func TestExtractTheme(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		dir := t.TempDir()
		zipPath := writeThemeZip(t,
			themeFile("app.js", "js"),
			themeFile("nested/good.css", "css"),
			themeFile("~font.css", "css"),
			themeFile("$logo.svg", "svg"),
		)

		require.NoError(t, ExtractTheme(zipPath, dir))
		assert.FileExists(t, filepath.Join(dir, "app.js"))
		assert.FileExists(t, filepath.Join(dir, "nested", "good.css"))
		assert.FileExists(t, filepath.Join(dir, "~font.css"))
		assert.FileExists(t, filepath.Join(dir, "$logo.svg"))
	})
	t.Run("HiddenAndSystemNamesAreSkipped", func(t *testing.T) {
		dir := t.TempDir()
		zipPath := writeThemeZip(t,
			themeFile("app.js", "js"),
			themeFile(".hidden", "x"),
			themeFile("@eaDir/thumb.jpg", "x"),
			themeFile("@x.txt", "x"),
			themeFile("__MACOSX/app.js", "x"),
			themeFile("nested/._good.css", "x"),
			themeFile(".git/config.css", "x"),
			themeFile("css\\.hidden.css", "x"),
			themeFile("./main.css", "css"),
		)

		require.NoError(t, ExtractTheme(zipPath, dir))
		assert.FileExists(t, filepath.Join(dir, "app.js"))
		assert.FileExists(t, filepath.Join(dir, "main.css"))
		assert.NoFileExists(t, filepath.Join(dir, ".hidden"))
		assert.NoFileExists(t, filepath.Join(dir, "@x.txt"))
		assert.NoFileExists(t, filepath.Join(dir, "nested", "._good.css"))
		assert.NoDirExists(t, filepath.Join(dir, "@eaDir"))
		assert.NoDirExists(t, filepath.Join(dir, ".git"))
		assert.NoDirExists(t, filepath.Join(dir, "css"))
		assert.NoDirExists(t, filepath.Join(dir, "__MACOSX"))
	})
	t.Run("CreationMode", func(t *testing.T) {
		dir := t.TempDir()
		zipPath := writeThemeZip(t,
			themeZipEntry{name: "app.js", content: "js", mode: os.ModeSetuid | os.ModeSetgid | os.ModeSticky | 0o777},
			themeZipEntry{name: "app.css", content: "css", mode: 0o400},
		)

		require.NoError(t, ExtractTheme(zipPath, dir))

		for _, name := range []string{"app.js", "app.css"} {
			info, err := os.Stat(filepath.Join(dir, name))
			require.NoError(t, err)
			assert.Zero(t, info.Mode()&(os.ModeSetuid|os.ModeSetgid|os.ModeSticky), name)
			assert.Zero(t, info.Mode().Perm()&0o111, name)
			assert.Equal(t, os.FileMode(0o600), info.Mode().Perm()&0o600, name)
		}
	})
	t.Run("UnsafeEntriesFail", func(t *testing.T) {
		for _, name := range []string{"../outside.txt", "safe/../../escape.txt", "/absolute.txt", "C:/drive.txt", "..\\win-outside.txt", "fonts/a..b.css"} {
			t.Run(name, func(t *testing.T) {
				root := t.TempDir()
				dir := filepath.Join(root, "theme")
				require.NoError(t, os.Mkdir(dir, 0o700))

				err := ExtractTheme(writeThemeZip(t, themeFile("app.js", "js"), themeFile(name, "x")), dir)

				require.Error(t, err)
				assert.NoFileExists(t, filepath.Join(dir, "app.js"))
				assert.NoFileExists(t, filepath.Join(root, "outside.txt"))
				assert.NoFileExists(t, filepath.Join(root, "escape.txt"))
			})
		}
	})
	t.Run("NestedFilesAreRemoved", func(t *testing.T) {
		// Implicit parent directories may remain, but no extracted file does.
		dir := t.TempDir()
		zipPath := writeThemeZip(t, themeFile("app.js", "js"), themeFile("css/nested/app.css", "css"), themeFile("../outside.txt", "x"))

		require.Error(t, ExtractTheme(zipPath, dir))
		assert.NoFileExists(t, filepath.Join(dir, "app.js"))
		assert.NoFileExists(t, filepath.Join(dir, "css", "nested", "app.css"))
	})
	t.Run("SymlinkFails", func(t *testing.T) {
		dir := t.TempDir()
		zipPath := writeThemeZip(t, themeFile("app.js", "js"), themeZipEntry{name: "link.css", content: "/etc/passwd", mode: os.ModeSymlink | 0o777})

		require.Error(t, ExtractTheme(zipPath, dir))
		assert.NoFileExists(t, filepath.Join(dir, "app.js"))
		assert.NoFileExists(t, filepath.Join(dir, "link.css"))
	})
	t.Run("BackslashNameUsesSlash", func(t *testing.T) {
		dir := t.TempDir()

		require.NoError(t, ExtractTheme(writeThemeZip(t, themeFile("nested\\win.css", "css")), dir))
		assert.FileExists(t, filepath.Join(dir, "nested", "win.css"))
	})
	t.Run("FileAtLimit", func(t *testing.T) {
		useThemeLimits(t, 40, 100)
		dir := t.TempDir()

		require.NoError(t, ExtractTheme(writeThemeZip(t, themeFile("app.css", strings.Repeat("a", 40))), dir))
		assert.FileExists(t, filepath.Join(dir, "app.css"))
	})
	t.Run("FileOverLimit", func(t *testing.T) {
		useThemeLimits(t, 40, 100)
		dir := t.TempDir()
		zipPath := writeThemeZip(t, themeFile("app.js", "js"), themeFile("app.css", strings.Repeat("a", 41)))

		require.Error(t, ExtractTheme(zipPath, dir))
		assert.NoFileExists(t, filepath.Join(dir, "app.js"))
		assert.NoFileExists(t, filepath.Join(dir, "app.css"))
	})
	t.Run("TotalBelowLimit", func(t *testing.T) {
		useThemeLimits(t, 40, 100)
		zipPath := writeThemeZip(t, themeFile("a.css", strings.Repeat("a", 40)), themeFile("b.css", strings.Repeat("b", 40)), themeFile("c.css", strings.Repeat("c", 19)))

		require.NoError(t, ExtractTheme(zipPath, t.TempDir()))
	})
	t.Run("TotalAtLimitFails", func(t *testing.T) {
		// fs.Unzip requires each file to leave at least one byte of the total budget.
		useThemeLimits(t, 40, 100)
		dir := t.TempDir()
		zipPath := writeThemeZip(t, themeFile("app.js", strings.Repeat("a", 40)), themeFile("b.css", strings.Repeat("b", 40)), themeFile("c.css", strings.Repeat("c", 20)))

		require.Error(t, ExtractTheme(zipPath, dir))
		assert.NoFileExists(t, filepath.Join(dir, "app.js"))
		assert.NoFileExists(t, filepath.Join(dir, "b.css"))
	})
	t.Run("MissingArchive", func(t *testing.T) {
		require.Error(t, ExtractTheme(filepath.Join(t.TempDir(), "missing.zip"), t.TempDir()))
	})
}

func TestThemeNameHidden(t *testing.T) {
	t.Run("Visible", func(t *testing.T) {
		for _, name := range []string{"app.js", "css/app.css", "./main.css", "fonts/", "~font.css", "$logo.svg", "nested\\win.css"} {
			assert.False(t, themeNameHidden(name), name)
		}
	})
	t.Run("Hidden", func(t *testing.T) {
		for _, name := range []string{".hidden", "@x.txt", "@eaDir/thumb.jpg", ".git/config", "nested/._good.css", "__MACOSX/app.js", "css\\.hidden.css", ".well-known/"} {
			assert.True(t, themeNameHidden(name), name)
		}
	})
}
