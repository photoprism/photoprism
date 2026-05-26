package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs/disk"
)

func TestIndex_StorageLow(t *testing.T) {
	cfg := config.NewMinimalTestConfig(t.TempDir())
	ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())
	require.NotNil(t, ind)

	// Seeding the cache keeps the verdict independent of the host filesystem and of
	// duf's ability to resolve the temp directory to a mount point.
	t.Run("Healthy", func(t *testing.T) {
		disk.FlushFree()
		t.Cleanup(disk.FlushFree)
		disk.SetFree(cfg.StoragePath(), 999, 1000)

		assert.False(t, ind.storageLow())
	})

	t.Run("Low", func(t *testing.T) {
		disk.FlushFree()
		t.Cleanup(disk.FlushFree)
		disk.SetFree(cfg.StoragePath(), 1, 1000)

		assert.True(t, ind.storageLow())
	})
}
