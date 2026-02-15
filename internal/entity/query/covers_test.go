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

	origThumb := album.Thumb
	origSrc := album.ThumbSrc

	if err := UnscopedDb().Where("album_type = ? AND thumb_src = ? AND album_path <> '' AND thumb <> ''", entity.AlbumFolder, entity.SrcAuto).First(&album).Error; err != nil {
		// Make the data look like it is needed for the test.  Updating the fixtures directly breaks to many other tests.
		require.NoError(t, Db().Save(entity.AlbumFixtures.Pointer("april-1990")).Error)
		require.NoError(t, UnscopedDb().Model(entity.Album{}).Where("id = ?", entity.AlbumFixtures.Get("april-1990").ID).UpdateColumns(entity.Values{"thumb": "justtestdata", "thumb_src": entity.SrcAuto}).Error)
		require.NoError(t, UnscopedDb().Model(entity.Photo{}).Where("id = ?", entity.PhotoFixtures.Get("pho44to").ID).UpdateColumns(entity.Values{"photo_year": 1990, "photo_month": 4}).Error)
		require.NoError(t, UnscopedDb().Where("album_type = ? AND thumb_src = ? AND album_path <> '' AND thumb <> ''", entity.AlbumFolder, entity.SrcAuto).First(&album).Error)
		origThumb = entity.AlbumFixtures.Get("april-1990").Thumb
		origSrc = entity.AlbumFixtures.Get("april-1990").ThumbSrc
	} else {
		origThumb = album.Thumb
		origSrc = album.ThumbSrc
	}

	t.Cleanup(func() {
		require.NoError(t, Db().Save(entity.AlbumFixtures.Pointer("april-1990")).Error)
		require.NoError(t, Db().Save(entity.PhotoFixtures.Pointer("pho44to")).Error)
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

	origThumb := album.Thumb
	origSrc := album.ThumbSrc

	if err := UnscopedDb().Where("album_type = ? AND thumb_src = ? AND album_year <> 0 AND thumb <> ''", entity.AlbumMonth, entity.SrcAuto).First(&album).Error; err != nil {
		// Make the data look like it is needed for the test.  Updating the fixtures directly breaks to many other tests.
		require.NoError(t, Db().Save(entity.AlbumFixtures.Pointer("september-2021")).Error)
		require.NoError(t, UnscopedDb().Model(entity.Album{}).Where("id = ?", entity.AlbumFixtures.Get("september-2021").ID).UpdateColumns(entity.Values{"thumb": "justtestdata", "thumb_src": entity.SrcAuto}).Error)
		require.NoError(t, UnscopedDb().Model(entity.Photo{}).Where("id = ?", entity.PhotoFixtures.Get("pho44to").ID).UpdateColumns(entity.Values{"photo_year": 2021, "photo_month": 9}).Error)
		require.NoError(t, UnscopedDb().Where("album_type = ? AND thumb_src = ? AND album_year <> 0 AND thumb <> ''", entity.AlbumMonth, entity.SrcAuto).First(&album).Error)
		origThumb = entity.AlbumFixtures.Get("september-2021").Thumb
		origSrc = entity.AlbumFixtures.Get("september-2021").ThumbSrc
	} else {
		origThumb = album.Thumb
		origSrc = album.ThumbSrc
	}

	t.Cleanup(func() {
		require.NoError(t, Db().Save(entity.AlbumFixtures.Pointer("september-2021")).Error)
		require.NoError(t, Db().Save(entity.PhotoFixtures.Pointer("pho44to")).Error)
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
