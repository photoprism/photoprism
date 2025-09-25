package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolve_FileAndSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "file.txt")
	link := filepath.Join(dir, "link.txt")

	assert.NoError(t, os.WriteFile(target, []byte("x"), ModeFile))
	// Create symlink if supported on this platform
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks not supported: %v", err)
		return
	}

	// Resolving the file returns its absolute path
	absFile, err := Resolve(target)
	assert.NoError(t, err)
	assert.True(t, filepath.IsAbs(absFile))

	// Resolving the link returns the target absolute path
	absLink, err := Resolve(link)
	assert.NoError(t, err)
	assert.Equal(t, absFile, absLink)
}

func TestResolve_BrokenSymlink_Error(t *testing.T) {
	dir := t.TempDir()
	broken := filepath.Join(dir, "broken.txt")
	// Symlink to missing target
	if err := os.Symlink(filepath.Join(dir, "missing.txt"), broken); err != nil {
		t.Skipf("symlinks not supported: %v", err)
		return
	}
	_, err := Resolve(broken)
	assert.Error(t, err)
}
