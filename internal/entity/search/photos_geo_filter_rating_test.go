package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

func TestPhotosGeoFilterRating(t *testing.T) {
	photo := entity.PhotoFixtures.Get("Photo01")

	if err := entity.UnscopedDb().Model(&entity.Photo{}).Where("id = ?", photo.ID).Updates(entity.Values{"photo_rating": 4, "rating_src": entity.SrcMeta}).Error; err != nil {
		t.Fatal(err)
	}

	defer entity.UnscopedDb().Model(&entity.Photo{}).Where("id = ?", photo.ID).Updates(entity.Values{"photo_rating": entity.PhotoRatingUnknown, "rating_src": ""})

	var f form.SearchPhotosGeo

	f.Rating = "4"

	photos, _, err := Geo(f)

	if err != nil {
		t.Fatal(err)
	}

	for _, r := range photos {
		assert.Equal(t, 4, r.PhotoRating)
	}

	assert.NotEmpty(t, photos)
}

func TestPhotosGeoFilterRatingUIDMismatch(t *testing.T) {
	photo := entity.PhotoFixtures.Get("Photo01")

	if err := entity.UnscopedDb().Model(&entity.Photo{}).Where("id = ?", photo.ID).Updates(entity.Values{"photo_rating": 4, "rating_src": entity.SrcMeta}).Error; err != nil {
		t.Fatal(err)
	}

	defer entity.UnscopedDb().Model(&entity.Photo{}).Where("id = ?", photo.ID).Updates(entity.Values{"photo_rating": entity.PhotoRatingUnknown, "rating_src": ""})

	var f form.SearchPhotosGeo

	f.UID = photo.PhotoUID
	f.Rating = "5"

	photos, _, err := Geo(f)

	if err != nil {
		t.Fatal(err)
	}

	assert.Empty(t, photos)
}
