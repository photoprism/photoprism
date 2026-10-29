package commands

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

// writeThemeZip creates a zip archive with file entries for unzip tests.
func writeThemeZip(t *testing.T, entries map[string]string) string {
	t.Helper()

	zipPath := filepath.Join(t.TempDir(), "theme.zip")
	file, err := os.Create(zipPath) //nolint:gosec // zipPath points to a test temp directory path.
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}

	zipWriter := zip.NewWriter(file)

	for name, content := range entries {
		w, createErr := zipWriter.Create(name)
		if createErr != nil {
			t.Fatalf("create zip entry %q: %v", name, createErr)
		}

		if _, writeErr := w.Write([]byte(content)); writeErr != nil {
			t.Fatalf("write zip entry %q: %v", name, writeErr)
		}
	}

	if closeErr := zipWriter.Close(); closeErr != nil {
		t.Fatalf("close zip writer: %v", closeErr)
	}

	if closeErr := file.Close(); closeErr != nil {
		t.Fatalf("close zip file: %v", closeErr)
	}

	return zipPath
}

func TestUnzipSafe_ValidatesZipEntryPaths(t *testing.T) {
	zipPath := writeThemeZip(t, map[string]string{
		"ok.txt":                "ok\n",
		"nested/good.txt":       "good\n",
		"../outside.txt":        "blocked\n",
		"safe/../../escape.txt": "blocked\n",
		"/absolute.txt":         "blocked\n",
		"C:/drive.txt":          "blocked\n",
		"..\\win-outside.txt":   "blocked\n",
		"nested\\win.txt":       "blocked\n",
		".hidden":               "blocked\n",
		"@x.txt":                "blocked\n",
		"~font.css":             "ok\n",
		"$logo.svg":             "ok\n",
	})

	rootDir := t.TempDir()
	destDir := filepath.Join(rootDir, "theme")
	outsidePath := filepath.Join(rootDir, "outside.txt")
	escapePath := filepath.Join(rootDir, "escape.txt")

	assert.NoError(t, fs.MkdirAll(destDir))
	assert.NoError(t, unzipSafe(zipPath, destDir))

	assert.FileExists(t, filepath.Join(destDir, "ok.txt"))
	assert.FileExists(t, filepath.Join(destDir, "nested", "good.txt"))
	assert.FileExists(t, filepath.Join(destDir, "~font.css"))
	assert.FileExists(t, filepath.Join(destDir, "$logo.svg"))
	assert.NoFileExists(t, filepath.Join(destDir, "@x.txt"))

	_, err := os.Stat(outsidePath)
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat(escapePath)
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat(filepath.Join(destDir, ".hidden"))
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat(filepath.Join(destDir, "nested\\win.txt"))
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat(filepath.Join(destDir, "C:", "drive.txt"))
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestUnzipSafe_CreationMode(t *testing.T) {
	zipPath := filepath.Join(t.TempDir(), "theme.zip")
	file, err := os.Create(zipPath) //nolint:gosec // zipPath points to a test temp directory path.
	require.NoError(t, err)
	zipWriter := zip.NewWriter(file)
	modes := map[string]os.FileMode{"app.js": os.ModeSetuid | os.ModeSetgid | os.ModeSticky | 0o777, "app.css": 0o400}

	for name, mode := range modes {
		header := &zip.FileHeader{Name: name, Method: zip.Store}
		header.SetMode(mode)
		w, createErr := zipWriter.CreateHeader(header)
		require.NoError(t, createErr)
		_, err = w.Write([]byte(name))
		require.NoError(t, err)
	}

	require.NoError(t, zipWriter.Close())
	require.NoError(t, file.Close())

	destDir := t.TempDir()
	require.NoError(t, unzipSafe(zipPath, destDir))

	for name := range modes {
		info, statErr := os.Stat(filepath.Join(destDir, name))
		require.NoError(t, statErr)
		assert.Zero(t, info.Mode()&(os.ModeSetuid|os.ModeSetgid|os.ModeSticky), name)
		assert.Zero(t, info.Mode().Perm()&0o111, name)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm()&0o600, name)
	}
}

func TestUnzipSafe_SizeLimit(t *testing.T) {
	prevFile, prevTotal := themeFileSizeLimit, themeTotalSizeLimit
	t.Cleanup(func() { themeFileSizeLimit, themeTotalSizeLimit = prevFile, prevTotal })
	themeFileSizeLimit, themeTotalSizeLimit = 40, 100

	t.Run("FileAtLimit", func(t *testing.T) {
		destDir := t.TempDir()
		require.NoError(t, unzipSafe(writeThemeZip(t, map[string]string{"app.css": strings.Repeat("a", 40)}), destDir))
		assert.FileExists(t, filepath.Join(destDir, "app.css"))
	})
	t.Run("FileOverLimit", func(t *testing.T) {
		destDir := t.TempDir()
		require.Error(t, unzipSafe(writeThemeZip(t, map[string]string{"app.css": strings.Repeat("a", 41)}), destDir))
		assert.NoFileExists(t, filepath.Join(destDir, "app.css"))
	})
	t.Run("TotalAtLimit", func(t *testing.T) {
		files := map[string]string{"a.css": strings.Repeat("a", 40), "b.css": strings.Repeat("b", 40), "c.css": strings.Repeat("c", 20)}
		require.NoError(t, unzipSafe(writeThemeZip(t, files), t.TempDir()))
	})
	t.Run("TotalOverLimit", func(t *testing.T) {
		files := map[string]string{"a.css": strings.Repeat("a", 40), "b.css": strings.Repeat("b", 40), "c.css": strings.Repeat("c", 21)}
		require.Error(t, unzipSafe(writeThemeZip(t, files), t.TempDir()))
	})
}
