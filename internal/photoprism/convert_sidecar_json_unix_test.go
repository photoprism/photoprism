//go:build unix

package photoprism

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExifToolCacheValid_Pipe(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "cache.json")
	require.NoError(t, syscall.Mkfifo(fileName, 0o600))
	assert.False(t, exifToolCacheValid(fileName))
}

func TestWriteExifToolCache_Unix(t *testing.T) {
	t.Run("ReplacesLink", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "target.json")
		fileName := filepath.Join(dir, "cache.json")
		require.NoError(t, os.WriteFile(target, []byte("[]"), 0o600))
		require.NoError(t, os.Symlink(target, fileName))
		require.NoError(t, writeExifToolCache(fileName, []byte("[{}]\n")))
		info, err := os.Lstat(fileName)
		require.NoError(t, err)
		assert.True(t, info.Mode().IsRegular(), "the link is replaced by a regular file")
		data, err := os.ReadFile(target) //nolint:gosec // test-owned file
		require.NoError(t, err)
		assert.Equal(t, "[]", string(data), "the link target is not written")
	})
	t.Run("ReadOnlyFile", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("root can write read-only files")
		}

		fileName := filepath.Join(t.TempDir(), "cache.json")
		require.NoError(t, os.WriteFile(fileName, []byte("[{"), 0o400))
		require.NoError(t, writeExifToolCache(fileName, []byte("[{}]\n")))
		data, err := os.ReadFile(fileName) //nolint:gosec // test-owned file
		require.NoError(t, err)
		assert.Equal(t, "[{}]\n", string(data))
	})
}
