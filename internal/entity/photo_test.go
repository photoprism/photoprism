package entity

import (
	"github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSavePhotoForm(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		f := form.Photo{
			TakenAt:          time.Date(2008, 1, 1, 2, 0, 0, 0, time.UTC),
			TakenAtLocal:     time.Date(2008, 1, 1, 2, 0, 0, 0, time.UTC),
			TakenSrc:         "manual",
			TimeZone:         "test",
			PhotoTitle:       "Pink beach",
			TitleSrc:         "manual",
			PhotoFavorite:    true,
			PhotoPrivate:     true,
			PhotoStory:       false,
			PhotoLat:         7.9999,
			PhotoLng:         8.8888,
			PhotoAltitude:    2,
			PhotoIso:         5,
			PhotoFocalLength: 10,
			PhotoFNumber:     3.3,
			PhotoExposure:    "exposure",
			CameraID:         uint(3),
			CameraSrc:        "exif",
			LensID:           uint(6),
			LocationID:       "1234",
			LocationSrc:      "geo",
			PlaceID:          "765",
			PhotoCountry:     "de",
		}

		m := PhotoFixtures["Photo08"]

		err := SavePhotoForm(m, f, "places")

		if err != nil {
			t.Fatal(err)
		}

		Db().First(&m)

		assert.Equal(t, "manual", m.TakenSrc)
		assert.Equal(t, "test", m.TimeZone)
		assert.Equal(t, "Pink beach", m.PhotoTitle)
		assert.Equal(t, "manual", m.TitleSrc)
		assert.Equal(t, true, m.PhotoFavorite)
		assert.Equal(t, true, m.PhotoPrivate)
		assert.Equal(t, false, m.PhotoStory)
		assert.Equal(t, float32(7.9999), m.PhotoLat)
		assert.NotNil(t, m.EditedAt)

	})
}

func TestPhoto_Save(t *testing.T) {
	t.Run("new photo", func(t *testing.T) {
		photo := Photo{
			TakenAt:          time.Date(2008, 1, 1, 2, 0, 0, 0, time.UTC),
			TakenAtLocal:     time.Date(2008, 1, 1, 2, 0, 0, 0, time.UTC),
			TakenSrc:         "exif",
			TimeZone:         "UTC",
			PhotoTitle:       "Black beach",
			TitleSrc:         "manual",
			PhotoFavorite:    false,
			PhotoPrivate:     false,
			PhotoStory:       true,
			PhotoLat:         9.9999,
			PhotoLng:         8.8888,
			PhotoAltitude:    2,
			PhotoIso:         5,
			PhotoFocalLength: 10,
			PhotoFNumber:     3.3,
			PhotoExposure:    "exposure",
			CameraID:         uint(3),
			CameraSrc:        "exif",
			LensID:           uint(6),
			LocationID:       "1234",
			LocationSrc:      "geo",
			PlaceID:          "765",
			PhotoCountry:     "de"}

		err := photo.Save()
		if err != nil {
			t.Fatal("error")
		}
	})
	t.Run("existing photo", func(t *testing.T) {
		err := PhotoFixture19800101_000002_D640C559.Save()
		if err != nil {
			t.Fatal("error")
		}
	})
}

func TestPhoto_ClassifyLabels(t *testing.T) {
	t.Run("new photo", func(t *testing.T) {
		m := PhotoFixturePhoto01
		Db().Set("gorm:auto_preload", true).Model(&m).Related(&m.Labels)
		labels := m.ClassifyLabels()
		assert.Empty(t, labels)
	})
	t.Run("existing photo", func(t *testing.T) {
		m := PhotoFixture19800101_000002_D640C559
		Db().Set("gorm:auto_preload", true).Model(&m).Related(&m.Labels)
		labels := m.ClassifyLabels()
		assert.LessOrEqual(t, 2, labels.Len())
	})
}
