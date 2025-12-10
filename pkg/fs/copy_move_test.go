package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopy_NewDestination_Succeeds(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "sub", "dst.txt")

	assert.NoError(t, os.WriteFile(src, []byte("hello"), ModeFile))

	err := Copy(src, dst, false)
	assert.NoError(t, err)
	b, _ := os.ReadFile(dst) //nolint:gosec // test helper reads temp file
	assert.Equal(t, "hello", string(b))
}

func TestCopy_ExistingNonEmpty_NoForce_Error(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	assert.NoError(t, os.WriteFile(src, []byte("short"), ModeFile))
	assert.NoError(t, os.WriteFile(dst, []byte("existing"), ModeFile))

	err := Copy(src, dst, false)
	assert.Error(t, err)
	b, _ := os.ReadFile(dst) //nolint:gosec // test helper reads temp file
	assert.Equal(t, "existing", string(b))
}

func TestCopy_ExistingNonEmpty_Force_TruncatesAndOverwrites(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	assert.NoError(t, os.WriteFile(src, []byte("short"), ModeFile))
	// Destination contains longer content which must be truncated when force=true
	assert.NoError(t, os.WriteFile(dst, []byte("existing-long"), ModeFile))

	err := Copy(src, dst, true)
	assert.NoError(t, err)
	b, _ := os.ReadFile(dst) //nolint:gosec // test helper reads temp file
	assert.Equal(t, "short", string(b))
}

func TestCopy_ExistingEmpty_NoForce_AllowsReplace(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	assert.NoError(t, os.WriteFile(src, []byte("data"), ModeFile))
	assert.NoError(t, os.WriteFile(dst, []byte{}, ModeFile))

	err := Copy(src, dst, false)
	assert.NoError(t, err)
	b, _ := os.ReadFile(dst) //nolint:gosec // test helper reads temp file
	assert.Equal(t, "data", string(b))
}

func TestCopy_SamePath_Error(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "file.txt")
	assert.NoError(t, os.WriteFile(src, []byte("x"), ModeFile))
	err := Copy(src, src, true)
	assert.Error(t, err)
}

func TestCopy_InvalidPaths_Error(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "file.txt")
	assert.NoError(t, os.WriteFile(src, []byte("x"), ModeFile))
	assert.Error(t, Copy("", filepath.Join(dir, "a.txt"), false))
	assert.Error(t, Copy(src, "", false))
	assert.Error(t, Copy(src, ".", false))
}

func TestMove_NewDestination_Succeeds(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "sub", "dst.txt")

	assert.NoError(t, os.WriteFile(src, []byte("hello"), ModeFile))

	err := Move(src, dst, false)
	assert.NoError(t, err)
	// Source is removed; dest contains data
	_, serr := os.Stat(src)
	assert.True(t, os.IsNotExist(serr))
	b, _ := os.ReadFile(dst) //nolint:gosec // test helper reads temp file
	assert.Equal(t, "hello", string(b))
}

func TestMove_ExistingNonEmpty_NoForce_Error(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	assert.NoError(t, os.WriteFile(src, []byte("src"), ModeFile))
	assert.NoError(t, os.WriteFile(dst, []byte("dst"), ModeFile))

	err := Move(src, dst, false)
	assert.Error(t, err)
	// Verify both files unchanged
	bsrc, _ := os.ReadFile(src) //nolint:gosec // test helper reads temp file
	bdst, _ := os.ReadFile(dst) //nolint:gosec // test helper reads temp file
	assert.Equal(t, "src", string(bsrc))
	assert.Equal(t, "dst", string(bdst))
}

func TestMove_ExistingEmpty_NoForce_AllowsReplace(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	assert.NoError(t, os.WriteFile(src, []byte("src"), ModeFile))
	assert.NoError(t, os.WriteFile(dst, []byte{}, ModeFile))

	err := Move(src, dst, false)
	assert.NoError(t, err)
	_, serr := os.Stat(src)
	assert.True(t, os.IsNotExist(serr))
	bdst, _ := os.ReadFile(dst) //nolint:gosec // test helper reads temp file
	assert.Equal(t, "src", string(bdst))
}

func TestMove_ExistingNonEmpty_Force_Succeeds(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	assert.NoError(t, os.WriteFile(src, []byte("AAA"), ModeFile))
	assert.NoError(t, os.WriteFile(dst, []byte("BBBBB"), ModeFile))

	err := Move(src, dst, true)
	assert.NoError(t, err)
	_, serr := os.Stat(src)
	assert.True(t, os.IsNotExist(serr))
	bdst, _ := os.ReadFile(dst) //nolint:gosec // test helper reads temp file
	assert.Equal(t, "AAA", string(bdst))
}

func TestMove_SamePath_Error(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "file.txt")
	assert.NoError(t, os.WriteFile(src, []byte("x"), ModeFile))
	err := Move(src, src, true)
	assert.Error(t, err)
}

func TestMove_InvalidPaths_Error(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "file.txt")
	assert.NoError(t, os.WriteFile(src, []byte("x"), ModeFile))
	assert.Error(t, Move("", filepath.Join(dir, "a.txt"), false))
	assert.Error(t, Move(src, "", false))
	assert.Error(t, Move(src, ".", false))
}
