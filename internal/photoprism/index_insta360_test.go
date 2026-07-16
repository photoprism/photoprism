package photoprism

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/media"
)

// TestReconcileInsta360Photos verifies force-reindex state preservation and file reassignment.
func TestReconcileInsta360Photos(t *testing.T) {
	t.Setenv("PHOTOPRISM_TEST_DSN", filepath.Join(t.TempDir(), "index-insta360-reconcile.db"))
	cfg := config.NewMinimalTestConfigWithDb("index-insta360-reconcile", filepath.Join(t.TempDir(), "storage"))
	oldCfg := Config()
	SetConfig(cfg)
	t.Cleanup(func() {
		SetConfig(oldCfg)
		oldCfg.RegisterDb()
	})

	dir := filepath.Join(cfg.OriginalsPath(), "capture")
	leftName := writeInsta360CaptureFile(t, dir, "VID_20220625_140410_00_008.insv", "testdata/flash.jpg")
	rightName := writeInsta360CaptureFile(t, dir, "VID_20220625_140410_10_008.insv", "testdata/flash.jpg")
	proxyName := writeInsta360CaptureFile(t, dir, "LRV_20220625_140410_11_008.insv", "testdata/flash.jpg")

	left, err := NewMediaFile(leftName)
	require.NoError(t, err)
	related, err := left.RelatedFiles(false)
	require.NoError(t, err)

	photos := make([]entity.Photo, 3)
	for i := range photos {
		photos[i] = entity.NewPhoto(true)
		photos[i].PhotoTitle = ""
		require.NoError(t, photos[i].Create())
	}

	archivedAt := entity.Now()
	photos[0].DeletedAt = &archivedAt
	photos[1].DeletedAt = &archivedAt
	photos[1].PhotoFavorite = true
	photos[2].PhotoCaption = "preserved manual caption"
	photos[2].CaptionSrc = entity.SrcManual
	for i := range photos {
		require.NoError(t, photos[i].Save())
	}

	for i, fileName := range []string{leftName, rightName, proxyName} {
		mediaFile, mediaErr := NewMediaFile(fileName)
		require.NoError(t, mediaErr)
		file := entity.File{
			PhotoID:   photos[i].ID,
			PhotoUID:  photos[i].PhotoUID,
			FileName:  mediaFile.RootRelName(),
			FileRoot:  entity.RootOriginals,
			FileHash:  mediaFile.Hash(),
			FileType:  mediaFile.FileType().String(),
			MediaType: media.Video.String(),
			FileVideo: true,
		}
		require.NoError(t, file.Create())
	}

	require.NoError(t, reconcileInsta360Photos(related))

	var canonical entity.Photo
	require.NoError(t, entity.UnscopedDb().First(&canonical, "id = ?", photos[0].ID).Error)
	assert.True(t, canonical.PhotoFavorite)
	assert.Nil(t, canonical.DeletedAt)
	assert.Equal(t, "preserved manual caption", canonical.PhotoCaption)
	assert.Equal(t, entity.SrcManual, canonical.CaptionSrc)

	var files []entity.File
	require.NoError(t, entity.UnscopedDb().Where("file_name IN (?)", []string{
		left.RootRelName(),
		RootRelName(rightName),
		RootRelName(proxyName),
	}).Find(&files).Error)
	require.Len(t, files, 3)
	for _, file := range files {
		assert.Equal(t, canonical.ID, file.PhotoID)
		assert.Equal(t, canonical.PhotoUID, file.PhotoUID)
	}
}
