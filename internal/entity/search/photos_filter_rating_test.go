package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

func TestPhotosFilterRating(t *testing.T) {
	t.Run("Direct", func(t *testing.T) {
		photo := entity.PhotoFixtures.Get("Photo02")

		if err := entity.UnscopedDb().Model(&entity.Photo{}).Where("id = ?", photo.ID).Updates(entity.Values{"photo_rating": 4, "rating_src": entity.SrcMeta}).Error; err != nil {
			t.Fatal(err)
		}

		defer entity.UnscopedDb().Model(&entity.Photo{}).Where("id = ?", photo.ID).Updates(entity.Values{"photo_rating": entity.PhotoRatingUnknown, "rating_src": ""})

		var f form.SearchPhotos

		f.Rating = "4"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.Equal(t, 4, r.PhotoRating)
		}

		assert.NotEmpty(t, photos)
	})

	t.Run("UIDMismatch", func(t *testing.T) {
		photo := entity.PhotoFixtures.Get("Photo02")

		if err := entity.UnscopedDb().Model(&entity.Photo{}).Where("id = ?", photo.ID).Updates(entity.Values{"photo_rating": 4, "rating_src": entity.SrcMeta}).Error; err != nil {
			t.Fatal(err)
		}

		defer entity.UnscopedDb().Model(&entity.Photo{}).Where("id = ?", photo.ID).Updates(entity.Values{"photo_rating": entity.PhotoRatingUnknown, "rating_src": ""})

		var f form.SearchPhotos

		f.UID = photo.PhotoUID
		f.Rating = "5"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Empty(t, photos)
	})

	t.Run("Invalid", func(t *testing.T) {
		var f form.SearchPhotos

		f.Rating = "%gold"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Greater(t, len(photos), 0)
	})
}
