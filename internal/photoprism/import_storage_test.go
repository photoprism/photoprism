package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs/disk"
)

func TestImport_InsufficientStorage(t *testing.T) {
	cfg := config.NewMinimalTestConfig(t.TempDir())
	convert := NewConvert(cfg)
	ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())
	require.NotNil(t, ind)
	imp := NewImport(cfg, ind, convert)
	require.NotNil(t, imp)

	// Seed the disk cache so the verdict is deterministic on any host filesystem.
	t.Run("Healthy", func(t *testing.T) {
		disk.FlushFree()
		t.Cleanup(disk.FlushFree)
		disk.SetFree(cfg.StoragePath(), 999, 1000)
		cfg.Options().FilesQuota = 0

		assert.False(t, imp.insufficientStorage())
	})

	t.Run("StorageLow", func(t *testing.T) {
		disk.FlushFree()
		t.Cleanup(disk.FlushFree)
		disk.SetFree(cfg.StoragePath(), 1, 1000)
		cfg.Options().FilesQuota = 0

		assert.True(t, imp.insufficientStorage())
	})
}
