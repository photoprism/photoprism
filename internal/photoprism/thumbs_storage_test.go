package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs/disk"
)

func TestThumbs_StorageLow(t *testing.T) {
	cfg := config.NewMinimalTestConfig(t.TempDir())
	w := NewThumbs(cfg)
	require.NotNil(t, w)

	// Reset the storage-check latch, which an earlier test may have tripped by probing a
	// temp path that duf cannot resolve to a mount point.
	config.DisableStorageCheck.Store(false)

	// Seeding the cache keeps the verdict independent of the host filesystem and of
	// duf's ability to resolve the temp directory to a mount point.
	t.Run("Healthy", func(t *testing.T) {
		disk.FlushFree()
		t.Cleanup(disk.FlushFree)
		disk.SetFree(cfg.StoragePath(), 999*disk.MB, 1000*disk.MB)

		assert.False(t, w.storageLow())
	})
	t.Run("Low", func(t *testing.T) {
		disk.FlushFree()
		t.Cleanup(disk.FlushFree)
		disk.SetFree(cfg.StoragePath(), 1*disk.MB, 1000*disk.MB)

		assert.True(t, w.storageLow())
	})
}
