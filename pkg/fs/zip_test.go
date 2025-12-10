package fs

import (
	"archive/zip"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func writeZip(t *testing.T, path string, entries map[string][]byte) {
	t.Helper()
	f, err := os.Create(path) //nolint:gosec // test helper creates temp zip file
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

	t.Run("UnlimitedTotalSize", func(t *testing.T) {
		files, skipped, err := Unzip(zipPath, filepath.Join(dir, "a"), 0, 0)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{
			filepath.Join(dir, "a", "ok1.txt"),
			filepath.Join(dir, "a", "ok2.txt"),
		}, files)
		assert.GreaterOrEqual(t, len(skipped), 2) // __MACOSX and evil path skipped
	})
	t.Run("WithEntryAndTotalLimits", func(t *testing.T) {
		outDir := filepath.Join(dir, "b")
		files, skipped, err := Unzip(zipPath, outDir, 2, 3) // file limit=2 bytes; total limit=3 bytes
		assert.NoError(t, err)

		// ok1 (3 bytes) skipped by file limit; evil skipped by '..'; __MACOSX skipped by prefix
		// ok2 (1 byte) allowed; total limit reduces to 2; nothing else left that fits
		assert.ElementsMatch(t, []string{filepath.Join(outDir, "ok2.txt")}, files)
		// Ensure file written
		b, rerr := os.ReadFile(filepath.Join(outDir, "ok2.txt")) //nolint:gosec // test helper reads temp file
		assert.NoError(t, rerr)
		assert.Equal(t, []byte("x"), b)
		// Skipped contains at least the three excluded entries
		assert.GreaterOrEqual(t, len(skipped), 3)
	})
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

func TestUnzip_SkipsVeryLargeEntry(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "huge.zip")

	writeZip64Stub(t, zipPath, "huge.bin", math.MaxUint64)

	files, skipped, err := Unzip(zipPath, filepath.Join(dir, "out"), 0, -1)
	assert.NoError(t, err)
	assert.Empty(t, files)
	assert.Contains(t, skipped, "huge.bin")
}

func TestUnzip_EntryLimit(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "limit.zip")

	entries := map[string][]byte{}
	for i := 0; i < 5; i++ {
		entries[fmt.Sprintf("f%d.txt", i)] = []byte("x")
	}
	writeZip(t, zipPath, entries)

	orig := MaxUnzipEntries
	MaxUnzipEntries = 3
	defer func() { MaxUnzipEntries = orig }()

	_, _, err := Unzip(zipPath, filepath.Join(dir, "out"), 0, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "entry limit")
}

// writeZip64Stub writes a minimal ZIP64 archive with one stored entry and custom size values.
func writeZip64Stub(t *testing.T, path, name string, size uint64) {
	t.Helper()

	var buf []byte

	bw := func(data []byte) {
		buf = append(buf, data...)
	}

	writeLE := func(v any) {
		var b [8]byte
		switch x := v.(type) {
		case uint16:
			binary.LittleEndian.PutUint16(b[:2], x)
			bw(b[:2])
		case uint32:
			binary.LittleEndian.PutUint32(b[:4], x)
			bw(b[:4])
		case uint64:
			binary.LittleEndian.PutUint64(b[:8], x)
			bw(b[:8])
		default:
			t.Fatalf("unsupported type %T", v)
		}
	}

	filename := []byte(name)
	const (
		sigLocal   = 0x04034b50
		sigCentral = 0x02014b50
		sigEnd     = 0x06054b50
	)

	zip64ExtraLen := uint16(4 + 16) // header id + size + two uint64 values
	localExtraLen := zip64ExtraLen
	centralExtraLen := zip64ExtraLen

	// Local file header
	writeLE(uint32(sigLocal))
	writeLE(uint16(45)) // version needed (zip64)
	writeLE(uint16(0))  // flags
	writeLE(uint16(0))  // method store
	writeLE(uint16(0))  // mod time
	writeLE(uint16(0))  // mod date
	writeLE(uint32(0))  // crc
	writeLE(uint32(0xFFFFFFFF))
	writeLE(uint32(0xFFFFFFFF))
	if len(filename) > math.MaxUint16 {
		t.Fatalf("filename too long")
	}
	writeLE(uint16(len(filename))) //nolint:gosec // filename length checked above
	writeLE(localExtraLen)
	bw(filename)
	// zip64 extra
	writeLE(uint16(0x0001)) // header id
	writeLE(uint16(16))     // data size
	writeLE(size)           // uncompressed size
	writeLE(size)           // compressed size
	// no file data (size 0) to keep archive tiny

	localLen := len(buf)

	// Central directory header
	writeLE(uint32(sigCentral))
	writeLE(uint16(45)) // version made by
	writeLE(uint16(45)) // version needed
	writeLE(uint16(0))  // flags
	writeLE(uint16(0))  // method
	writeLE(uint16(0))  // time
	writeLE(uint16(0))  // date
	writeLE(uint32(0))  // crc
	writeLE(uint32(0xFFFFFFFF))
	writeLE(uint32(0xFFFFFFFF))
	if len(filename) > math.MaxUint16 {
		t.Fatalf("filename too long")
	}
	writeLE(uint16(len(filename))) //nolint:gosec // filename length checked above
	writeLE(centralExtraLen)
	writeLE(uint16(0)) // comment len
	writeLE(uint16(0)) // disk start
	writeLE(uint16(0)) // int attrs
	writeLE(uint32(0)) // ext attrs
	writeLE(uint32(0)) // rel offset (zip64 overrides)
	bw(filename)
	// zip64 extra
	writeLE(uint16(0x0001))
	writeLE(uint16(16))
	writeLE(size) // uncompressed
	writeLE(size) // compressed

	centralLen := len(buf) - localLen

	// End of central directory (not zip64 EOCD; minimal to satisfy reader)
	writeLE(uint32(sigEnd))
	writeLE(uint16(0)) // disk
	writeLE(uint16(0)) // start disk
	writeLE(uint16(1)) // entries this disk
	writeLE(uint16(1)) // total entries
	if centralLen > math.MaxUint32 || localLen > math.MaxUint32 {
		t.Fatalf("central or local length exceeds uint32")
	}
	writeLE(uint32(centralLen)) //nolint:gosec // lengths checked above
	writeLE(uint32(localLen))   //nolint:gosec
	writeLE(uint16(0))          // comment length

	if err := os.WriteFile(path, buf, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestUnzipFileWithLimit_DetectsOverrun(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "small.zip")
	writeZip(t, zipPath, map[string][]byte{"a.txt": []byte("abc")}) // 3 bytes

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	if len(r.File) != 1 {
		t.Fatalf("expected one file, got %d", len(r.File))
	}

	_, err = unzipFileWithLimit(r.File[0], dir, 1) // limit below actual size
	if err == nil {
		t.Fatalf("expected limit overrun error")
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

func TestSafeJoin(t *testing.T) {
	base := filepath.Clean("/tmp/base")

	if runtime.GOOS == "windows" {
		base = filepath.Clean(`C:\tmp\base`)
	}

	type testCase struct {
		name     string
		baseDir  string
		input    string
		wantErr  bool
		wantPath string
	}

	tests := []testCase{
		{
			name:     "SimpleFile",
			baseDir:  base,
			input:    "a.txt",
			wantPath: filepath.Join(base, "a.txt"),
		},
		{
			name:     "NestedRelative",
			baseDir:  base,
			input:    "nested/dir/file.jpg",
			wantPath: filepath.Join(base, "nested", "dir", "file.jpg"),
		},
		{
			name:    "EmptyName",
			baseDir: base,
			input:   "",
			wantErr: true,
		},
		{
			name:    "DotDotTraversal",
			baseDir: base,
			input:   "../secret.txt",
			wantErr: true,
		},
		{
			name:     "MixedSeparatorsTraversal",
			baseDir:  base,
			input:    `dir\..\evil.txt`,
			wantPath: filepath.Join(base, "evil.txt"),
		},
		{
			name:     "ContainsParentInMiddle",
			baseDir:  base,
			input:    "dir/../evil.txt",
			wantPath: filepath.Join(base, "evil.txt"),
		},
		{
			name:    "AbsoluteUnix",
			baseDir: base,
			input:   "/etc/passwd",
			wantErr: true,
		},
		{
			name:    "WindowsVolumeForwardSlash",
			baseDir: base,
			input:   "C:/Windows/System32/evil.txt",
			wantErr: true,
		},
		{
			name:    "WindowsVolumeBackslash",
			baseDir: base,
			input:   `D:\Data\evil.txt`,
			wantErr: true,
		},
		{
			name:     "CleansInsideBase",
			baseDir:  base,
			input:    "sub/../ok.txt",
			wantPath: filepath.Join(base, "ok.txt"),
		},
		{
			name:     "RepeatedSeparators",
			baseDir:  base,
			input:    "dir//file.txt",
			wantPath: filepath.Join(base, "dir", "file.txt"),
		},
		{
			name:    "VolumeNameOnly",
			baseDir: base,
			input:   "C:",
			wantErr: true,
		},
		{
			name:    "RootedBackslashUnix",
			baseDir: base,
			input:   `\evil.txt`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := safeJoin(tc.baseDir, tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got none (path=%q)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantPath {
				t.Fatalf("unexpected path: got %q want %q", got, tc.wantPath)
			}
		})
	}
}
