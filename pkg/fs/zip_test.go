package fs

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeZip(t *testing.T, path string, entries map[string][]byte) {
	t.Helper()
	f, err := os.Create(path) //nolint:gosec // test helper creates temp zip file
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		assert.NoError(t, f.Close())
	})

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

func openZipReader(t *testing.T, zipPath string) *zip.ReadCloser {
	t.Helper()

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		assert.NoError(t, r.Close())
	})

	return r
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
		files, skipped, err := Unzip(zipPath, filepath.Join(dir, "a"), 0, -1)
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
	for i := range 5 {
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

// zipTestEntry is a named archive entry for writeOrderedZip; a name ending in "/" is a directory.
type zipTestEntry struct {
	name string
	data []byte
}

// writeOrderedZip writes a zip archive with stored entries in the given order.
func writeOrderedZip(t *testing.T, path string, entries []zipTestEntry) {
	t.Helper()

	var buf bytes.Buffer

	zw := zip.NewWriter(&buf)

	for _, e := range entries {
		w, err := zw.CreateHeader(&zip.FileHeader{Name: e.name, Method: zip.Store})
		require.NoError(t, err)
		_, err = w.Write(e.data)
		require.NoError(t, err)
	}

	require.NoError(t, zw.Close())
	require.NoError(t, os.WriteFile(path, buf.Bytes(), 0o600))
}

// writeZipWithDeclaredSize writes a zip archive with stored entries around one whose header declares the given size.
func writeZipWithDeclaredSize(t *testing.T, path string, entries []zipTestEntry, name string, data []byte, declared uint64, after ...zipTestEntry) {
	t.Helper()
	writeZipWithHeader(t, path, entries, name, data, declared, crc32.ChecksumIEEE(data), after...)
}

// writeZipWithHeader writes a zip archive with stored entries around one whose header declares the given size and CRC.
func writeZipWithHeader(t *testing.T, path string, entries []zipTestEntry, name string, data []byte, declared uint64, crc uint32, after ...zipTestEntry) {
	t.Helper()

	var buf bytes.Buffer

	zw := zip.NewWriter(&buf)

	for _, e := range entries {
		w, err := zw.CreateHeader(&zip.FileHeader{Name: e.name, Method: zip.Store})
		require.NoError(t, err)
		_, err = w.Write(e.data)
		require.NoError(t, err)
	}

	w, err := zw.CreateRaw(&zip.FileHeader{
		Name:               name,
		Method:             zip.Store,
		CRC32:              crc,
		CompressedSize64:   uint64(len(data)),
		UncompressedSize64: declared,
	})
	require.NoError(t, err)
	_, err = w.Write(data)
	require.NoError(t, err)

	for _, e := range after {
		w, err = zw.CreateHeader(&zip.FileHeader{Name: e.name, Method: zip.Store})
		require.NoError(t, err)
		_, err = w.Write(e.data)
		require.NoError(t, err)
	}

	require.NoError(t, zw.Close())
	require.NoError(t, os.WriteFile(path, buf.Bytes(), 0o600))
}

// regularFiles returns the regular files below dir, relative to it and with slash separators.
func regularFiles(t *testing.T, dir string) (result []string) {
	t.Helper()

	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		} else if d.Type().IsRegular() {
			rel, relErr := filepath.Rel(dir, p)
			result = append(result, filepath.ToSlash(rel))
			return relErr
		}

		return nil
	})

	if !os.IsNotExist(err) {
		require.NoError(t, err)
	}

	return result
}

// TestUnzip_TotalSizeLimit checks that the total size budget applies to every file entry of an archive.
func TestUnzip_TotalSizeLimit(t *testing.T) {
	t.Run("OverBudgetEntryStopsExtraction", func(t *testing.T) {
		dir := t.TempDir()
		zipPath := filepath.Join(dir, "budget.zip")
		out := filepath.Join(dir, "out")
		writeOrderedZip(t, zipPath, []zipTestEntry{
			{"a.txt", []byte("aaa")},
			{"big.txt", bytes.Repeat([]byte("b"), 10)},
			{"c.txt", []byte("c")},
			{"d.txt", []byte("d")},
		})

		files, skipped, err := Unzip(zipPath, out, 0, 8)

		require.NoError(t, err)
		assert.Equal(t, []string{filepath.Join(out, "a.txt")}, files)
		assert.Equal(t, []string{"big.txt", "c.txt", "d.txt"}, skipped)
		assert.Equal(t, []string{"a.txt"}, regularFiles(t, out))
	})
	t.Run("ExactFitRejected", func(t *testing.T) {
		dir := t.TempDir()
		zipPath := filepath.Join(dir, "exact.zip")
		out := filepath.Join(dir, "out")
		writeOrderedZip(t, zipPath, []zipTestEntry{
			{"a.txt", []byte("aa")},
			{"b.txt", []byte("bbb")},
			{"c.txt", []byte("c")},
		})

		// b.txt would leave exactly 0 bytes, which the upload size check rejects as well.
		files, skipped, err := Unzip(zipPath, out, 0, 5)

		require.NoError(t, err)
		assert.Equal(t, []string{filepath.Join(out, "a.txt")}, files)
		assert.Equal(t, []string{"b.txt", "c.txt"}, skipped)
		assert.Equal(t, []string{"a.txt"}, regularFiles(t, out))
	})
	t.Run("OneByteLeftFits", func(t *testing.T) {
		dir := t.TempDir()
		zipPath := filepath.Join(dir, "fit.zip")
		out := filepath.Join(dir, "out")
		writeOrderedZip(t, zipPath, []zipTestEntry{
			{"a.txt", []byte("aa")},
			{"b.txt", []byte("bbb")},
		})

		files, skipped, err := Unzip(zipPath, out, 0, 6)

		require.NoError(t, err)
		assert.Equal(t, []string{filepath.Join(out, "a.txt"), filepath.Join(out, "b.txt")}, files)
		assert.Empty(t, skipped)
	})
	t.Run("ZeroBudget", func(t *testing.T) {
		dir := t.TempDir()
		zipPath := filepath.Join(dir, "zero.zip")
		out := filepath.Join(dir, "out")
		writeOrderedZip(t, zipPath, []zipTestEntry{
			{"sub/", nil},
			{"sub/a.txt", []byte("a")},
			{"empty.txt", nil},
		})

		// Directory entries cost no bytes and are still created; every file entry is skipped.
		files, skipped, err := Unzip(zipPath, out, 0, 0)

		require.NoError(t, err)
		assert.Equal(t, []string{filepath.Join(out, "sub")}, files)
		assert.Equal(t, []string{"sub/a.txt", "empty.txt"}, skipped)
		assert.Empty(t, regularFiles(t, out))
	})
	t.Run("Unlimited", func(t *testing.T) {
		dir := t.TempDir()
		zipPath := filepath.Join(dir, "unlimited.zip")
		out := filepath.Join(dir, "out")
		writeOrderedZip(t, zipPath, []zipTestEntry{
			{"a.txt", bytes.Repeat([]byte("a"), 1000)},
			{"b.txt", []byte("b")},
		})

		files, skipped, err := Unzip(zipPath, out, 0, -1)

		require.NoError(t, err)
		assert.Equal(t, []string{filepath.Join(out, "a.txt"), filepath.Join(out, "b.txt")}, files)
		assert.Empty(t, skipped)
	})
	t.Run("AllEntriesFit", func(t *testing.T) {
		dir := t.TempDir()
		zipPath := filepath.Join(dir, "fit.zip")
		out := filepath.Join(dir, "out")
		writeOrderedZip(t, zipPath, []zipTestEntry{
			{"a.txt", []byte("aaa")},
			{"sub/b.txt", []byte("bb")},
			{"c.txt", []byte("c")},
		})

		files, skipped, err := Unzip(zipPath, out, 10, 100)

		require.NoError(t, err)
		assert.Equal(t, []string{filepath.Join(out, "a.txt"), filepath.Join(out, "sub/b.txt"), filepath.Join(out, "c.txt")}, files)
		assert.Empty(t, skipped)
		assert.Equal(t, []string{"a.txt", "c.txt", "sub/b.txt"}, regularFiles(t, out))
	})
	t.Run("EntryShorterThanDeclared", func(t *testing.T) {
		dir := t.TempDir()
		zipPath := filepath.Join(dir, "short.zip")
		out := filepath.Join(dir, "out")
		writeZipWithDeclaredSize(t, zipPath, []zipTestEntry{{"a.txt", []byte("aaa")}}, "short.txt", []byte("ss"), 4, zipTestEntry{"b.txt", []byte("b")})

		files, _, err := Unzip(zipPath, out, 0, 8)

		require.ErrorIs(t, err, io.ErrUnexpectedEOF)
		assert.Equal(t, []string{filepath.Join(out, "a.txt")}, files)
		assert.Equal(t, []string{"a.txt"}, regularFiles(t, out))
	})
	t.Run("ChecksumMismatch", func(t *testing.T) {
		data := []byte("crc-control")

		// Covers a cap above the declared size, a cap equal to it, and no limits at all.
		for name, limits := range map[string][2]int64{
			"CapAboveSize": {1024, 1024},
			"CapEqualSize": {0, int64(len(data)) + 1},
			"Unlimited":    {0, -1},
		} {
			t.Run(name, func(t *testing.T) {
				dir := t.TempDir()
				zipPath := filepath.Join(dir, "crc.zip")
				out := filepath.Join(dir, "out")
				writeZipWithHeader(t, zipPath, nil, "bad.txt", data, uint64(len(data)), crc32.ChecksumIEEE(data)+1)

				files, _, err := Unzip(zipPath, out, limits[0], limits[1])

				require.ErrorIs(t, err, zip.ErrChecksum)
				assert.Empty(t, files)
				assert.Empty(t, regularFiles(t, out))
			})
		}
	})
	t.Run("DataExceedsDeclaredSizeAtCap", func(t *testing.T) {
		// The cap equals the declared size, without limits and with one byte of budget to spare.
		for name, total := range map[string]int64{"Unlimited": -1, "Budget": 32} {
			t.Run(name, func(t *testing.T) {
				dir := t.TempDir()
				zipPath := filepath.Join(dir, "declared.zip")
				out := filepath.Join(dir, "out")
				writeZipWithDeclaredSize(t, zipPath, nil, "b.txt", bytes.Repeat([]byte("b"), 64), 31)

				files, _, err := Unzip(zipPath, out, 0, total)

				require.ErrorIs(t, err, zip.ErrFormat)
				assert.Empty(t, files)
				assert.Empty(t, regularFiles(t, out))
			})
		}
	})
	t.Run("DataExceedsDeclaredSize", func(t *testing.T) {
		dir := t.TempDir()
		zipPath := filepath.Join(dir, "declared.zip")
		out := filepath.Join(dir, "out")
		writeZipWithDeclaredSize(t, zipPath, []zipTestEntry{{"a.txt", []byte("aaa")}}, "b.txt", bytes.Repeat([]byte("b"), 64), 4)

		files, _, err := Unzip(zipPath, out, 0, 32)

		require.Error(t, err)
		assert.NotContains(t, files, filepath.Join(out, "b.txt"))
		assert.NoFileExists(t, filepath.Join(out, "b.txt"))
		assert.Equal(t, []string{"a.txt"}, regularFiles(t, out))
	})
}

// TestUnzipEntry checks that an entry is written up to its cap and removed when its data exceeds it.
func TestUnzipEntry(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "entry.zip")
	writeOrderedZip(t, zipPath, []zipTestEntry{{"sub/", nil}, {"a.txt", []byte("abcdef")}})
	r := openZipReader(t, zipPath)

	t.Run("Success", func(t *testing.T) {
		out := t.TempDir()
		fileName, written, err := unzipEntry(r.File[1], out, 6)

		require.NoError(t, err)
		assert.Equal(t, filepath.Join(out, "a.txt"), fileName)
		assert.Equal(t, int64(6), written)
		assert.FileExists(t, fileName)
	})
	t.Run("DeclaredSizeWithoutCap", func(t *testing.T) {
		out := t.TempDir()
		_, written, err := unzipEntry(r.File[1], out, 0)

		require.NoError(t, err)
		assert.Equal(t, int64(6), written)
	})
	t.Run("Directory", func(t *testing.T) {
		out := t.TempDir()
		fileName, written, err := unzipEntry(r.File[0], out, 1)

		require.NoError(t, err)
		assert.Equal(t, int64(0), written)
		assert.DirExists(t, fileName)
	})
	t.Run("ShorterThanDeclared", func(t *testing.T) {
		shortZip := filepath.Join(dir, "short.zip")
		writeZipWithDeclaredSize(t, shortZip, nil, "short.txt", []byte("ss"), 4)
		short := openZipReader(t, shortZip)
		out := t.TempDir()

		fileName, _, err := unzipEntry(short.File[0], out, 0)

		require.ErrorIs(t, err, io.ErrUnexpectedEOF)
		require.Equal(t, filepath.Join(out, "short.txt"), fileName)
		assert.NoFileExists(t, fileName)
	})
	t.Run("ChecksumMismatchAtDeclaredSize", func(t *testing.T) {
		crcZip := filepath.Join(dir, "crc.zip")
		writeZipWithHeader(t, crcZip, nil, "bad.txt", []byte("abc"), 3, crc32.ChecksumIEEE([]byte("abc"))+1)
		bad := openZipReader(t, crcZip)
		out := t.TempDir()

		fileName, _, err := unzipEntry(bad.File[0], out, 0)

		require.ErrorIs(t, err, zip.ErrChecksum)
		require.Equal(t, filepath.Join(out, "bad.txt"), fileName)
		assert.NoFileExists(t, fileName)
	})
	t.Run("ExceedsCap", func(t *testing.T) {
		out := t.TempDir()
		fileName, written, err := unzipEntry(r.File[1], out, 4)

		require.Error(t, err)
		assert.LessOrEqual(t, written, int64(4))
		assert.NoFileExists(t, fileName)
		assert.Empty(t, regularFiles(t, out))
	})
	t.Run("CreationMode", func(t *testing.T) {
		modeZip := filepath.Join(dir, "mode.zip")
		zf, err := os.Create(modeZip) //nolint:gosec // test path in a temporary directory
		require.NoError(t, err)
		zw := zip.NewWriter(zf)
		modes := []os.FileMode{os.ModeSetuid | os.ModeSetgid | os.ModeSticky | 0o777, 0o400}
		for i, mode := range modes {
			header := &zip.FileHeader{Name: fmt.Sprintf("mode%d.txt", i), Method: zip.Store}
			header.SetMode(mode)
			w, createErr := zw.CreateHeader(header)
			require.NoError(t, createErr)
			_, err = w.Write([]byte("mode"))
			require.NoError(t, err)
		}
		require.NoError(t, zw.Close())
		require.NoError(t, zf.Close())

		entries := openZipReader(t, modeZip).File
		require.NotZero(t, entries[0].Mode()&os.ModeSetuid)
		out := t.TempDir()

		for _, entry := range entries {
			fileName, _, err := unzipEntry(entry, out, 0)
			require.NoError(t, err)
			info, err := os.Stat(fileName)
			require.NoError(t, err)

			assert.Zero(t, info.Mode()&(os.ModeSetuid|os.ModeSetgid|os.ModeSticky), entry.Name)
			assert.Zero(t, info.Mode().Perm()&0o111, entry.Name)
			assert.Equal(t, os.FileMode(0o600), info.Mode().Perm()&0o600, entry.Name)
		}
	})
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

	r := openZipReader(t, zipPath)

	if len(r.File) != 1 {
		t.Fatalf("expected one file, got %d", len(r.File))
	}

	_, err := unzipFileWithLimit(r.File[0], dir, 1) // limit below actual size
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

// TestUnzip_Filters checks file selection before extraction without changing unfiltered callers.
func TestUnzip_Filters(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "batch.zip")
	writeZip(t, archive, map[string][]byte{"keep.jpg": []byte("image-control"), "nested/omit.yml": []byte("metadata-control")})
	// images selects a file type without reading entry contents.
	images := func(name string, isDir bool) bool { return isDir || FileType(name) == ImageJpeg }
	t.Run("SelectedBeforeWrite", func(t *testing.T) {
		dest := t.TempDir()
		existing := filepath.Join(dest, "nested/omit.yml")

		assert.NoError(t, MkdirAll(filepath.Dir(existing)))
		assert.NoError(t, os.WriteFile(existing, []byte("existing-control"), ModeFile))

		files, skipped, err := Unzip(archive, dest, 1024, 1024, images)
		assert.NoError(t, err)
		assert.Equal(t, []string{filepath.Join(dest, "keep.jpg")}, files)
		assert.Equal(t, []string{"nested/omit.yml"}, skipped)

		data, err := os.ReadFile(existing) //nolint:gosec // Test reads its temporary control file.

		assert.NoError(t, err)
		assert.Equal(t, "existing-control", string(data))
	})
	t.Run("DirectoryFlag", func(t *testing.T) {
		zipPath := filepath.Join(t.TempDir(), "directories.zip")
		archive, err := os.Create(zipPath) //nolint:gosec // Test creates its temporary archive.

		assert.NoError(t, err)

		writer := zip.NewWriter(archive)

		for _, name := range []string{"keep", "omit"} {
			hdr := &zip.FileHeader{Name: name}
			hdr.SetMode(os.ModeDir | ModeDir)
			_, err = writer.CreateHeader(hdr)
			assert.NoError(t, err)
		}

		assert.NoError(t, writer.Close())
		assert.NoError(t, archive.Close())

		dest := t.TempDir()

		files, skipped, err := Unzip(zipPath, dest, 1024, 1024, func(name string, isDir bool) bool {
			assert.True(t, isDir)
			return name == "keep"
		})

		assert.NoError(t, err)
		assert.Equal(t, []string{filepath.Join(dest, "keep")}, files)
		assert.Equal(t, []string{"omit"}, skipped)
		assert.DirExists(t, filepath.Join(dest, "keep"))
		assert.NoDirExists(t, filepath.Join(dest, "omit"))
	})
	t.Run("NilFilter", func(t *testing.T) {
		files, skipped, err := Unzip(archive, t.TempDir(), 1024, 1024, nil)
		assert.NoError(t, err)
		assert.Len(t, files, 2)
		assert.Empty(t, skipped)
	})
	t.Run("AllFiltersRequired", func(t *testing.T) {
		files, skipped, err := Unzip(archive, t.TempDir(), 1024, 1024, images, func(string, bool) bool { return false })
		assert.NoError(t, err)
		assert.Empty(t, files)
		assert.Len(t, skipped, 2)
	})
}

// TestUnzipReservedNames checks the baseline path policy without caller-supplied filters.
func TestUnzipReservedNames(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "reserved.zip")
	blocked := []string{".ppignore", ".dockerignore", ".python_history", ".bash_history-04218.tmp", ".my.cnf", ".mylogin.cnf", ".rsyncignore", ".rsync-filter", ".gitignore", "_netrc", PPStorageFilename, SigningKeyFile, JoinTokenFile, ClientSecretFile, ".git/config", ".svn/photo.jpg", ".hg/photo.jpg", ".ssh/key", ".gnupg/key", ".env", "nested/.env.production", ".config/photo.jpg", ".photoprism/photo.jpg", "nested/.GiT/photo.jpg", "photo.rclonelink", "nested/LINK.RcloneLink/photo.jpg", ".config/"}

	for _, name := range ReservedPathNames() {
		blocked = append(blocked, "nested/"+name+"/photo.jpg")
	}

	entries := map[string][]byte{"allowed.json": []byte("metadata-control"), ".hidden/photo.txt": []byte("ordinary-control"), ".github-backup/photo.txt": []byte("ordinary-control")}

	for _, name := range blocked {
		entries[name] = []byte("excluded-control")
		if name == ".config/" {
			entries[name] = nil
		}
	}

	writeZip(t, archive, entries)

	dest := filepath.Join(dir, "output")

	require.NoError(t, MkdirAll(dest))
	preserved := []string{".my.cnf", ".mylogin.cnf", ".rsyncignore", ".rsync-filter", ".bash_history-04218.tmp", ".gitignore", PPStorageFilename, SigningKeyFile, JoinTokenFile, ClientSecretFile}
	for _, name := range preserved {
		require.NoError(t, os.WriteFile(filepath.Join(dest, name), []byte("existing-control"), ModeFile))
	}

	files, skipped, err := Unzip(archive, dest, 1024, 100000)

	assert.NoError(t, err)
	assert.Len(t, files, 3)
	assert.ElementsMatch(t, blocked, skipped)

	for _, name := range preserved {
		data, readErr := os.ReadFile(filepath.Join(dest, name)) //nolint:gosec // Test reads its temporary filesystem control.

		require.NoError(t, readErr)
		assert.Equal(t, "existing-control", string(data), name)
	}

	assert.FileExists(t, filepath.Join(dest, "allowed.json"))
	assert.NoFileExists(t, filepath.Join(dest, ".ppignore"))
	assert.FileExists(t, filepath.Join(dest, ".hidden/photo.txt"))
	assert.NoDirExists(t, filepath.Join(dest, ".git"))
	assert.NoDirExists(t, filepath.Join(dest, ".config"))

	reader := openZipReader(t, archive)

	for _, entry := range reader.File {
		if entry.Name == ".git/config" {
			_, err = UnzipFile(entry, filepath.Join(dir, "single"))
			assert.ErrorIs(t, err, ErrReservedPath)
		}
	}

	assert.NoDirExists(t, filepath.Join(dir, "single"))
}

// TestUnzipOperatorLinks preserves operator-managed directory mappings.
func TestUnzipOperatorLinks(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out")
	require.NoError(t, MkdirAll(filepath.Join(dest, ".config")))
	normal := filepath.Join(dir, "normal")
	require.NoError(t, MkdirAll(normal))
	require.NoError(t, os.Symlink(filepath.Join(dest, ".config"), filepath.Join(dest, "reserved-alias")))
	require.NoError(t, os.Symlink(normal, filepath.Join(dest, "ordinary-alias")))
	archive := filepath.Join(dir, "alias.zip")
	writeZip(t, archive, map[string][]byte{"reserved-alias/new.txt": []byte("excluded"), "ordinary-alias/new.txt": []byte("allowed")})
	files, skipped, err := Unzip(archive, dest, 1024, 10000)
	require.NoError(t, err)
	assert.Len(t, files, 2)
	assert.Empty(t, skipped)
	assert.FileExists(t, filepath.Join(dest, ".config/new.txt"))
	data, err := os.ReadFile(filepath.Join(normal, "new.txt")) //nolint:gosec // Test reads its temporary transfer output.
	require.NoError(t, err)
	assert.Equal(t, "allowed", string(data))
}

// TestUnzipSymlinkEntries checks omission of archive links before any destination write.
func TestUnzipSymlinkEntries(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "links.zip")
	f, err := os.Create(archive) //nolint:gosec // Test creates an archive in its temporary directory.
	require.NoError(t, err)
	w := zip.NewWriter(f)

	for _, name := range []string{"link.jpg", "folder-link"} {
		header := &zip.FileHeader{Name: name}
		header.SetMode(os.ModeSymlink | ModeFile)
		entry, createErr := w.CreateHeader(header)
		require.NoError(t, createErr)
		_, err = entry.Write([]byte("../outside"))
		require.NoError(t, err)
	}

	entry, err := w.Create("folder-link/ordinary.txt")
	require.NoError(t, err)
	_, err = entry.Write([]byte("ordinary-control"))
	require.NoError(t, err)
	require.NoError(t, w.Close())
	require.NoError(t, f.Close())

	dest := filepath.Join(dir, "out")
	require.NoError(t, MkdirAll(dest))
	require.NoError(t, os.WriteFile(filepath.Join(dest, "link.jpg"), []byte("existing-control"), ModeFile))
	files, skipped, err := Unzip(archive, dest, 1024, 10000)
	require.NoError(t, err)
	assert.Equal(t, []string{filepath.Join(dest, "folder-link/ordinary.txt")}, files)
	assert.ElementsMatch(t, []string{"link.jpg", "folder-link"}, skipped)
	assert.DirExists(t, filepath.Join(dest, "folder-link"))
	assert.NoFileExists(t, filepath.Join(dir, "outside"))

	info, err := os.Lstat(filepath.Join(dest, "folder-link"))
	require.NoError(t, err)
	assert.Zero(t, info.Mode()&os.ModeSymlink)
	data, err := os.ReadFile(filepath.Join(dest, "link.jpg")) //nolint:gosec // Test reads its temporary filesystem control.
	require.NoError(t, err)
	assert.Equal(t, "existing-control", string(data))

	reader := openZipReader(t, archive)
	for _, file := range reader.File {
		if file.Name == "link.jpg" {
			require.NotZero(t, file.Mode()&os.ModeSymlink)
			_, err = UnzipFile(file, filepath.Join(dir, "single"))
			assert.ErrorIs(t, err, ErrArchiveSymlink)
		}
	}
	assert.NoDirExists(t, filepath.Join(dir, "single"))
}
