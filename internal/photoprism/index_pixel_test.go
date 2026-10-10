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
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
)

// newGooglePixelReconcileFixture indexes one Google Pixel Camera capture as three unrelated photos,
// as an older index run would have left them, and returns them in canonical order.
func newGooglePixelReconcileFixture(t *testing.T, name string) (RelatedFiles, []entity.Photo, []string) {
	t.Helper()
	cfg := config.NewMinimalTestConfigWithDb(name, filepath.Join(t.TempDir(), "storage"))
	oldCfg := Config()
	SetConfig(cfg)
	t.Cleanup(func() {
		SetConfig(oldCfg)
		oldCfg.RegisterDb()
	})

	dir := filepath.Join(cfg.OriginalsPath(), name)
	require.NoError(t, fs.MkdirAll(dir))

	coverFile := filepath.Join(dir, "PXL_20240926_143000123.RAW-01.jpg")
	rawFile1 := filepath.Join(dir, "PXL_20240926_143000123.RAW-02.ORIGINAL.dng")
	rawFile2 := filepath.Join(dir, "PXL_20240926_143000123.RAW-03.ORIGINAL.dng")

	require.NoError(t, fs.Copy("testdata/flash.jpg", coverFile, false))
	require.NoError(t, fs.Copy(filepath.Join(cfg.SamplesPath(), "canon_eos_6d.dng"), rawFile1, false))
	require.NoError(t, fs.Copy(filepath.Join(cfg.SamplesPath(), "canon_eos_6d.dng"), rawFile2, false))

	fileNames := []string{coverFile, rawFile1, rawFile2}

	coverMedia, err := NewMediaFile(fileNames[0])
	require.NoError(t, err)
	related, err := coverMedia.RelatedFiles(false)
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
			MediaType: media.Image.String(),
		}
		require.NoError(t, file.Create())
		relNames = append(relNames, file.FileName)
	}

	return related, photos, relNames
}

// linkGooglePixelAssociations adds one album, label, and keyword relation to the specified photo.
func linkGooglePixelAssociations(t *testing.T, photo entity.Photo, albumUID string, labelID, keywordID uint) {
	t.Helper()
	require.NoError(t, entity.NewPhotoAlbum(photo.PhotoUID, albumUID).Create())
	require.NoError(t, entity.NewPhotoLabel(photo.ID, labelID, 100, entity.SrcManual).Create())
	require.NoError(t, entity.NewPhotoKeyword(photo.ID, keywordID).Create())
}

// googlePixelAlbumUIDs returns the album UIDs currently linked to the specified photo UID.
func googlePixelAlbumUIDs(t *testing.T, photoUID string) []string {
	t.Helper()
	result := make([]string, 0, 2)
	require.NoError(t, entity.UnscopedDb().Model(&entity.PhotoAlbum{}).
		Where("photo_uid = ?", photoUID).Pluck("album_uid", &result).Error)
	return result
}

// googlePixelLabelIDs returns the label IDs currently linked to the specified photo.
func googlePixelLabelIDs(t *testing.T, photoID uint) []uint {
	t.Helper()
	result := make([]uint, 0, 2)
	require.NoError(t, entity.UnscopedDb().Model(&entity.PhotoLabel{}).
		Where("photo_id = ?", photoID).Pluck("label_id", &result).Error)
	return result
}

// googlePixelKeywordIDs returns the keyword IDs currently linked to the specified photo.
func googlePixelKeywordIDs(t *testing.T, photoID uint) []uint {
	t.Helper()
	result := make([]uint, 0, 2)
	require.NoError(t, entity.UnscopedDb().Model(&entity.PhotoKeyword{}).
		Where("photo_id = ?", photoID).Pluck("keyword_id", &result).Error)
	return result
}

// setGooglePixelPhotoState sets the archive state and quality of a photo without side effects.
func setGooglePixelPhotoState(t *testing.T, photo entity.Photo, deletedAt *time.Time, quality int) {
	t.Helper()
	require.NoError(t, entity.UnscopedDb().Model(&entity.Photo{}).Where("id = ?", photo.ID).
		UpdateColumns(entity.Values{"deleted_at": deletedAt, "photo_quality": quality}).Error)
}

// findGooglePixelPhoto returns the photo with the specified ID, including deleted rows.
func findGooglePixelPhoto(t *testing.T, id uint) (result entity.Photo) {
	t.Helper()
	require.NoError(t, entity.UnscopedDb().First(&result, "id = ?", id).Error)
	return result
}

// TestGooglePixelPhotoRemoved verifies that only deleted rows at quality -1 count as removed.
func TestGooglePixelPhotoRemoved(t *testing.T) {
	deletedAt := entity.Now()
	assert.True(t, googlePixelPhotoRemoved(&entity.Photo{DeletedAt: &deletedAt, PhotoQuality: -1}))
	assert.False(t, googlePixelPhotoRemoved(&entity.Photo{DeletedAt: &deletedAt, PhotoQuality: 0}))
	assert.False(t, googlePixelPhotoRemoved(&entity.Photo{PhotoQuality: -1}))
	assert.False(t, googlePixelPhotoRemoved(&entity.Photo{}))
	assert.False(t, googlePixelPhotoRemoved(nil))
}

// TestGooglePixelCaptureState verifies which photo decides the archive state of a merged capture.
func TestGooglePixelCaptureState(t *testing.T) {
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
			result := googlePixelCaptureState(tc.photos)
			require.NotNil(t, result)
			assert.Equal(t, tc.expect, result.ID)
		})
	}
	t.Run("Empty", func(t *testing.T) {
		assert.Nil(t, googlePixelCaptureState(nil))
		assert.Nil(t, googlePixelCaptureState(entity.Photos{nil}))
		assert.Equal(t, uint(1), googlePixelCaptureState(entity.Photos{photo(1, &removedAt, -1), nil}).ID)
	})
}

// TestReconcileGooglePixelPhotos verifies force-reindex state preservation and file reassignment.
func TestReconcileGooglePixelPhotos(t *testing.T) {
	// Success: verifies that canonical photo absorbs duplicates and migrates all metadata and associations.
	t.Run("Success", func(t *testing.T) {
		related, photos, relNames := newGooglePixelReconcileFixture(t, "pixelreconcile")

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
		linkGooglePixelAssociations(t, photos[0], sharedAlbum.AlbumUID, sharedLabel.ID, sharedKeyword.ID)
		linkGooglePixelAssociations(t, photos[1], sharedAlbum.AlbumUID, sharedLabel.ID, sharedKeyword.ID)
		linkGooglePixelAssociations(t, photos[2], uniqueAlbum.AlbumUID, uniqueLabel.ID, uniqueKeyword.ID)

		logger := logrus.StandardLogger()
		oldHooks := make(logrus.LevelHooks, len(logger.Hooks))
		for level, hooks := range logger.Hooks {
			oldHooks[level] = append([]logrus.Hook(nil), hooks...)
		}
		hook := test.NewGlobal()
		t.Cleanup(func() { logger.ReplaceHooks(oldHooks) })

		require.NoError(t, reconcileGooglePixelPhotos(related))

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

		assert.ElementsMatch(t, []string{sharedAlbum.AlbumUID, uniqueAlbum.AlbumUID}, googlePixelAlbumUIDs(t, canonical.PhotoUID))
		assert.ElementsMatch(t, []uint{sharedLabel.ID, uniqueLabel.ID}, googlePixelLabelIDs(t, canonical.ID))
		assert.ElementsMatch(t, []uint{sharedKeyword.ID, uniqueKeyword.ID}, googlePixelKeywordIDs(t, canonical.ID))
		for _, duplicate := range photos[1:] {
			assert.Empty(t, googlePixelAlbumUIDs(t, duplicate.PhotoUID))
			assert.Empty(t, googlePixelLabelIDs(t, duplicate.ID))
			assert.Empty(t, googlePixelKeywordIDs(t, duplicate.ID))
		}
	})

	// AllArchived: verifies that canonical photo remains archived when all members are archived.
	t.Run("AllArchived", func(t *testing.T) {
		related, photos, _ := newGooglePixelReconcileFixture(t, "pixelarchived")

		archivedAt := entity.Now()
		for i := range photos {
			photos[i].DeletedAt = &archivedAt
			require.NoError(t, photos[i].Save())
		}

		require.NoError(t, reconcileGooglePixelPhotos(related))

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

	// CanonicalArchived: verifies that canonical photo retains its archived state.
	t.Run("CanonicalArchived", func(t *testing.T) {
		related, photos, _ := newGooglePixelReconcileFixture(t, "pixelcanonicalarchived")
		archivedAt := entity.Now()
		setGooglePixelPhotoState(t, photos[0], &archivedAt, 3)

		require.NoError(t, reconcileGooglePixelPhotos(related))

		canonical := findGooglePixelPhoto(t, photos[0].ID)
		assert.NotNil(t, canonical.DeletedAt)
		assert.Equal(t, 3, canonical.PhotoQuality)
	})

	// MemberArchived: verifies that canonical photo remains active when a member is archived.
	t.Run("MemberArchived", func(t *testing.T) {
		related, photos, _ := newGooglePixelReconcileFixture(t, "pixelmemberarchived")
		archivedAt := entity.Now()
		setGooglePixelPhotoState(t, photos[0], nil, 3)
		setGooglePixelPhotoState(t, photos[1], &archivedAt, 3)

		require.NoError(t, reconcileGooglePixelPhotos(related))

		canonical := findGooglePixelPhoto(t, photos[0].ID)
		assert.Nil(t, canonical.DeletedAt)
		assert.Equal(t, 3, canonical.PhotoQuality)
	})

	// MemberRemoved: verifies that canonical photo remains active when a member is auto-removed.
	t.Run("MemberRemoved", func(t *testing.T) {
		related, photos, _ := newGooglePixelReconcileFixture(t, "pixelmemberremoved")
		removedAt := entity.Now()
		setGooglePixelPhotoState(t, photos[0], nil, 3)
		setGooglePixelPhotoState(t, photos[2], &removedAt, -1)

		require.NoError(t, reconcileGooglePixelPhotos(related))

		canonical := findGooglePixelPhoto(t, photos[0].ID)
		assert.Nil(t, canonical.DeletedAt)
		assert.Equal(t, 3, canonical.PhotoQuality)
	})

	// CanonicalRemoved: verifies that canonical photo is restored to active if an original member was still active.
	t.Run("CanonicalRemoved", func(t *testing.T) {
		related, photos, _ := newGooglePixelReconcileFixture(t, "pixelcanonicalremoved")
		removedAt := entity.Now()
		setGooglePixelPhotoState(t, photos[0], &removedAt, -1)
		setGooglePixelPhotoState(t, photos[1], nil, 3)

		require.NoError(t, reconcileGooglePixelPhotos(related))

		canonical := findGooglePixelPhoto(t, photos[0].ID)
		assert.Nil(t, canonical.DeletedAt)
		assert.Equal(t, 3, canonical.PhotoQuality)
	})

	// CanonicalRemovedMemberArchived: verifies that canonical photo keeps the archive state of an archived member.
	t.Run("CanonicalRemovedMemberArchived", func(t *testing.T) {
		related, photos, _ := newGooglePixelReconcileFixture(t, "pixelremovedarchived")
		removedAt := entity.Now()
		setGooglePixelPhotoState(t, photos[0], &removedAt, -1)
		setGooglePixelPhotoState(t, photos[1], &removedAt, -1)
		setGooglePixelPhotoState(t, photos[2], &removedAt, 4)

		require.NoError(t, reconcileGooglePixelPhotos(related))

		canonical := findGooglePixelPhoto(t, photos[0].ID)
		assert.NotNil(t, canonical.DeletedAt)
		assert.Equal(t, 4, canonical.PhotoQuality)
	})

	// MemberArchivedBeforeRemoval: verifies that canonical photo remains removed if member was archived prior to removal.
	t.Run("MemberArchivedBeforeRemoval", func(t *testing.T) {
		related, photos, _ := newGooglePixelReconcileFixture(t, "pixelarchivedbefore")
		archivedAt := entity.Now().Add(-time.Minute)
		removedAt := entity.Now()
		setGooglePixelPhotoState(t, photos[0], &removedAt, -1)
		setGooglePixelPhotoState(t, photos[1], &archivedAt, 4)
		setGooglePixelPhotoState(t, photos[2], &removedAt, -1)

		require.NoError(t, reconcileGooglePixelPhotos(related))

		// The indexer restores the photo when it indexes the capture again.
		canonical := findGooglePixelPhoto(t, photos[0].ID)
		assert.NotNil(t, canonical.DeletedAt)
		assert.Equal(t, -1, canonical.PhotoQuality)
	})

	// AllRemoved: verifies that canonical photo remains removed when all members are removed.
	t.Run("AllRemoved", func(t *testing.T) {
		related, photos, _ := newGooglePixelReconcileFixture(t, "pixelallremoved")
		removedAt := entity.Now()
		for _, photo := range photos {
			setGooglePixelPhotoState(t, photo, &removedAt, -1)
		}

		require.NoError(t, reconcileGooglePixelPhotos(related))

		// The indexer restores the photo when it indexes the capture again.
		canonical := findGooglePixelPhoto(t, photos[0].ID)
		assert.NotNil(t, canonical.DeletedAt)
		assert.Equal(t, -1, canonical.PhotoQuality)
	})

	// AllRemovedRestored: verifies that indexing the capture again merges its rows and restores the photo.
	t.Run("AllRemovedRestored", func(t *testing.T) {
		_, photos, relNames := newGooglePixelReconcileFixture(t, "pixelallremovedrestored")
		removedAt := entity.Now()
		for _, photo := range photos {
			setGooglePixelPhotoState(t, photo, &removedAt, -1)
		}

		// Indexing the capture again merges its rows and restores the photo.
		indexGooglePixelFolder(Config(), "pixelallremovedrestored", true)

		owners := googlePixelStackOwners(t, "pixelallremovedrestored")
		require.Len(t, owners, len(relNames))
		for _, relName := range relNames {
			owner := owners[filepath.Base(relName)]
			assert.Equal(t, photos[0].ID, owner.ID, relName)
			assert.Nil(t, owner.DeletedAt, relName)
			assert.GreaterOrEqual(t, owner.PhotoQuality, 0, relName)
		}
	})

	// Mixed: verifies proper precedence when members have mixed active, archived, and removed states.
	t.Run("Mixed", func(t *testing.T) {
		related, photos, _ := newGooglePixelReconcileFixture(t, "pixelmixed")
		archivedAt := entity.Now()
		setGooglePixelPhotoState(t, photos[0], &archivedAt, 3)
		setGooglePixelPhotoState(t, photos[1], nil, 3)
		setGooglePixelPhotoState(t, photos[2], &archivedAt, -1)

		require.NoError(t, reconcileGooglePixelPhotos(related))

		canonical := findGooglePixelPhoto(t, photos[0].ID)
		assert.NotNil(t, canonical.DeletedAt)
		assert.Equal(t, 3, canonical.PhotoQuality)
		for _, duplicate := range photos[1:] {
			merged := findGooglePixelPhoto(t, duplicate.ID)
			assert.NotNil(t, merged.DeletedAt)
			assert.Equal(t, -1, merged.PhotoQuality)
		}
	})

	// IncompletePair: verifies that an incomplete capture does not merge anything.
	t.Run("IncompletePair", func(t *testing.T) {
		related, photos, relNames := newGooglePixelReconcileFixture(t, "pixelincomplete")

		// Removing the raw originals leaves an incomplete capture, which must not merge anything.
		require.NoError(t, entity.UnscopedDb().Where("file_name IN (?)", []string{relNames[1], relNames[2]}).Delete(&entity.File{}).Error)
		require.NoError(t, os.Remove(filepath.Join(Config().OriginalsPath(), relNames[1])))
		require.NoError(t, os.Remove(filepath.Join(Config().OriginalsPath(), relNames[2])))

		reduced, err := related.Main.RelatedFiles(false)
		require.NoError(t, err)
		require.NoError(t, reconcileGooglePixelPhotos(reduced))

		for _, photo := range photos {
			var unchanged entity.Photo
			require.NoError(t, entity.UnscopedDb().First(&unchanged, "id = ?", photo.ID).Error)
			assert.Nil(t, unchanged.DeletedAt)
			assert.NotEqual(t, -1, unchanged.PhotoQuality)
		}
	})
}
