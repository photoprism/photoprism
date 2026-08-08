package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/media/projection"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// createFisheyeTestPhoto adds a photo shaped like an indexed Insta360 capture, with the dewarp
// derivative as primary file and the fisheye original beside it.
// The rows are removed again so shared fixture counts other tests assert on stay stable.
func createFisheyeTestPhoto(t *testing.T, originalProjection projection.Type) entity.Photo {
	t.Helper()

	photo := entity.Photo{
		PhotoUID:     rnd.GenerateUID(entity.PhotoUID),
		PhotoName:    "20220625_140410_FISHEYE",
		PhotoPath:    "fisheye",
		PhotoType:    entity.MediaImage,
		PhotoQuality: 3,
		TakenAt:      entity.Now(),
		TakenAtLocal: entity.Now(),
		TakenSrc:     entity.SrcMeta,
	}

	require.NoError(t, entity.Db().Create(&photo).Error)

	files := []entity.File{
		{
			PhotoID:        photo.ID,
			PhotoUID:       photo.PhotoUID,
			FileUID:        rnd.GenerateUID(entity.FileUID),
			FileName:       "fisheye/20220625_140410_FISHEYE.insp.jpg",
			FileRoot:       entity.RootSidecar,
			FileHash:       rnd.UUIDv7(),
			FileType:       "jpg",
			MediaType:      media.Image.String(),
			FilePrimary:    true,
			FileProjection: projection.Equirectangular.String(),
		},
		{
			PhotoID:        photo.ID,
			PhotoUID:       photo.PhotoUID,
			FileUID:        rnd.GenerateUID(entity.FileUID),
			FileName:       "fisheye/20220625_140410_FISHEYE.insp",
			FileRoot:       entity.RootOriginals,
			FileHash:       rnd.UUIDv7(),
			FileType:       "insp",
			MediaType:      media.Image.String(),
			FileProjection: originalProjection.String(),
		},
	}

	for i := range files {
		require.NoError(t, entity.Db().Create(&files[i]).Error)
	}

	// media_id is a maintained column the fisheye subquery scopes on, exactly as the base query does.
	entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID}.RegenerateIndex()

	t.Cleanup(func() {
		entity.UnscopedDb().Where("photo_id = ?", photo.ID).Delete(&entity.File{})
		entity.UnscopedDb().Where("id = ?", photo.ID).Delete(&entity.Photo{})
	})

	return photo
}

// containsPhotoUID reports whether the results include the specified photo.
func containsPhotoUID(results PhotoResults, photoUID string) bool {
	for _, result := range results {
		if result.PhotoUID == photoUID {
			return true
		}
	}

	return false
}

func TestPhotosQueryFisheye(t *testing.T) {
	t.Run("MatchesDualFisheyeOriginal", func(t *testing.T) {
		photo := createFisheyeTestPhoto(t, projection.DualFisheye)

		var frm form.SearchPhotos
		frm.Query = "fisheye:true"
		frm.Merged = true
		require.NoError(t, frm.ParseQueryString())

		results, _, err := Photos(frm)
		require.NoError(t, err)
		assert.True(t, containsPhotoUID(results, photo.PhotoUID), "the dual-fisheye original must match")
	})
	t.Run("MatchesFisheyeOriginal", func(t *testing.T) {
		photo := createFisheyeTestPhoto(t, projection.Fisheye)

		var frm form.SearchPhotos
		frm.Query = "fisheye:true"
		frm.Merged = true
		require.NoError(t, frm.ParseQueryString())

		results, _, err := Photos(frm)
		require.NoError(t, err)
		assert.True(t, containsPhotoUID(results, photo.PhotoUID), "the single-fisheye original must match")
	})
	t.Run("SkipsEquirectangularOnly", func(t *testing.T) {
		// A photo whose original was already equirectangular is not a fisheye capture.
		photo := createFisheyeTestPhoto(t, projection.Equirectangular)

		var frm form.SearchPhotos
		frm.Query = "fisheye:true"
		frm.Merged = true
		require.NoError(t, frm.ParseQueryString())

		results, _, err := Photos(frm)
		require.NoError(t, err)
		assert.False(t, containsPhotoUID(results, photo.PhotoUID), "an equirectangular original must not match")
	})
	t.Run("NarrowsResults", func(t *testing.T) {
		createFisheyeTestPhoto(t, projection.DualFisheye)

		var frm form.SearchPhotos
		frm.Query = "fisheye:true"
		frm.Merged = true
		require.NoError(t, frm.ParseQueryString())

		matched, _, err := Photos(frm)
		require.NoError(t, err)

		var all form.SearchPhotos
		all.Merged = true

		total, _, err := Photos(all)
		require.NoError(t, err)

		assert.NotEmpty(t, matched)
		assert.Less(t, len(matched), len(total), "the filter must exclude the non-fisheye fixtures")
	})
	t.Run("NavigationTerm", func(t *testing.T) {
		// The "Fisheye" navigation item searches for the bare term rather than "fisheye:true".
		photo := createFisheyeTestPhoto(t, projection.DualFisheye)

		var frm form.SearchPhotos
		frm.Query = "fisheye"
		frm.Merged = true
		require.NoError(t, frm.ParseQueryString())

		results, _, err := Photos(frm)
		require.NoError(t, err)
		assert.True(t, containsPhotoUID(results, photo.PhotoUID), "the bare term must apply the same filter")
	})
}
