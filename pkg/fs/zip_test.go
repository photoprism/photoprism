package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZip(t *testing.T) {
	t.Run("Compressed", func(t *testing.T) {
		zipDir := filepath.Join(os.TempDir(), "pkg/fs")
		zipName := filepath.Join(zipDir, "compressed.zip")
		unzipDir := filepath.Join(zipDir, "compressed")
		files := []string{"./testdata/directory/example.jpg"}

		if err := Zip(zipName, files, true); err != nil {
			t.Fatal(err)
		}

		assert.FileExists(t, zipName)

		if info, err := os.Stat(zipName); err != nil {
			t.Error(err)
		} else {
			t.Logf("%s: %d bytes", zipName, info.Size())
		}

		if unzipFiles, err := Unzip(zipName, unzipDir); err != nil {
			t.Error(err)
		} else {
			t.Logf("%s: %#v", zipName, unzipFiles)
		}

		if err := os.Remove(zipName); err != nil {
			t.Fatal(err)
		}

		if err := os.RemoveAll(unzipDir); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("Uncompressed", func(t *testing.T) {
		zipDir := filepath.Join(os.TempDir(), "pkg/fs")
		zipName := filepath.Join(zipDir, "uncompressed.zip")
		unzipDir := filepath.Join(zipDir, "uncompressed")
		files := []string{"./testdata/directory/example.jpg"}

		if err := Zip(zipName, files, false); err != nil {
			t.Fatal(err)
		}

		assert.FileExists(t, zipName)

		if info, err := os.Stat(zipName); err != nil {
			t.Error(err)
		} else {
			t.Logf("%s: %d bytes", zipName, info.Size())
		}

		if unzipFiles, err := Unzip(zipName, unzipDir); err != nil {
			t.Error(err)
		} else {
			t.Logf("%s: %#v", zipName, unzipFiles)
		}

		if err := os.Remove(zipName); err != nil {
			t.Fatal(err)
		}

		if err := os.RemoveAll(unzipDir); err != nil {
			t.Fatal(err)
		}
	})
}
