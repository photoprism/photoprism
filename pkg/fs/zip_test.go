package fs

import (
	"archive/zip"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func writeZip(t *testing.T, path string, entries map[string][]byte) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)

	for name, data := range entries {
		hdr := &zip.FileHeader{Name: name, Method: zip.Store}
		w, createErr := zw.CreateHeader(hdr)
		if createErr != nil {
			t.Fatal(createErr)
		}
		if _, writeErr := w.Write(data); writeErr != nil {
			t.Fatal(writeErr)
		}
	}
	assert.NoError(t, zw.Close())
}

func TestUnzip_SkipRulesAndLimits(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "test.zip")

	entries := map[string][]byte{
		"__MACOSX/._junk": []byte("meta"),  // skipped by prefix
		"ok1.txt":         []byte("abc"),   // 3 bytes
		"dir/../evil.txt": []byte("pwned"), // skipped due to ..
		"ok2.txt":         []byte("x"),     // 1 byte
	}
	writeZip(t, zipPath, entries)

	// totalSizeLimit == 0 → skip everything
	files, skipped, err := Unzip(zipPath, filepath.Join(dir, "a"), 0, 0)
	assert.NoError(t, err)
	assert.Empty(t, files)
	assert.GreaterOrEqual(t, len(skipped), 1)

	// Apply per-file and total limits
	outDir := filepath.Join(dir, "b")
	files, skipped, err = Unzip(zipPath, outDir, 2, 3) // file limit=2 bytes; total limit=3 bytes
	assert.NoError(t, err)

	// ok1 (3 bytes) skipped by file limit; evil skipped by '..'; __MACOSX skipped by prefix
	// ok2 (1 byte) allowed; total limit reduces to 2; nothing else left that fits
	assert.ElementsMatch(t, []string{filepath.Join(outDir, "ok2.txt")}, files)
	// Ensure file written
	b, rerr := os.ReadFile(filepath.Join(outDir, "ok2.txt"))
	assert.NoError(t, rerr)
	assert.Equal(t, []byte("x"), b)
	// Skipped contains at least the three excluded entries
	assert.GreaterOrEqual(t, len(skipped), 3)
}

func TestUnzip_AbsolutePathRejected(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "abs.zip")
	absName := string(os.PathSeparator) + filepath.Join("tmp", "abs.txt")
	entries := map[string][]byte{absName: []byte("bad")}
	writeZip(t, zipPath, entries)

	_, _, err := Unzip(zipPath, filepath.Join(dir, "out"), 0, 10)
	if err == nil {
		t.Fatalf("expected error for absolute path entry")
	}
}

func TestUnzip_WindowsVolumePathRejected(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("volume path semantics only apply on Windows")
	}
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "vol.zip")
	entries := map[string][]byte{"C:/Windows/System32/evil.txt": []byte("bad")}
	writeZip(t, zipPath, entries)

	_, _, err := Unzip(zipPath, filepath.Join(dir, "out"), 0, 10)
	if err == nil {
		t.Fatalf("expected error for volume path entry on Windows")
	}
}

func TestUnzip_WindowsBackslashVolumePathRejected(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("volume path semantics only apply on Windows")
	}
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "vol_bs.zip")
	entries := map[string][]byte{"C:\\Windows\\System32\\evil.txt": []byte("bad")}
	writeZip(t, zipPath, entries)

	_, _, err := Unzip(zipPath, filepath.Join(dir, "out"), 0, 10)
	if err == nil {
		t.Fatalf("expected error for backslash volume path entry on Windows")
	}
}

func TestUnzip_CreatesDirectoriesAndNestedFiles(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "nested.zip")
	entries := map[string][]byte{
		"nested/":          nil, // directory entry
		"nested/a.txt":     []byte("A"),
		"nested/sub/":      nil, // nested dir entry
		"nested/sub/b.txt": []byte("BB"),
	}
	writeZip(t, zipPath, entries)

	outDir := filepath.Join(dir, "out")
	files, skipped, err := Unzip(zipPath, outDir, 10, 100)
	if err != nil {
		t.Fatal(err)
	}
	// Expect both files extracted; directories may also be included in the returned list.
	expectedA := filepath.Join(outDir, "nested/a.txt")
	expectedB := filepath.Join(outDir, "nested/sub/b.txt")
	m := map[string]bool{}
	for _, f := range files {
		m[f] = true
	}
	if !m[expectedA] || !m[expectedB] {
		t.Fatalf("extracted list missing expected files: %v", files)
	}
	if len(skipped) != 0 {
		t.Fatalf("unexpected skipped: %v", skipped)
	}
	// Check directories exist
	if fi, err := os.Stat(filepath.Join(outDir, "nested")); err != nil || !fi.IsDir() {
		t.Fatalf("nested dir missing")
	}
	if fi, err := os.Stat(filepath.Join(outDir, "nested/sub")); err != nil || !fi.IsDir() {
		t.Fatalf("nested subdir missing")
	}
}

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

		if unzipFiles, skippedFiles, err := Unzip(zipName, unzipDir, 2*GB, -1); err != nil {
			t.Error(err)
		} else {
			t.Logf("%s: extracted %#v", zipName, unzipFiles)
			t.Logf("%s: skipped %#v", zipName, skippedFiles)
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

		if unzipFiles, skippedFiles, err := Unzip(zipName, unzipDir, 2*GB, -1); err != nil {
			t.Error(err)
		} else {
			t.Logf("%s: extracted %#v", zipName, unzipFiles)
			t.Logf("%s: skipped %#v", zipName, skippedFiles)
		}

		if err := os.Remove(zipName); err != nil {
			t.Fatal(err)
		}

		if err := os.RemoveAll(unzipDir); err != nil {
			t.Fatal(err)
		}
	})
}
