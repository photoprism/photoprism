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
	disk.SetFree(conf.StoragePath(), 1*disk.MB, 1000*disk.MB)

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
	disk.SetFree(conf.StoragePath(), 1*disk.MB, 1000*disk.MB)

	err = Database(backupPath, "", false, true, 2)

	assert.True(t, errors.Is(err, status.ErrInsufficientStorage), "expected status.ErrInsufficientStorage, got %v", err)
}

func TestDatabase_StdoutBypassesStorageGate(t *testing.T) {
	conf := get.Config()
	disk.FlushFree()
	t.Cleanup(disk.FlushFree)
	disk.SetFree(conf.StoragePath(), 1*disk.MB, 1000*disk.MB)

	// Stdout dumps must not be blocked by the storage gate so an operator can
	// still offload a backup (for example over ssh) when the local volume is
	// full. The actual mariadb-dump invocation will fail in this minimal
	// fixture, but the error must not be status.ErrInsufficientStorage.
	err := Database("", "-", true, true, 0)

	assert.False(t, errors.Is(err, status.ErrInsufficientStorage),
		"stdout backups must bypass the storage gate, got %v", err)
}
