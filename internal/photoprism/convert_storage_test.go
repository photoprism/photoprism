package photoprism

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs/disk"
	"github.com/photoprism/photoprism/pkg/log/status"
)

func TestConvert_InsufficientStorage(t *testing.T) {
	cfg := config.NewMinimalTestConfig(t.TempDir())
	convert := NewConvert(cfg)
	require.NotNil(t, convert)

	// Reset the storage-check latch, which an earlier test may have tripped by probing a
	// temp path that duf cannot resolve to a mount point.
	config.DisableStorageCheck.Store(false)

	t.Run("Healthy", func(t *testing.T) {
		disk.FlushFree()
		t.Cleanup(disk.FlushFree)
		disk.SetFree(cfg.StoragePath(), 999*disk.MB, 1000*disk.MB)
		cfg.Options().FilesQuota = 0

		assert.False(t, convert.insufficientStorage())
	})
	t.Run("StorageLow", func(t *testing.T) {
		disk.FlushFree()
		t.Cleanup(disk.FlushFree)
		disk.SetFree(cfg.StoragePath(), 1*disk.MB, 1000*disk.MB)
		cfg.Options().FilesQuota = 0

		assert.True(t, convert.insufficientStorage())
	})
}

func TestConvert_ToImage_InsufficientStorage(t *testing.T) {
	cfg := config.TestConfig()
	convert := NewConvert(cfg)

	// Use a RAW source so ToImage actually attempts a write
	// (preview images and missing files short-circuit before the storage gate).
	mf, err := NewMediaFile(cfg.SamplesPath() + "/canon_eos_6d.dng")
	require.NoError(t, err)

	disk.FlushFree()
	t.Cleanup(disk.FlushFree)
	disk.SetFree(cfg.StoragePath(), 1*disk.MB, 1000*disk.MB)
	config.DisableStorageCheck.Store(false)

	_, err = convert.ToImage(mf, false)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, status.ErrInsufficientStorage), "expected status.ErrInsufficientStorage, got %v", err)
}

func TestConvert_ToAvc_InsufficientStorage(t *testing.T) {
	cfg := config.TestConfig()
	convert := NewConvert(cfg)

	mf, err := NewMediaFile(cfg.SamplesPath() + "/gopher-video.mp4")
	require.NoError(t, err)

	disk.FlushFree()
	t.Cleanup(disk.FlushFree)
	disk.SetFree(cfg.StoragePath(), 1*disk.MB, 1000*disk.MB)
	config.DisableStorageCheck.Store(false)

	_, err = convert.ToAvc(mf, cfg.FFmpegEncoder(), false, false)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, status.ErrInsufficientStorage), "expected status.ErrInsufficientStorage, got %v", err)
}
