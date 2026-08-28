package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
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
		require.NoError(t, UnscopedDb().Model(&entity.Album{}).Where("id = ?", entity.AlbumFixtures.Get("april-1990").ID).UpdateColumns(entity.Values{"thumb": "justtestdata", "thumb_src": entity.SrcAuto}).Error)
		require.NoError(t, UnscopedDb().Model(&entity.Photo{}).Where("id = ?", entity.PhotoFixtures.Get("pho44to").ID).UpdateColumns(entity.Values{"photo_year": 1990, "photo_month": 4}).Error)
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
		require.NoError(t, UnscopedDb().Save(entity.AlbumFixtures.Pointer("september-2021")).Error)
		require.NoError(t, UnscopedDb().Model(&entity.Album{}).Where("id = ?", entity.AlbumFixtures.Get("september-2021").ID).UpdateColumns(entity.Values{"thumb": "justtestdata", "thumb_src": entity.SrcAuto}).Error)
		require.NoError(t, UnscopedDb().Model(&entity.Photo{}).Where("id = ?", entity.PhotoFixtures.Get("pho44to").ID).UpdateColumns(entity.Values{"photo_year": 2021, "photo_month": 9}).Error)
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

// Files the cover tests attach their markers to: bridge.jpg belongs to a public picture,
// reunion.jpg to a private one.
const (
	coverPublicFileUID  = "fs6sg6bw45bn0004"
	coverPrivateFileUID = "fs6sg6bw45bn0005"
)

// newCoverSubject creates an auto-managed person with the given cover and removes it afterwards.
func newCoverSubject(t *testing.T, thumb, thumbSrc string) *entity.Subject {
	t.Helper()

	subj := &entity.Subject{
		SubjUID:  rnd.GenerateUID('j'),
		SubjType: entity.SubjPerson,
		SubjSrc:  entity.SrcManual,
		SubjName: "Cover Test " + rnd.GenerateUID('j'),
		Thumb:    thumb,
		ThumbSrc: thumbSrc,
	}

	require.NoError(t, entity.Db().Create(subj).Error)
	t.Cleanup(func() { entity.UnscopedDb().Delete(&entity.Subject{}, "subj_uid = ?", subj.SubjUID) })

	return subj
}

// newCoverMarker creates a cover candidate, filling in the fields the caller left unset, and
// removes it afterwards.
func newCoverMarker(t *testing.T, m entity.Marker) *entity.Marker {
	t.Helper()

	if m.MarkerUID == "" {
		m.MarkerUID = rnd.GenerateUID('m')
	}

	if m.MarkerType == "" {
		m.MarkerType = entity.MarkerFace
	}

	if m.FileUID == "" {
		m.FileUID = coverPublicFileUID
	}

	m.MarkerSrc = entity.SrcImage
	m.X, m.Y, m.W, m.H = 0.1, 0.1, 0.2, 0.2

	require.NoError(t, entity.Db().Create(&m).Error)
	t.Cleanup(func() { entity.UnscopedDb().Delete(&entity.Marker{}, "marker_uid = ?", m.MarkerUID) })

	return &m
}

// coverThumb returns the cover currently stored for a person, and fails if the column is null,
// which GORM would otherwise scan into an empty string.
func coverThumb(t *testing.T, subjUID string) string {
	t.Helper()

	var subj entity.Subject
	var nulls int64

	require.NoError(t, UnscopedDb().Where("subj_uid = ?", subjUID).First(&subj).Error)
	require.NoError(t, UnscopedDb().Model(entity.Subject{}).Where("subj_uid = ? AND thumb IS NULL", subjUID).Count(&nulls).Error)
	require.Zero(t, nulls, "cover must not be null")

	return subj.Thumb
}

func TestUpdateSubjectCovers(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert.NoError(t, UpdateSubjectCovers(false))
		assert.NoError(t, UpdateSubjectCovers(true))
	})
	t.Run("PicksTheLargestFace", func(t *testing.T) {
		subj := newCoverSubject(t, "", entity.SrcAuto)
		// The thumbs sort in the reverse of the size order, so a cover picked by thumb fails here.
		newCoverMarker(t, entity.Marker{SubjUID: subj.SubjUID, Size: 60, Score: 95, Thumb: "zzz-" + subj.SubjUID})
		newCoverMarker(t, entity.Marker{SubjUID: subj.SubjUID, Size: 200, Score: 70, Thumb: "aaa-" + subj.SubjUID})
		newCoverMarker(t, entity.Marker{SubjUID: subj.SubjUID, Size: 120, Score: 80, Thumb: "mmm-" + subj.SubjUID})
		require.NoError(t, UpdateSubjectCovers(true))
		assert.Equal(t, "aaa-"+subj.SubjUID, coverThumb(t, subj.SubjUID))
	})
	t.Run("PrefersAnAssignedSubject", func(t *testing.T) {
		subj := newCoverSubject(t, "", entity.SrcAuto)
		newCoverMarker(t, entity.Marker{SubjUID: subj.SubjUID, SubjSrc: entity.SrcManual, Size: 60, Score: 70, Thumb: "manual-" + subj.SubjUID})
		newCoverMarker(t, entity.Marker{SubjUID: subj.SubjUID, SubjSrc: entity.SrcAuto, Size: 400, Score: 95, Thumb: "auto-" + subj.SubjUID})
		require.NoError(t, UpdateSubjectCovers(true))
		assert.Equal(t, "manual-"+subj.SubjUID, coverThumb(t, subj.SubjUID))
	})
	t.Run("RanksAssignedSubjectsBySize", func(t *testing.T) {
		// Every source other than automatic clustering shares one rank, so size decides between
		// them. A byte sort over the source string would take the sidecar name here instead.
		subj := newCoverSubject(t, "", entity.SrcAuto)
		newCoverMarker(t, entity.Marker{SubjUID: subj.SubjUID, SubjSrc: entity.SrcXmp, Size: 60, Score: 95, Thumb: "xmp-" + subj.SubjUID})
		newCoverMarker(t, entity.Marker{SubjUID: subj.SubjUID, SubjSrc: entity.SrcManual, Size: 200, Score: 70, Thumb: "manual-" + subj.SubjUID})
		require.NoError(t, UpdateSubjectCovers(true))
		assert.Equal(t, "manual-"+subj.SubjUID, coverThumb(t, subj.SubjUID))
	})
	t.Run("RanksASidecarNameByTheSameRule", func(t *testing.T) {
		// The converse, and the case that decides what "one rank" means: a larger face named from
		// a sidecar takes the cover from a smaller one a person typed. Without this, reintroducing
		// a precedence between the two sources would pass every other subtest here.
		subj := newCoverSubject(t, "", entity.SrcAuto)
		newCoverMarker(t, entity.Marker{SubjUID: subj.SubjUID, SubjSrc: entity.SrcManual, Size: 60, Score: 95, Thumb: "manual-" + subj.SubjUID})
		newCoverMarker(t, entity.Marker{SubjUID: subj.SubjUID, SubjSrc: entity.SrcXmp, Size: 200, Score: 70, Thumb: "xmp-" + subj.SubjUID})
		require.NoError(t, UpdateSubjectCovers(true))
		assert.Equal(t, "xmp-"+subj.SubjUID, coverThumb(t, subj.SubjUID))
	})
	t.Run("PicksTheMostConfidentOfEqualSize", func(t *testing.T) {
		subj := newCoverSubject(t, "", entity.SrcAuto)
		newCoverMarker(t, entity.Marker{SubjUID: subj.SubjUID, Size: 200, Score: 70, Thumb: "zzz-" + subj.SubjUID})
		newCoverMarker(t, entity.Marker{SubjUID: subj.SubjUID, Size: 200, Score: 95, Thumb: "aaa-" + subj.SubjUID})
		require.NoError(t, UpdateSubjectCovers(true))
		assert.Equal(t, "aaa-"+subj.SubjUID, coverThumb(t, subj.SubjUID))
	})
	t.Run("IgnoresInvalidMarkers", func(t *testing.T) {
		subj := newCoverSubject(t, "", entity.SrcAuto)
		newCoverMarker(t, entity.Marker{SubjUID: subj.SubjUID, MarkerInvalid: true, Size: 400, Score: 95, Thumb: "invalid-" + subj.SubjUID})
		newCoverMarker(t, entity.Marker{SubjUID: subj.SubjUID, Size: 120, Score: 80, Thumb: "valid-" + subj.SubjUID})
		require.NoError(t, UpdateSubjectCovers(true))
		assert.Equal(t, "valid-"+subj.SubjUID, coverThumb(t, subj.SubjUID))
	})
	t.Run("IgnoresNonFaceMarkers", func(t *testing.T) {
		subj := newCoverSubject(t, "", entity.SrcAuto)
		newCoverMarker(t, entity.Marker{SubjUID: subj.SubjUID, MarkerType: entity.MarkerLabel, Size: 400, Score: 95, Thumb: "label-" + subj.SubjUID})
		newCoverMarker(t, entity.Marker{SubjUID: subj.SubjUID, Size: 120, Score: 80, Thumb: "face-" + subj.SubjUID})
		require.NoError(t, UpdateSubjectCovers(true))
		assert.Equal(t, "face-"+subj.SubjUID, coverThumb(t, subj.SubjUID))
	})
	t.Run("ClearsTheCoverWhenNothingIsEligible", func(t *testing.T) {
		// A cover resolves to a file hash that the thumbnail endpoint serves without a privacy
		// filter, so a person whose every picture turned private must not keep the crop.
		subj := newCoverSubject(t, "stale-"+rnd.GenerateUID('j'), entity.SrcAuto)
		newCoverMarker(t, entity.Marker{SubjUID: subj.SubjUID, FileUID: coverPrivateFileUID, Size: 400, Score: 95, Thumb: "private-" + subj.SubjUID})
		require.NoError(t, UpdateSubjectCovers(true))
		assert.Empty(t, coverThumb(t, subj.SubjUID))
		// The same marker is eligible once private pictures count, so only privacy excluded it.
		require.NoError(t, UpdateSubjectCovers(false))
		assert.Equal(t, "private-"+subj.SubjUID, coverThumb(t, subj.SubjUID))
	})
	t.Run("Deterministic", func(t *testing.T) {
		subj := newCoverSubject(t, "", entity.SrcAuto)
		first := newCoverMarker(t, entity.Marker{SubjUID: subj.SubjUID, Size: 200, Score: 80, Thumb: "one-" + subj.SubjUID})
		second := newCoverMarker(t, entity.Marker{SubjUID: subj.SubjUID, Size: 200, Score: 80, Thumb: "two-" + subj.SubjUID})
		require.NoError(t, UpdateSubjectCovers(true))
		picked := coverThumb(t, subj.SubjUID)
		require.NoError(t, UpdateSubjectCovers(true))
		assert.Equal(t, picked, coverThumb(t, subj.SubjUID))
		if first.MarkerUID < second.MarkerUID {
			assert.Equal(t, first.Thumb, picked)
		} else {
			assert.Equal(t, second.Thumb, picked)
		}
	})
	t.Run("LeavesAManualCoverAlone", func(t *testing.T) {
		subj := newCoverSubject(t, "chosen-"+rnd.GenerateUID('j'), entity.SrcManual)
		newCoverMarker(t, entity.Marker{SubjUID: subj.SubjUID, Size: 400, Score: 95, Thumb: "better-" + subj.SubjUID})
		require.NoError(t, UpdateSubjectCovers(true))
		assert.Equal(t, subj.Thumb, coverThumb(t, subj.SubjUID))
	})
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
