package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/media"
)

// newInsta360ReconcileFixture indexes one capture as three unrelated photos, as a previous index
// run would have left them, and returns the related files, those photos, and their file names
// relative to the originals root, all in canonical lens order.
//
// The capture directory is named after the test case because a MariaDB run gives the whole package
// a single database, so a shared directory name would collide across subtests. Do not isolate the
// database with PHOTOPRISM_TEST_DSN here: that leaves the driver pointing at MySQL and terminates
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

// TestReconcileInsta360Photos verifies force-reindex state preservation and file reassignment.
func TestReconcileInsta360Photos(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		related, photos, relNames := newInsta360ReconcileFixture(t, "insta360reconcile")

		archivedAt := entity.Now()
		photos[0].DeletedAt = &archivedAt
		photos[1].DeletedAt = &archivedAt
		photos[1].PhotoFavorite = true
		photos[2].PhotoCaption = "preserved manual caption"
		photos[2].CaptionSrc = entity.SrcManual
		for i := range photos {
			require.NoError(t, photos[i].Save())
		}

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

		require.NoError(t, reconcileInsta360Photos(related))

		var canonical entity.Photo
		require.NoError(t, entity.UnscopedDb().First(&canonical, "id = ?", photos[0].ID).Error)
		assert.True(t, canonical.PhotoFavorite)
		assert.Nil(t, canonical.DeletedAt)
		assert.Equal(t, "preserved manual caption", canonical.PhotoCaption)
		assert.Equal(t, entity.SrcManual, canonical.CaptionSrc)

		var files []entity.File
		require.NoError(t, entity.UnscopedDb().Where("file_name IN (?)", relNames).Find(&files).Error)
		require.Len(t, files, 3)
		for _, file := range files {
			assert.Equal(t, canonical.ID, file.PhotoID)
			assert.Equal(t, canonical.PhotoUID, file.PhotoUID)
		}

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
