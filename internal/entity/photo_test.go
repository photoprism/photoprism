package entity

import (
	"github.com/photoprism/photoprism/internal/classify"
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
			PhotoVideo:       false,
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
		assert.Equal(t, false, m.PhotoVideo)
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
			PhotoVideo:       true,
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
			t.Fatal(err)
		}
	})
	t.Run("existing photo", func(t *testing.T) {
		m := PhotoFixtures.Get("19800101_000002_D640C559")
		err := m.Save()
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestPhoto_ClassifyLabels(t *testing.T) {
	t.Run("new photo", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		Db().Set("gorm:auto_preload", true).Model(&m).Related(&m.Labels)
		labels := m.ClassifyLabels()
		assert.Empty(t, labels)
	})
	t.Run("existing photo", func(t *testing.T) {
		m := PhotoFixtures.Get("19800101_000002_D640C559")
		Db().Set("gorm:auto_preload", true).Model(&m).Related(&m.Labels)
		labels := m.ClassifyLabels()
		assert.LessOrEqual(t, 2, labels.Len())
	})
}

func TestPhoto_PreloadFiles(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		assert.Empty(t, m.Files)
		m.PreloadFiles()
		assert.NotEmpty(t, m.Files)
	})
}

func TestPhoto_PreloadKeywords(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		assert.Empty(t, m.Keywords)
		m.PreloadKeywords()
		assert.NotEmpty(t, m.Keywords)
	})
}

func TestPhoto_PreloadAlbums(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		assert.Empty(t, m.Albums)
		m.PreloadAlbums()
		assert.NotEmpty(t, m.Albums)
	})
}

func TestPhoto_PreloadMany(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		assert.Empty(t, m.Albums)
		assert.Empty(t, m.Files)
		assert.Empty(t, m.Keywords)

		m.PreloadMany()

		assert.NotEmpty(t, m.Files)
		assert.NotEmpty(t, m.Albums)
		assert.NotEmpty(t, m.Keywords)
	})
}

func TestPhoto_NoLocation(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		assert.True(t, m.NoLocation())
	})
	t.Run("false", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo08")
		assert.False(t, m.NoLocation())
	})
}

func TestPhoto_HasLocation(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		assert.False(t, m.HasLocation())
	})
	t.Run("true", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo08")
		assert.True(t, m.HasLocation())
	})
}

func TestPhoto_HasLatLng(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		assert.True(t, m.HasLatLng())
	})
	t.Run("false", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo08")
		assert.False(t, m.HasLatLng())
	})
}

func TestPhoto_NoLatLng(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		assert.False(t, m.NoLatLng())
	})
	t.Run("true", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo08")
		assert.True(t, m.NoLatLng())
	})
}

func TestPhoto_NoPlace(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		assert.True(t, m.NoPlace())
	})
	t.Run("false", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo08")
		assert.False(t, m.NoPlace())
	})
}

func TestPhoto_HasPlace(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		assert.False(t, m.HasPlace())
	})
	t.Run("true", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo08")
		assert.True(t, m.HasPlace())
	})
}

func TestPhoto_HasTitle(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo03")
		assert.False(t, m.HasTitle())
	})
	t.Run("true", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo04")
		assert.True(t, m.HasTitle())
	})
}

func TestPhoto_NoTitle(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo03")
		assert.True(t, m.NoTitle())
	})
	t.Run("false", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo04")
		assert.False(t, m.NoTitle())
	})
}

func TestPhoto_NoCameraSerial(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo04")
		assert.True(t, m.NoCameraSerial())
	})
	t.Run("false", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo05")
		assert.False(t, m.NoCameraSerial())
	})
}

func TestPhoto_DescriptionLoaded(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		m := PhotoFixtures.Get("19800101_000002_D640C559")
		assert.True(t, m.DescriptionLoaded())
	})
	t.Run("false", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo05")
		assert.False(t, m.DescriptionLoaded())
	})
}

func TestPhoto_UpdateTitle(t *testing.T) {
	t.Run("wont update title was modified", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo08")
		classifyLabels := &classify.Labels{}
		assert.Equal(t, "Black beach", m.PhotoTitle)
		err := m.UpdateTitle(*classifyLabels)
		if err == nil {
			t.Fatal()
		}
		assert.Equal(t, "Black beach", m.PhotoTitle)
		assert.Equal(t, "photo: won't update title, was modified", err.Error())
	})
	t.Run("photo with location without city and label", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo10")
		classifyLabels := &classify.Labels{{Name: "tree", Uncertainty: 30, Source: "manual", Priority: 5, Categories: []string{"plant"}}}
		assert.Equal(t, "Title", m.PhotoTitle)
		err := m.UpdateTitle(*classifyLabels)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "Tree / Germany / 2016", m.PhotoTitle)
	})
	t.Run("photo with location and short city and label", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo09")
		classifyLabels := &classify.Labels{{Name: "tree", Uncertainty: 30, Source: "manual", Priority: 5, Categories: []string{"plant"}}}
		assert.Equal(t, "Title", m.PhotoTitle)
		err := m.UpdateTitle(*classifyLabels)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "Tree / Teotihuacán / 2016", m.PhotoTitle)
	})
	t.Run("photo with location and locname >45", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo13")
		classifyLabels := &classify.Labels{}
		assert.Equal(t, "Title", m.PhotoTitle)
		err := m.UpdateTitle(*classifyLabels)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "LonglonglonglonglonglonglonglonglonglonglonglonglongName", m.PhotoTitle)
	})
	t.Run("photo with location and locname >20", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo14")
		classifyLabels := &classify.Labels{}
		assert.Equal(t, "Title", m.PhotoTitle)
		err := m.UpdateTitle(*classifyLabels)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "longlonglonglonglonglongName / 2016", m.PhotoTitle)
	})

	t.Run("photo with location and short city", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo09")
		classifyLabels := &classify.Labels{}
		assert.Equal(t, "Title", m.PhotoTitle)
		err := m.UpdateTitle(*classifyLabels)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "Adosada Platform / Teotihuacán / 2016", m.PhotoTitle)
	})
	t.Run("photo with location without city", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo10")
		classifyLabels := &classify.Labels{}
		assert.Equal(t, "Title", m.PhotoTitle)
		err := m.UpdateTitle(*classifyLabels)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "Holiday Park / Germany / 2016", m.PhotoTitle)
	})

	t.Run("photo with location without  loc name and long city", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo11")
		classifyLabels := &classify.Labels{}
		assert.Equal(t, "Title", m.PhotoTitle)
		err := m.UpdateTitle(*classifyLabels)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "longlonglonglonglongcity / 2016", m.PhotoTitle)
	})
	t.Run("photo with location without loc name and short city", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo12")
		classifyLabels := &classify.Labels{}
		assert.Equal(t, "Title", m.PhotoTitle)
		err := m.UpdateTitle(*classifyLabels)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "shortcity / Germany / 2016", m.PhotoTitle)
	})
	t.Run("no location", func(t *testing.T) {
		m := PhotoFixtures.Get("19800101_000002_D640C559")
		classifyLabels := &classify.Labels{{Name: "classify", Uncertainty: 30, Source: "manual", Priority: 5, Categories: []string{"flower", "plant"}}}
		assert.Equal(t, "", m.PhotoTitle)
		err := m.UpdateTitle(*classifyLabels)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "Classify / 2008", m.PhotoTitle)
	})
	t.Run("no location no labels", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo02")
		classifyLabels := &classify.Labels{}
		assert.Equal(t, "", m.PhotoTitle)
		err := m.UpdateTitle(*classifyLabels)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "Unknown / 2008", m.PhotoTitle)
	})
	t.Run("no location no labels no takenAt", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo03")
		classifyLabels := &classify.Labels{}
		assert.Equal(t, "", m.PhotoTitle)
		err := m.UpdateTitle(*classifyLabels)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "Unknown", m.PhotoTitle)
	})
}
