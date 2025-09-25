package photoprism

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
)

func writeFile(t *testing.T, p string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), fs.ModeDir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, data, fs.ModeFile); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, p string) []byte {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestMediaFile_Copy_Existing_NoForce(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.bin")
	dst := filepath.Join(dir, "dst.bin")

	writeFile(t, src, []byte("ABC"))
	writeFile(t, dst, []byte("LONGER_DEST_CONTENT"))

	m, err := NewMediaFile(src)
	if err != nil {
		t.Fatal(err)
	}

	err = m.Copy(dst, false)
	assert.Error(t, err)
	assert.Equal(t, "LONGER_DEST_CONTENT", string(readFile(t, dst)))
}

func TestMediaFile_Copy_ExistingEmpty_NoForce_AllowsReplace(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.bin")
	dst := filepath.Join(dir, "dst.bin")

	writeFile(t, src, []byte("ABC"))
	// Create an empty destination file.
	writeFile(t, dst, []byte{})

	m, err := NewMediaFile(src)
	if err != nil {
		t.Fatal(err)
	}

	if err = m.Copy(dst, false); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "ABC", string(readFile(t, dst)))
}

func TestMediaFile_Copy_Existing_Force_TruncatesAndOverwrites(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.bin")
	dst := filepath.Join(dir, "dst.bin")

	writeFile(t, src, []byte("ABC"))
	writeFile(t, dst, []byte("LONGER_DEST_CONTENT"))

	m, err := NewMediaFile(src)
	if err != nil {
		t.Fatal(err)
	}

	// Set a known mod time via MediaFile to update cache and file mtime.
	known := time.Date(2020, 5, 4, 3, 2, 1, 0, time.UTC)
	_ = m.SetModTime(known)

	if err = m.Copy(dst, true); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "ABC", string(readFile(t, dst)))
	// Check mtime propagated to destination (second resolution).
	if st, err := os.Stat(dst); err == nil {
		assert.Equal(t, known, st.ModTime().UTC().Truncate(time.Second))
	} else {
		t.Fatal(err)
	}
}

func TestMediaFile_Copy_SamePath_Error(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "file.bin")
	writeFile(t, src, []byte("DATA"))

	m, err := NewMediaFile(src)
	if err != nil {
		t.Fatal(err)
	}
	err = m.Copy(src, true)
	assert.Error(t, err)
}

func TestMediaFile_Copy_InvalidDestPath(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "file.bin")
	writeFile(t, src, []byte("DATA"))

	m, err := NewMediaFile(src)
	if err != nil {
		t.Fatal(err)
	}
	err = m.Copy(".", true)
	assert.Error(t, err)
}

func TestMediaFile_Move_Existing_NoForce(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.bin")
	dst := filepath.Join(dir, "dst.bin")

	writeFile(t, src, []byte("AAA"))
	writeFile(t, dst, []byte("BBB"))

	m, err := NewMediaFile(src)
	if err != nil {
		t.Fatal(err)
	}

	err = m.Move(dst, false)
	assert.Error(t, err)
	// Verify no changes
	assert.FileExists(t, src)
	assert.Equal(t, "BBB", string(readFile(t, dst)))
}

func TestMediaFile_Move_ExistingEmpty_NoForce_AllowsReplace(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.bin")
	dst := filepath.Join(dir, "dst.bin")

	writeFile(t, src, []byte("AAA"))
	// Pre-create empty destination file
	writeFile(t, dst, []byte{})

	m, err := NewMediaFile(src)
	if err != nil {
		t.Fatal(err)
	}

	if err = m.Move(dst, false); err != nil {
		t.Fatal(err)
	}

	// Source removed, destination replaced.
	_, srcErr := os.Stat(src)
	assert.True(t, os.IsNotExist(srcErr))
	assert.Equal(t, "AAA", string(readFile(t, dst)))
	assert.Equal(t, dst, m.FileName())
}

func TestMediaFile_Move_Existing_Force(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.bin")
	dst := filepath.Join(dir, "dst.bin")

	writeFile(t, src, []byte("AAA"))
	writeFile(t, dst, []byte("BBB"))

	m, err := NewMediaFile(src)
	if err != nil {
		t.Fatal(err)
	}

	if err = m.Move(dst, true); err != nil {
		t.Fatal(err)
	}

	// Source removed, destination replaced.
	_, srcErr := os.Stat(src)
	assert.True(t, os.IsNotExist(srcErr))
	assert.Equal(t, "AAA", string(readFile(t, dst)))
	assert.Equal(t, dst, m.FileName())
}

func TestMediaFile_Move_SamePath_Error(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "file.bin")
	writeFile(t, src, []byte("DATA"))

	m, err := NewMediaFile(src)
	if err != nil {
		t.Fatal(err)
	}

	err = m.Move(src, true)
	assert.Error(t, err)
}
