package commands

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

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
	})

	rootDir := t.TempDir()
	destDir := filepath.Join(rootDir, "theme")
	outsidePath := filepath.Join(rootDir, "outside.txt")
	escapePath := filepath.Join(rootDir, "escape.txt")

	assert.NoError(t, fs.MkdirAll(destDir))
	assert.NoError(t, unzipSafe(zipPath, destDir))

	assert.FileExists(t, filepath.Join(destDir, "ok.txt"))
	assert.FileExists(t, filepath.Join(destDir, "nested", "good.txt"))

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
