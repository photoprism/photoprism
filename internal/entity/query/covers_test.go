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

	assert.NoError(t, UpdateAlbumManualCovers())
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

func TestRefreshManualAlbumCoverPrivate(t *testing.T) {
	// A cover is published as a file hash that clients resolve through the thumbnail endpoint,
	// which applies no privacy filter, so a private picture must never become the cover.
	var album entity.Album

	if err := UnscopedDb().Where("album_type = ? AND deleted_at IS NULL", entity.AlbumManual).First(&album).Error; err != nil {
		t.Skipf("no manual album available: %v", err)
	}

	file, err := AlbumCoverByUID(album.AlbumUID, false)

	if err != nil {
		t.Skipf("album %s has no cover candidate: %v", album.AlbumUID, err)
	}

	photo := entity.Photo{}
	require.NoError(t, UnscopedDb().Where("id = ?", file.PhotoID).First(&photo).Error)

	origThumb, origSrc, origPrivate := album.Thumb, album.ThumbSrc, photo.PhotoPrivate

	t.Cleanup(func() {
		_ = UnscopedDb().Model(&entity.Photo{}).Where("id = ?", file.PhotoID).Update("photo_private", origPrivate).Error
		_ = entity.UpdateAlbum(album.AlbumUID, entity.Values{"thumb": origThumb, "thumb_src": origSrc})
		entity.FlushAlbumCache()
	})

	require.NoError(t, UnscopedDb().Model(&entity.Photo{}).Where("id = ?", file.PhotoID).Update("photo_private", true).Error)
	require.NoError(t, entity.UpdateAlbum(album.AlbumUID, entity.Values{"thumb": "", "thumb_src": entity.SrcAuto}))
	entity.FlushAlbumCache()

	require.NoError(t, refreshManualAlbumCover(album))
	entity.FlushAlbumCache()

	refreshed, err := AlbumByUID(album.AlbumUID)
	require.NoError(t, err)
	assert.NotEqual(t, file.FileHash, refreshed.Thumb)
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
	// Drain the async goroutine so it doesn't race the next test that
	// might temporarily swap out the entity DB provider.
	entity.WaitForAsyncJobs()
	assert.NoError(t, UpdateCovers())
}

func TestUpdateCovers_NilDbReturnsCleanly(t *testing.T) {
	// Mirrors TestUpdateCounts_NilDbReturnsCleanly: after CloseDb has
	// nilled the entity DB provider, UpdateCovers must return nil instead
	// of panicking on a nil dialect lookup so a stray UpdateCoversAsync
	// goroutine does not crash the process during shutdown.
	prev := swapDbProvider(nil)
	defer swapDbProvider(prev)

	assert.NoError(t, UpdateCovers())
}

// swapDbProvider replaces the package-level entity DB provider with the
// supplied value and returns a snapshot of the previous one wrapped in
// staticDbProvider so callers can restore the original *gorm.DB. The
// query package's staticDbProvider helper is reused to mirror existing
// override patterns in faces_test.go.
func swapDbProvider(p entity.Gorm) entity.Gorm {
	var prev entity.Gorm
	if currentDb := entity.Db(); currentDb != nil {
		prev = staticDbProvider{db: currentDb}
	}
	entity.SetDbProvider(p)
	return prev
}
