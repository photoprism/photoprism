package photoprism

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/media"
)

// newInsta360ReconcileFixture indexes one capture as three unrelated photos, as an older index run
// would have left them, and returns them in canonical lens order.
// The capture directory is named after the test case because MariaDB runs share one database per
// package. Do not isolate it with PHOTOPRISM_TEST_DSN: that points the driver at MySQL and kills
// the test binary, see entity.TestDbDSN.
func newInsta360ReconcileFixture(t *testing.T, name string) (RelatedFiles, []entity.Photo, []string) {
	t.Helper()
	cfg := config.NewMinimalTestConfigWithDb(name, filepath.Join(t.TempDir(), "storage"))
	oldCfg := Config()
	SetConfig(cfg)
	t.Cleanup(func() {
		SetConfig(oldCfg)
		oldCfg.RegisterDb()
	})

	dir := filepath.Join(cfg.OriginalsPath(), name)
	fileNames := []string{
		writeInsta360CaptureFile(t, dir, "VID_20220625_140410_00_008.insv", "testdata/flash.jpg"),
		writeInsta360CaptureFile(t, dir, "VID_20220625_140410_10_008.insv", "testdata/flash.jpg"),
		writeInsta360CaptureFile(t, dir, "LRV_20220625_140410_11_008.insv", "testdata/flash.jpg"),
	}

	left, err := NewMediaFile(fileNames[0])
	require.NoError(t, err)
	related, err := left.RelatedFiles(false)
	require.NoError(t, err)

	photos := make([]entity.Photo, 3)
	for i := range photos {
		photos[i] = entity.NewPhoto(true)
		photos[i].PhotoTitle = ""
		require.NoError(t, photos[i].Create())
	}

	relNames := make([]string, 0, len(fileNames))
	for i, fileName := range fileNames {
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
		relNames = append(relNames, file.FileName)
	}

	return related, photos, relNames
}

// linkInsta360Associations adds one album, label, and keyword relation to the specified photo.
func linkInsta360Associations(t *testing.T, photo entity.Photo, albumUID string, labelID, keywordID uint) {
	t.Helper()
	require.NoError(t, entity.NewPhotoAlbum(photo.PhotoUID, albumUID).Create())
	require.NoError(t, entity.NewPhotoLabel(photo.ID, labelID, 100, entity.SrcManual).Create())
	require.NoError(t, entity.NewPhotoKeyword(photo.ID, keywordID).Create())
}

// insta360AlbumUIDs returns the album UIDs currently linked to the specified photo UID.
func insta360AlbumUIDs(t *testing.T, photoUID string) []string {
	t.Helper()
	result := make([]string, 0, 2)
	require.NoError(t, entity.UnscopedDb().Model(&entity.PhotoAlbum{}).
		Where("photo_uid = ?", photoUID).Pluck("album_uid", &result).Error)
	return result
}

// insta360LabelIDs returns the label IDs currently linked to the specified photo.
func insta360LabelIDs(t *testing.T, photoID uint) []uint {
	t.Helper()
	result := make([]uint, 0, 2)
	require.NoError(t, entity.UnscopedDb().Model(&entity.PhotoLabel{}).
		Where("photo_id = ?", photoID).Pluck("label_id", &result).Error)
	return result
}

// insta360KeywordIDs returns the keyword IDs currently linked to the specified photo.
func insta360KeywordIDs(t *testing.T, photoID uint) []uint {
	t.Helper()
	result := make([]uint, 0, 2)
	require.NoError(t, entity.UnscopedDb().Model(&entity.PhotoKeyword{}).
		Where("photo_id = ?", photoID).Pluck("keyword_id", &result).Error)
	return result
}

// setInsta360PhotoState sets the archive state and quality of a photo without side effects.
func setInsta360PhotoState(t *testing.T, photo entity.Photo, deletedAt *time.Time, quality int) {
	t.Helper()
	require.NoError(t, entity.UnscopedDb().Model(&entity.Photo{}).Where("id = ?", photo.ID).
		UpdateColumns(entity.Values{"deleted_at": deletedAt, "photo_quality": quality}).Error)
}

// findInsta360Photo returns the photo with the specified ID, including deleted rows.
func findInsta360Photo(t *testing.T, id uint) (result entity.Photo) {
	t.Helper()
	require.NoError(t, entity.UnscopedDb().First(&result, "id = ?", id).Error)
	return result
}

// TestInsta360PhotoRemoved verifies that only deleted rows at quality -1 count as removed.
func TestInsta360PhotoRemoved(t *testing.T) {
	deletedAt := entity.Now()
	assert.True(t, insta360PhotoRemoved(&entity.Photo{DeletedAt: &deletedAt, PhotoQuality: -1}))
	assert.False(t, insta360PhotoRemoved(&entity.Photo{DeletedAt: &deletedAt, PhotoQuality: 0}))
	assert.False(t, insta360PhotoRemoved(&entity.Photo{PhotoQuality: -1}))
	assert.False(t, insta360PhotoRemoved(&entity.Photo{}))
	assert.False(t, insta360PhotoRemoved(nil))
}

// TestInsta360CaptureState verifies which photo decides the archive state of a merged capture.
func TestInsta360CaptureState(t *testing.T) {
	removedAt := entity.Now()
	before, after := removedAt.Add(-time.Minute), removedAt.Add(time.Minute)
	photo := func(id uint, deletedAt *time.Time, quality int) *entity.Photo {
		return &entity.Photo{ID: id, CreatedAt: before, DeletedAt: deletedAt, PhotoQuality: quality}
	}
	created := func(p *entity.Photo, createdAt time.Time) *entity.Photo {
		p.CreatedAt = createdAt
		return p
	}
	later := after.Add(time.Minute)

	cases := []struct {
		name   string
		photos entity.Photos
		expect uint
	}{
		{"Active", entity.Photos{photo(1, nil, 3), photo(2, &after, 3)}, 1},
		{"Archived", entity.Photos{photo(1, &removedAt, 3), photo(2, nil, 3)}, 1},
		{"RemovedActiveMember", entity.Photos{photo(1, &removedAt, -1), photo(2, nil, 3)}, 2},
		{"ActiveWinsOverArchived", entity.Photos{photo(1, &removedAt, -1), photo(2, &after, 3), photo(3, nil, 3)}, 3},
		{"ArchivedAfterRemoval", entity.Photos{photo(1, &removedAt, -1), photo(2, &before, 3), photo(3, &after, 4)}, 3},
		{"SkipsRemovedMembers", entity.Photos{photo(1, &removedAt, -1), photo(2, &after, -1), photo(3, &after, 4)}, 3},
		{"SameSecond", entity.Photos{photo(1, &removedAt, -1), photo(2, &removedAt, 4)}, 2},
		{"ArchivedBeforeRemoval", entity.Photos{photo(1, &removedAt, -1), photo(2, &before, 3)}, 1},
		{"AllRemoved", entity.Photos{photo(1, &removedAt, -1), photo(2, &after, -1)}, 1},
		{"FirstArchivedAfterRemoval", entity.Photos{photo(1, &removedAt, -1), photo(2, &after, 4), photo(3, &later, 5)}, 2},
		{"HiddenIsNotVisible", entity.Photos{photo(1, &removedAt, -1), photo(2, nil, -1), photo(3, &after, 4)}, 3},
		{"HiddenOnly", entity.Photos{photo(1, &removedAt, -1), photo(2, nil, -1)}, 1},
		{"VisibleCreatedLater", entity.Photos{photo(1, &removedAt, -1), photo(2, &after, 4), created(photo(3, nil, 3), later)}, 2},
		{"VisibleCreatedSameSecond", entity.Photos{photo(1, &removedAt, -1), photo(2, &after, 4), created(photo(3, nil, 3), after)}, 2},
		{"VisibleCreatedLaterOnly", entity.Photos{photo(1, &removedAt, -1), created(photo(2, nil, 3), later)}, 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := insta360CaptureState(tc.photos)
			require.NotNil(t, result)
			assert.Equal(t, tc.expect, result.ID)
		})
	}
	t.Run("Empty", func(t *testing.T) {
		assert.Nil(t, insta360CaptureState(nil))
		assert.Nil(t, insta360CaptureState(entity.Photos{nil}))
		assert.Equal(t, uint(1), insta360CaptureState(entity.Photos{photo(1, &removedAt, -1), nil}).ID)
	})
}

// TestReconcileInsta360Photos verifies force-reindex state preservation and file reassignment.
func TestReconcileInsta360Photos(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		related, photos, relNames := newInsta360ReconcileFixture(t, "insta360reconcile")

		archivedAt := entity.Now()
		photos[1].DeletedAt = &archivedAt
		photos[1].PhotoFavorite = true
		photos[1].PhotoTitle = "Manual Title"
		photos[1].TitleSrc = entity.SrcManual
		photos[2].PhotoPrivate = true
		photos[2].PhotoPanorama = true
		photos[2].PhotoCaption = "preserved manual caption"
		photos[2].CaptionSrc = entity.SrcManual
		for i := range photos {
			require.NoError(t, photos[i].Save())
		}
		require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).Where("file_name IN (?)", relNames).
			UpdateColumn("file_primary", true).Error)

		sharedAlbum := entity.NewAlbum("Shared Album", entity.AlbumManual)
		require.NoError(t, sharedAlbum.Create())
		uniqueAlbum := entity.NewAlbum("Unique Album", entity.AlbumManual)
		require.NoError(t, uniqueAlbum.Create())
		sharedLabel := entity.NewLabel("Shared Label", 0)
		require.NoError(t, sharedLabel.Create())
		uniqueLabel := entity.NewLabel("Unique Label", 0)
		require.NoError(t, uniqueLabel.Create())
		sharedKeyword := entity.NewKeyword("shared")
		require.NoError(t, sharedKeyword.Create())
		uniqueKeyword := entity.NewKeyword("unique")
		require.NoError(t, uniqueKeyword.Create())

		// The canonical photo and the first duplicate deliberately share their relations so the
		// merge has to survive primary key conflicts, while the second duplicate contributes
		// relations that must survive the merge.
		linkInsta360Associations(t, photos[0], sharedAlbum.AlbumUID, sharedLabel.ID, sharedKeyword.ID)
		linkInsta360Associations(t, photos[1], sharedAlbum.AlbumUID, sharedLabel.ID, sharedKeyword.ID)
		linkInsta360Associations(t, photos[2], uniqueAlbum.AlbumUID, uniqueLabel.ID, uniqueKeyword.ID)

		logger := logrus.StandardLogger()
		oldHooks := make(logrus.LevelHooks, len(logger.Hooks))
		for level, hooks := range logger.Hooks {
			oldHooks[level] = append([]logrus.Hook(nil), hooks...)
		}
		hook := test.NewGlobal()
		t.Cleanup(func() { logger.ReplaceHooks(oldHooks) })

		require.NoError(t, reconcileInsta360Photos(related))

		var canonical entity.Photo
		require.NoError(t, entity.UnscopedDb().First(&canonical, "id = ?", photos[0].ID).Error)
		assert.True(t, canonical.PhotoFavorite)
		assert.True(t, canonical.PhotoPrivate)
		assert.True(t, canonical.PhotoPanorama)
		assert.Nil(t, canonical.DeletedAt)
		assert.Equal(t, "Manual Title", canonical.PhotoTitle)
		assert.Equal(t, entity.SrcManual, canonical.TitleSrc)
		assert.Equal(t, "preserved manual caption", canonical.PhotoCaption)
		assert.Equal(t, entity.SrcManual, canonical.CaptionSrc)

		var files []entity.File
		require.NoError(t, entity.UnscopedDb().Where("file_name IN (?)", relNames).Find(&files).Error)
		require.Len(t, files, 3)
		for _, file := range files {
			assert.Equal(t, canonical.ID, file.PhotoID)
			assert.Equal(t, canonical.PhotoUID, file.PhotoUID)
			assert.Equal(t, file.FileName == relNames[0], file.FilePrimary, file.FileName)
		}

		var merges []string
		for _, entry := range hook.AllEntries() {
			if entry.Level == logrus.InfoLevel && strings.HasPrefix(entry.Message, "index: merged ") {
				merges = append(merges, entry.Message)
			}
		}
		require.Len(t, merges, 1)
		assert.Contains(t, merges[0], photos[1].PhotoUID+", "+photos[2].PhotoUID+" into "+canonical.PhotoUID)

		assert.ElementsMatch(t, []string{sharedAlbum.AlbumUID, uniqueAlbum.AlbumUID}, insta360AlbumUIDs(t, canonical.PhotoUID))
		assert.ElementsMatch(t, []uint{sharedLabel.ID, uniqueLabel.ID}, insta360LabelIDs(t, canonical.ID))
		assert.ElementsMatch(t, []uint{sharedKeyword.ID, uniqueKeyword.ID}, insta360KeywordIDs(t, canonical.ID))
		for _, duplicate := range photos[1:] {
			assert.Empty(t, insta360AlbumUIDs(t, duplicate.PhotoUID))
			assert.Empty(t, insta360LabelIDs(t, duplicate.ID))
			assert.Empty(t, insta360KeywordIDs(t, duplicate.ID))
		}
	})
	t.Run("AllArchived", func(t *testing.T) {
		related, photos, _ := newInsta360ReconcileFixture(t, "insta360archived")

		archivedAt := entity.Now()
		for i := range photos {
			photos[i].DeletedAt = &archivedAt
			require.NoError(t, photos[i].Save())
		}

		require.NoError(t, reconcileInsta360Photos(related))

		var canonical entity.Photo
		require.NoError(t, entity.UnscopedDb().First(&canonical, "id = ?", photos[0].ID).Error)
		assert.NotNil(t, canonical.DeletedAt)
		for _, duplicate := range photos[1:] {
			var merged entity.Photo
			require.NoError(t, entity.UnscopedDb().First(&merged, "id = ?", duplicate.ID).Error)
			assert.NotNil(t, merged.DeletedAt)
			assert.Equal(t, -1, merged.PhotoQuality)
		}
	})
	t.Run("CanonicalArchived", func(t *testing.T) {
		related, photos, _ := newInsta360ReconcileFixture(t, "insta360canonicalarchived")
		archivedAt := entity.Now()
		setInsta360PhotoState(t, photos[0], &archivedAt, 3)

		require.NoError(t, reconcileInsta360Photos(related))

		canonical := findInsta360Photo(t, photos[0].ID)
		assert.NotNil(t, canonical.DeletedAt)
		assert.Equal(t, 3, canonical.PhotoQuality)
	})
	t.Run("MemberArchived", func(t *testing.T) {
		related, photos, _ := newInsta360ReconcileFixture(t, "insta360memberarchived")
		archivedAt := entity.Now()
		setInsta360PhotoState(t, photos[0], nil, 3)
		setInsta360PhotoState(t, photos[1], &archivedAt, 3)

		require.NoError(t, reconcileInsta360Photos(related))

		canonical := findInsta360Photo(t, photos[0].ID)
		assert.Nil(t, canonical.DeletedAt)
		assert.Equal(t, 3, canonical.PhotoQuality)
	})
	t.Run("MemberRemoved", func(t *testing.T) {
		related, photos, _ := newInsta360ReconcileFixture(t, "insta360memberremoved")
		removedAt := entity.Now()
		setInsta360PhotoState(t, photos[0], nil, 3)
		setInsta360PhotoState(t, photos[2], &removedAt, -1)

		require.NoError(t, reconcileInsta360Photos(related))

		canonical := findInsta360Photo(t, photos[0].ID)
		assert.Nil(t, canonical.DeletedAt)
		assert.Equal(t, 3, canonical.PhotoQuality)
	})
	t.Run("CanonicalRemoved", func(t *testing.T) {
		related, photos, _ := newInsta360ReconcileFixture(t, "insta360canonicalremoved")
		removedAt := entity.Now()
		setInsta360PhotoState(t, photos[0], &removedAt, -1)
		setInsta360PhotoState(t, photos[1], nil, 3)

		require.NoError(t, reconcileInsta360Photos(related))

		canonical := findInsta360Photo(t, photos[0].ID)
		assert.Nil(t, canonical.DeletedAt)
		assert.Equal(t, 3, canonical.PhotoQuality)
	})
	t.Run("CanonicalRemovedMemberArchived", func(t *testing.T) {
		related, photos, _ := newInsta360ReconcileFixture(t, "insta360removedarchived")
		removedAt := entity.Now()
		setInsta360PhotoState(t, photos[0], &removedAt, -1)
		setInsta360PhotoState(t, photos[1], &removedAt, -1)
		setInsta360PhotoState(t, photos[2], &removedAt, 4)

		require.NoError(t, reconcileInsta360Photos(related))

		canonical := findInsta360Photo(t, photos[0].ID)
		assert.NotNil(t, canonical.DeletedAt)
		assert.Equal(t, 4, canonical.PhotoQuality)
	})
	t.Run("MemberArchivedBeforeRemoval", func(t *testing.T) {
		related, photos, _ := newInsta360ReconcileFixture(t, "insta360archivedbefore")
		archivedAt := entity.Now().Add(-time.Minute)
		removedAt := entity.Now()
		setInsta360PhotoState(t, photos[0], &removedAt, -1)
		setInsta360PhotoState(t, photos[1], &archivedAt, 4)
		setInsta360PhotoState(t, photos[2], &removedAt, -1)

		require.NoError(t, reconcileInsta360Photos(related))

		// The indexer restores the photo when it indexes the capture again.
		canonical := findInsta360Photo(t, photos[0].ID)
		assert.NotNil(t, canonical.DeletedAt)
		assert.Equal(t, -1, canonical.PhotoQuality)
	})
	t.Run("AllRemoved", func(t *testing.T) {
		related, photos, _ := newInsta360ReconcileFixture(t, "insta360allremoved")
		removedAt := entity.Now()
		for _, photo := range photos {
			setInsta360PhotoState(t, photo, &removedAt, -1)
		}

		require.NoError(t, reconcileInsta360Photos(related))

		// The indexer restores the photo when it indexes the capture again.
		canonical := findInsta360Photo(t, photos[0].ID)
		assert.NotNil(t, canonical.DeletedAt)
		assert.Equal(t, -1, canonical.PhotoQuality)
	})
	t.Run("Mixed", func(t *testing.T) {
		related, photos, _ := newInsta360ReconcileFixture(t, "insta360mixed")
		archivedAt := entity.Now()
		setInsta360PhotoState(t, photos[0], &archivedAt, 3)
		setInsta360PhotoState(t, photos[1], nil, 3)
		setInsta360PhotoState(t, photos[2], &archivedAt, -1)

		require.NoError(t, reconcileInsta360Photos(related))

		canonical := findInsta360Photo(t, photos[0].ID)
		assert.NotNil(t, canonical.DeletedAt)
		assert.Equal(t, 3, canonical.PhotoQuality)
		for _, duplicate := range photos[1:] {
			merged := findInsta360Photo(t, duplicate.ID)
			assert.NotNil(t, merged.DeletedAt)
			assert.Equal(t, -1, merged.PhotoQuality)
		}
	})
	t.Run("IncompletePair", func(t *testing.T) {
		related, photos, relNames := newInsta360ReconcileFixture(t, "insta360incomplete")

		// Removing the right lens leaves an incomplete capture, which must not merge anything.
		require.NoError(t, entity.UnscopedDb().Where("file_name = ?", relNames[1]).Delete(&entity.File{}).Error)
		require.NoError(t, os.Remove(filepath.Join(Config().OriginalsPath(), relNames[1])))

		reduced, err := related.Main.RelatedFiles(false)
		require.NoError(t, err)
		require.NoError(t, reconcileInsta360Photos(reduced))

		for _, photo := range photos {
			var unchanged entity.Photo
			require.NoError(t, entity.UnscopedDb().First(&unchanged, "id = ?", photo.ID).Error)
			assert.Nil(t, unchanged.DeletedAt)
			assert.NotEqual(t, -1, unchanged.PhotoQuality)
		}
	})
}

// TestReconcileInsta360Photos_LensCodedPhotos verifies that photos with lens codes are never merged.
func TestReconcileInsta360Photos_LensCodedPhotos(t *testing.T) {
	name := "insta360reconcilephoto"
	cfg := config.NewMinimalTestConfigWithDb(name, filepath.Join(t.TempDir(), "storage"))
	oldCfg := Config()
	SetConfig(cfg)
	t.Cleanup(func() {
		SetConfig(oldCfg)
		oldCfg.RegisterDb()
	})

	dir := filepath.Join(cfg.OriginalsPath(), name)
	fileNames := []string{
		writeInsta360CaptureFile(t, dir, "IMG_20220625_140410_00_008.insp", "testdata/flash.jpg"),
		writeInsta360CaptureFile(t, dir, "IMG_20220625_140410_10_008.insp", "testdata/flash.jpg"),
	}

	photos := make([]entity.Photo, len(fileNames))

	for i, fileName := range fileNames {
		photos[i] = entity.NewPhoto(true)
		require.NoError(t, photos[i].Create())
		mediaFile, err := NewMediaFile(fileName)
		require.NoError(t, err)
		file := entity.File{
			PhotoID:   photos[i].ID,
			PhotoUID:  photos[i].PhotoUID,
			FileName:  mediaFile.RootRelName(),
			FileRoot:  entity.RootOriginals,
			FileHash:  mediaFile.Hash(),
			FileType:  mediaFile.FileType().String(),
			MediaType: media.Image.String(),
		}
		require.NoError(t, file.Create())
	}

	left, err := NewMediaFile(fileNames[0])
	require.NoError(t, err)
	related, err := left.RelatedFiles(false)
	require.NoError(t, err)
	require.Len(t, related.Files, 1)

	require.NoError(t, reconcileInsta360Photos(related))

	var files []entity.File
	require.NoError(t, entity.UnscopedDb().Where("file_name LIKE ?", name+"/%").Order("file_name").Find(&files).Error)
	require.Len(t, files, 2)

	for i, file := range files {
		assert.Equal(t, photos[i].ID, file.PhotoID, file.FileName)
	}

	var other entity.Photo
	require.NoError(t, entity.UnscopedDb().First(&other, "id = ?", photos[1].ID).Error)
	assert.Nil(t, other.DeletedAt)
}
