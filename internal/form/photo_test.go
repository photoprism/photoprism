package form

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewPhoto(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		photo := Photo{
			TakenAt:          time.Date(2008, 1, 1, 2, 0, 0, 0, time.UTC),
			TakenAtLocal:     time.Date(2008, 1, 1, 2, 0, 0, 0, time.UTC),
			TakenSrc:         "meta",
			TimeZone:         "UTC",
			PhotoTitle:       "Black beach",
			TitleSrc:         "manual",
			PhotoFavorite:    false,
			PhotoPrivate:     false,
			PhotoType:        "image",
			PhotoStack:       int8(1),
			PhotoLat:         9.9999,
			PhotoLng:         8.8888,
			PhotoAltitude:    2,
			PhotoIso:         5,
			PhotoFocalLength: 10,
			PhotoFNumber:     3.3,
			PhotoExposure:    "exposure",
			CameraID:         uint(3),
			CameraSrc:        "meta",
			LensID:           uint(6),
			CellID:           "1234",
			PlaceSrc:         "geo",
			PlaceID:          "765",
			PhotoCountry:     "de"}

		r, err := NewPhoto(photo)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, time.Date(2008, 1, 1, 2, 0, 0, 0, time.UTC), r.TakenAt)
		assert.Equal(t, time.Date(2008, 1, 1, 2, 0, 0, 0, time.UTC), r.TakenAtLocal)
		assert.Equal(t, "meta", r.TakenSrc)
		assert.Equal(t, "UTC", r.TimeZone)
		assert.Equal(t, "Black beach", r.PhotoTitle)
		assert.Equal(t, "manual", r.TitleSrc)
		assert.Equal(t, false, r.PhotoFavorite)
		assert.Equal(t, false, r.PhotoPrivate)
		assert.Equal(t, "image", r.PhotoType)
		assert.Equal(t, int8(1), r.PhotoStack)
		assert.Equal(t, 9.9999, r.PhotoLat)
		assert.Equal(t, 8.8888, r.PhotoLng)
		assert.Equal(t, 2, r.PhotoAltitude)
		assert.Equal(t, 5, r.PhotoIso)
		assert.Equal(t, 10, r.PhotoFocalLength)
		assert.Equal(t, float32(3.3), r.PhotoFNumber)
		assert.Equal(t, "exposure", r.PhotoExposure)
		assert.Equal(t, uint(3), r.CameraID)
		assert.Equal(t, "meta", r.CameraSrc)
		assert.Equal(t, uint(6), r.LensID)
		assert.Equal(t, "1234", r.CellID)
		assert.Equal(t, "geo", r.PlaceSrc)
		assert.Equal(t, "765", r.PlaceID)
		assert.Equal(t, "de", r.PhotoCountry)
	})
}
