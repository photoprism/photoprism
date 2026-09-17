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
	blocked := []string{".ppignore", ".dockerignore", ".python_history", ".bash_history-04218.tmp", ".my.cnf", ".mylogin.cnf", ".rsyncignore", ".rsync-filter", ".gitignore", "_netrc", PPStorageFilename, SigningKeyFile, JoinTokenFile, ClientSecretFile, ".git/config", ".svn/photo.jpg", ".hg/photo.jpg", ".ssh/key", ".gnupg/key", ".env", "nested/.env.production", ".config/photo.jpg", ".photoprism/photo.jpg", "nested/.GiT/photo.jpg", ".config/"}

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
