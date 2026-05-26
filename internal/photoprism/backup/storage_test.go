package backup

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/fs/disk"
	"github.com/photoprism/photoprism/pkg/log/status"
)

func TestAlbums_InsufficientStorage(t *testing.T) {
	backupPath, err := filepath.Abs("./testdata/albums-insufficient")
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(backupPath, fs.ModeDir))
	t.Cleanup(func() { _ = os.RemoveAll(backupPath) })

	conf := get.Config()
	disk.FlushFree()
	t.Cleanup(disk.FlushFree)
	disk.SetFree(conf.StoragePath(), 1, 1000)

	count, err := Albums(backupPath, true)

	assert.True(t, errors.Is(err, status.ErrInsufficientStorage), "expected status.ErrInsufficientStorage, got %v", err)
	// The loop aborts before writing the first YAML.
	assert.Equal(t, 0, count)
}

func TestDatabase_InsufficientStorage(t *testing.T) {
	backupPath, err := filepath.Abs("./testdata/sqlite-insufficient")
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(backupPath, fs.ModeDir))
	t.Cleanup(func() { _ = os.RemoveAll(backupPath) })

	conf := get.Config()
	disk.FlushFree()
	t.Cleanup(disk.FlushFree)
	disk.SetFree(conf.StoragePath(), 1, 1000)

	err = Database(backupPath, "", false, true, 2)

	assert.True(t, errors.Is(err, status.ErrInsufficientStorage), "expected status.ErrInsufficientStorage, got %v", err)
}
