package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestUpdateAlbumManualCovers(t *testing.T) {
	assert.NoError(t, UpdateAlbumManualCovers())
}

func TestUpdateAlbumManualCoversFiltered(t *testing.T) {
	var album entity.Album

	if err := UnscopedDb().Where("album_type = ? AND thumb_src = ? AND thumb <> ''", entity.AlbumManual, entity.SrcAuto).First(&album).Error; err != nil {
		t.Skipf("no auto-managed manual album available: %v", err)
	}

	origThumb := album.Thumb
	origSrc := album.ThumbSrc

	t.Cleanup(func() {
		_ = entity.UpdateAlbum(album.AlbumUID, entity.Values{"thumb": origThumb, "thumb_src": origSrc})
		entity.FlushAlbumCache()
	})

	require.NoError(t, entity.UpdateAlbum(album.AlbumUID, entity.Values{"thumb": "", "thumb_src": entity.SrcAuto}))
	entity.FlushAlbumCache()

	require.NoError(t, UpdateAlbumManualCovers(album))
	entity.FlushAlbumCache()

	refreshed, err := AlbumByUID(album.AlbumUID)
	require.NoError(t, err)
	assert.NotEmpty(t, refreshed.Thumb)
}

func TestUpdateAlbumFolderCovers(t *testing.T) {
	assert.NoError(t, UpdateAlbumFolderCovers())
}

func TestUpdateAlbumFolderCoversFiltered(t *testing.T) {
	var album entity.Album

	if err := UnscopedDb().Where("album_type = ? AND thumb_src = ? AND album_path <> '' AND thumb <> ''", entity.AlbumFolder, entity.SrcAuto).First(&album).Error; err != nil {
		t.Skipf("no auto-managed folder album available: %v", err)
	}

	origThumb := album.Thumb
	origSrc := album.ThumbSrc

	t.Cleanup(func() {
		_ = entity.UpdateAlbum(album.AlbumUID, entity.Values{"thumb": origThumb, "thumb_src": origSrc})
		entity.FlushAlbumCache()
	})

	require.NoError(t, entity.UpdateAlbum(album.AlbumUID, entity.Values{"thumb": "", "thumb_src": entity.SrcAuto}))
	entity.FlushAlbumCache()

	require.NoError(t, UpdateAlbumFolderCovers(album))
	entity.FlushAlbumCache()

	refreshed, err := AlbumByUID(album.AlbumUID)
	require.NoError(t, err)
	assert.NotEmpty(t, refreshed.Thumb)
}

func TestUpdateAlbumMonthCovers(t *testing.T) {
	assert.NoError(t, UpdateAlbumMonthCovers())
}

func TestUpdateAlbumMonthCoversFiltered(t *testing.T) {
	var album entity.Album

	if err := UnscopedDb().Where("album_type = ? AND thumb_src = ? AND album_year <> 0 AND thumb <> ''", entity.AlbumMonth, entity.SrcAuto).First(&album).Error; err != nil {
		t.Skipf("no auto-managed monthly album available: %v", err)
	}

	origThumb := album.Thumb
	origSrc := album.ThumbSrc

	t.Cleanup(func() {
		_ = entity.UpdateAlbum(album.AlbumUID, entity.Values{"thumb": origThumb, "thumb_src": origSrc})
		entity.FlushAlbumCache()
	})

	require.NoError(t, entity.UpdateAlbum(album.AlbumUID, entity.Values{"thumb": "", "thumb_src": entity.SrcAuto}))
	entity.FlushAlbumCache()

	require.NoError(t, UpdateAlbumMonthCovers(album))
	entity.FlushAlbumCache()

	refreshed, err := AlbumByUID(album.AlbumUID)
	require.NoError(t, err)
	assert.NotEmpty(t, refreshed.Thumb)
}

func TestUpdateAlbumCovers(t *testing.T) {
	assert.NoError(t, UpdateAlbumCovers())
}

func TestUpdateLabelCovers(t *testing.T) {
	assert.NoError(t, UpdateLabelCovers())
}

func TestUpdateSubjectCovers(t *testing.T) {
	assert.NoError(t, UpdateSubjectCovers(false))
	assert.NoError(t, UpdateSubjectCovers(true))
}

func TestUpdateCovers(t *testing.T) {
	// coversBusy.Store(true)
	UpdateCoversAsync()
	assert.NoError(t, UpdateCovers())
}
