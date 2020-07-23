package entity

import (
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
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
			PhotoType:        "image",
			PhotoLat:         7.9999,
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
			PlaceSrc:         "manual",
			PlaceID:          "765",
			PhotoCountry:     "de",
			Details: form.Details{
				PhotoID:   uint(1000008),
				Keywords:  "test cat dog",
				Subject:   "animals",
				Artist:    "Bender",
				Notes:     "notes",
				Copyright: "copy",
				License:   "",
			},
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
		assert.Equal(t, "image", m.PhotoType)
		assert.Equal(t, float32(7.9999), m.PhotoLat)
		assert.NotNil(t, m.EditedAt)

		t.Log(m.GetDetails().Keywords)
	})
}

func TestPhoto_SaveLabels(t *testing.T) {
	t.Run("new photo", func(t *testing.T) {
		photo := Photo{
			ID:               11111,
			TakenAt:          time.Date(2008, 1, 1, 2, 0, 0, 0, time.UTC),
			TakenAtLocal:     time.Date(2008, 1, 1, 2, 0, 0, 0, time.UTC),
			TakenSrc:         "meta",
			TimeZone:         "UTC",
			PhotoTitle:       "Black beach",
			TitleSrc:         "manual",
			PhotoFavorite:    false,
			PhotoPrivate:     false,
			PhotoType:        "video",
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
			PhotoCountry:     "de",
			Keywords:         []Keyword{},
			Details: &Details{
				PhotoID:   11111,
				Keywords:  "test cat dog",
				Subject:   "animals",
				Artist:    "Bender",
				Notes:     "notes",
				Copyright: "copy",
				License:   "",
			},
		}

		err := photo.SaveLabels()

		assert.EqualError(t, err, "photo: can't save to database, id is empty")
	})

	t.Run("existing photo", func(t *testing.T) {
		m := PhotoFixtures.Get("19800101_000002_D640C559")
		err := m.SaveLabels()
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestPhoto_ClassifyLabels(t *testing.T) {
	t.Run("new photo", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo19")
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
	t.Run("empty label", func(t *testing.T) {
		p := Photo{}
		labels := p.ClassifyLabels()
		assert.Empty(t, labels)
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
		assert.True(t, m.UnknownLocation())
	})
	t.Run("false", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo08")
		// t.Logf("MODEL: %+v", m)
		assert.True(t, m.HasLocation())
		assert.False(t, m.UnknownLocation())
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
		assert.True(t, m.UnknownPlace())
	})
	t.Run("false", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo08")
		assert.False(t, m.UnknownPlace())
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

func TestPhoto_GetDetails(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		m := PhotoFixtures.Get("19800101_000002_D640C559")
		result := m.GetDetails()

		if result == nil {
			t.Fatal("result should never be nil")
		}

		if result.PhotoID != 1000000 {
			t.Fatal("PhotoID should not be 1000000")
		}
	})
	t.Run("false", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo12")
		result := m.GetDetails()

		if result == nil {
			t.Fatal("result should never be nil")
		}

		if result.PhotoID != 1000012 {
			t.Fatal("PhotoID should not be 1000012")
		}
	})
	t.Run("no ID", func(t *testing.T) {
		m := Photo{}
		result := m.GetDetails()
		assert.Equal(t, uint(0x0), result.PhotoID)
	})
	t.Run("new photo with ID", func(t *testing.T) {
		m := Photo{ID: 79550, PhotoUID: "pthkffkgk"}
		result := m.GetDetails()
		assert.Equal(t, uint(0x136be), result.PhotoID)
	})
}

func TestPhoto_FileTitle(t *testing.T) {
	t.Run("changing-of-the-guard--buckingham-palace_7925318070_o.jpg", func(t *testing.T) {
		photo := Photo{PhotoName: "20200102_194030_9EFA9E5E", PhotoPath: "2000/05", OriginalName: "flickr import/changing-of-the-guard--buckingham-palace_7925318070_o.jpg"}
		result := photo.FileTitle()
		assert.Equal(t, "Changing of the Guard / Buckingham Palace", result)
	})
	t.Run("empty title", func(t *testing.T) {
		photo := Photo{PhotoName: "", PhotoPath: "", OriginalName: ""}
		result := photo.FileTitle()
		assert.Equal(t, "", result)
	})
	t.Run("return title", func(t *testing.T) {
		photo := Photo{PhotoName: "sun, beach, fun", PhotoPath: "", OriginalName: "", PhotoTitle: ""}
		result := photo.FileTitle()
		assert.Equal(t, "Sun, Beach, Fun", result)
	})
	t.Run("return title", func(t *testing.T) {
		photo := Photo{PhotoName: "", PhotoPath: "vacation", OriginalName: "20200102_194030_9EFA9E5E", PhotoTitle: ""}
		result := photo.FileTitle()
		assert.Equal(t, "Vacation", result)
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
		m := PhotoFixtures.Get("Photo19")
		classifyLabels := &classify.Labels{}
		assert.Equal(t, "", m.PhotoTitle)
		err := m.UpdateTitle(*classifyLabels)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "Unknown", m.PhotoTitle)
	})
}

func TestPhoto_AddLabels(t *testing.T) {
	t.Run("add label", func(t *testing.T) {
		m := PhotoFixtures.Get("19800101_000002_D640C559")
		classifyLabels := classify.Labels{{Name: "cactus", Uncertainty: 30, Source: "manual", Priority: 5, Categories: []string{"plant"}}}
		len1 := len(m.Labels)
		m.AddLabels(classifyLabels)
		assert.Greater(t, len(m.Labels), len1)
	})
	t.Run("update label", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		classifyLabels := classify.Labels{{Name: "landscape", Uncertainty: 10, Source: "manual", Priority: 5, Categories: []string{"plant"}}}
		assert.Equal(t, 20, m.Labels[0].Uncertainty)
		assert.Equal(t, "image", m.Labels[0].LabelSrc)
		len1 := len(m.Labels)
		m.AddLabels(classifyLabels)
		assert.Equal(t, len(m.Labels), len1)
		assert.Equal(t, 10, m.Labels[0].Uncertainty)
		assert.Equal(t, "manual", m.Labels[0].LabelSrc)
	})
}

func TestPhoto_SetTitle(t *testing.T) {
	t.Run("empty title", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, "TitleToBeSet", m.PhotoTitle)
		m.SetTitle("", "manual")
		assert.Equal(t, "TitleToBeSet", m.PhotoTitle)
	})
	t.Run("title not from the same source", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, "TitleToBeSet", m.PhotoTitle)
		m.SetTitle("NewTitleSet", "image")
		assert.Equal(t, "TitleToBeSet", m.PhotoTitle)
	})
	t.Run("success", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, "TitleToBeSet", m.PhotoTitle)
		m.SetTitle("NewTitleSet", "location")
		assert.Equal(t, "NewTitleSet", m.PhotoTitle)
	})
}

func TestPhoto_SetDescription(t *testing.T) {
	t.Run("empty description", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, "photo description blacklist", m.PhotoDescription)
		m.SetDescription("", "manual")
		assert.Equal(t, "photo description blacklist", m.PhotoDescription)
	})
	t.Run("description not from the same source", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, "photo description blacklist", m.PhotoDescription)
		m.SetDescription("new photo description", "image")
		assert.Equal(t, "photo description blacklist", m.PhotoDescription)
	})
	t.Run("success", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, "photo description blacklist", m.PhotoDescription)
		m.SetDescription("new photo description", "manual")
		assert.Equal(t, "new photo description", m.PhotoDescription)
	})
}

func TestPhoto_SetTakenAt(t *testing.T) {
	t.Run("empty taken", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC), m.TakenAt)
		m.SetTakenAt(time.Time{}, time.Time{}, "", "manual")
		assert.Equal(t, time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC), m.TakenAt)
	})
	t.Run("taken not from the same source", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC), m.TakenAt)
		m.SetTakenAt(time.Date(2019, 12, 11, 9, 7, 18, 0, time.UTC),
			time.Date(2019, 12, 11, 9, 7, 18, 0, time.UTC), "", "image")
		assert.Equal(t, time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC), m.TakenAt)
	})
	t.Run("success", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC), m.TakenAt)
		assert.Equal(t, time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC), m.TakenAtLocal)
		m.SetTakenAt(time.Date(2019, 12, 11, 9, 7, 18, 0, time.UTC),
			time.Date(2019, 12, 11, 10, 7, 18, 0, time.UTC), "", "location")
		assert.Equal(t, time.Date(2019, 12, 11, 9, 7, 18, 0, time.UTC), m.TakenAt)
		assert.Equal(t, time.Date(2019, 12, 11, 10, 7, 18, 0, time.UTC), m.TakenAtLocal)
	})
	t.Run("success with empty takenAtLocal", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC), m.TakenAt)
		assert.Equal(t, time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC), m.TakenAtLocal)
		m.SetTakenAt(time.Date(2019, 12, 11, 9, 7, 18, 0, time.UTC),
			time.Time{}, "test", "location")
		assert.Equal(t, time.Date(2019, 12, 11, 9, 7, 18, 0, time.UTC), m.TakenAt)
		assert.Equal(t, time.Date(2019, 12, 11, 9, 7, 18, 0, time.UTC), m.TakenAtLocal)
	})
}

func TestPhoto_SetCoordinates(t *testing.T) {
	t.Run("empty coordinates", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, float32(1.234), m.PhotoLat)
		assert.Equal(t, float32(4.321), m.PhotoLng)
		assert.Equal(t, 3, m.PhotoAltitude)
		m.SetCoordinates(0, 0, 5, "manual")
		assert.Equal(t, float32(1.234), m.PhotoLat)
		assert.Equal(t, float32(4.321), m.PhotoLng)
		assert.Equal(t, 3, m.PhotoAltitude)
	})
	t.Run("different source", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, float32(1.234), m.PhotoLat)
		assert.Equal(t, float32(4.321), m.PhotoLng)
		assert.Equal(t, 3, m.PhotoAltitude)
		m.SetCoordinates(5.555, 5.555, 5, "image")
		assert.Equal(t, float32(1.234), m.PhotoLat)
		assert.Equal(t, float32(4.321), m.PhotoLng)
		assert.Equal(t, 3, m.PhotoAltitude)
	})
	t.Run("success", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, float32(1.234), m.PhotoLat)
		assert.Equal(t, float32(4.321), m.PhotoLng)
		assert.Equal(t, 3, m.PhotoAltitude)
		m.SetCoordinates(5.555, 5.555, 5, "location")
		assert.Equal(t, float32(5.555), m.PhotoLat)
		assert.Equal(t, float32(5.555), m.PhotoLng)
		assert.Equal(t, 5, m.PhotoAltitude)
	})
}

func TestPhoto_Delete(t *testing.T) {
	t.Run("not permanent", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo16")
		err := m.Delete(false)
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("permanent", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo16")
		err := m.Delete(true)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestPhotos_UIDs(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		photo1 := Photo{PhotoUID: "abc123"}
		photo2 := Photo{PhotoUID: "abc456"}
		photos := Photos{photo1, photo2}
		assert.Equal(t, []string{"abc123", "abc456"}, photos.UIDs())
	})
}

func TestPhoto_String(t *testing.T) {
	t.Run("return original", func(t *testing.T) {
		photo := Photo{PhotoUID: "", PhotoName: "", OriginalName: "holidayOriginal"}
		assert.Equal(t, "holidayOriginal", photo.String())
	})
	t.Run("unknown", func(t *testing.T) {
		photo := Photo{PhotoUID: "", PhotoName: "", OriginalName: ""}
		assert.Equal(t, "(unknown)", photo.String())
	})
}

func TestPhoto_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		photo := Photo{PhotoUID: "567", PhotoName: "Holiday", OriginalName: "holidayOriginal2"}
		err := photo.Create()
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestPhoto_Save(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		photo := Photo{PhotoUID: "567", PhotoName: "Holiday", OriginalName: "holidayOriginal2"}
		err := photo.Save()
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("error", func(t *testing.T) {
		photo := Photo{PhotoUID: "pt9jtdre2lvl0yh0"}
		assert.Error(t, photo.Save())
	})
}

func TestPhoto_Find(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		photo := Photo{PhotoUID: "567", ID: 55, PhotoName: "Holiday", OriginalName: "holidayOriginal2"}
		err := photo.Save()
		if err != nil {
			t.Fatal(err)
		}
		err = photo.Find()
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("success", func(t *testing.T) {
		photo := Photo{PhotoUID: "pt9jtdre2lvl0yh0"}
		err := photo.Find()
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("photo uid and id empty", func(t *testing.T) {
		photo := Photo{}
		assert.Error(t, photo.Find())
	})
	t.Run("error", func(t *testing.T) {
		photo := Photo{ID: 647487}
		assert.Error(t, photo.Find())
	})
	t.Run("error", func(t *testing.T) {
		photo := Photo{PhotoUID: "pt9jtdre2lvl0iuj"}
		assert.Error(t, photo.Find())
	})
}
func TestPhoto_RemoveKeyword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		keyword := Keyword{Keyword: "snake"}
		keyword2 := Keyword{Keyword: "otter"}
		keywords := []Keyword{keyword, keyword2}
		photo := &Photo{Keywords: keywords}
		err := photo.Save()
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 2, len(photo.Keywords))
		err = photo.RemoveKeyword("otter")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 2, len(photo.Keywords))
	})
}

func TestPhoto_SyncKeywordLabels(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		keyword := Keyword{Keyword: "snake"}
		keyword2 := Keyword{Keyword: "otter"}
		keywords := []Keyword{keyword, keyword2}
		label := Label{LabelName: "otter", LabelSlug: "otter"}
		var deleteTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		label2 := Label{LabelName: "snake", LabelSlug: "snake", DeletedAt: &deleteTime}
		photo := &Photo{ID: 34567, Keywords: keywords}
		err := photo.Save()
		if err != nil {
			t.Fatal(err)
		}
		err = label.Save()
		if err != nil {
			t.Fatal(err)
		}
		err = label2.Save()
		if err != nil {
			t.Fatal(err)
		}
		err = photo.SyncKeywordLabels()
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 2, len(photo.Keywords))
	})
}

func TestPhoto_LocationLoaded(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		photo := Photo{PhotoUID: "56798", PhotoName: "Holiday", OriginalName: "holidayOriginal2"}
		assert.False(t, photo.LocationLoaded())
	})
	t.Run("false", func(t *testing.T) {
		location := &Cell{Place: nil}
		photo := Photo{PhotoName: "Holiday", Cell: location}
		assert.False(t, photo.LocationLoaded())
	})
}

func TestPhoto_LoadLocation(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		photo := PhotoFixtures.Get("Photo03")
		err := photo.LoadLocation()
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("unknown location", func(t *testing.T) {
		location := &Cell{Place: nil}
		photo := Photo{PhotoName: "Holiday", Cell: location}
		assert.Error(t, photo.LoadLocation())
	})
}

func TestPhoto_PlaceLoaded(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		photo := Photo{PhotoUID: "56798", PhotoName: "Holiday", OriginalName: "holidayOriginal2"}
		assert.False(t, photo.PlaceLoaded())
	})
}

func TestPhoto_LoadPlace(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		photo := PhotoFixtures.Get("Photo03")
		err := photo.LoadPlace()
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("unknown location", func(t *testing.T) {
		location := &Cell{Place: nil}
		photo := Photo{PhotoName: "Holiday", Cell: location}
		assert.Error(t, photo.LoadPlace())
	})
}

func TestPhoto_HasDescription(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		photo := Photo{PhotoDescription: ""}
		assert.False(t, photo.HasDescription())
	})
	t.Run("true", func(t *testing.T) {
		photo := Photo{PhotoDescription: "bcss"}
		assert.True(t, photo.HasDescription())
	})
}

func TestPhoto_NoDescription(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		photo := Photo{PhotoDescription: ""}
		assert.True(t, photo.NoDescription())
	})
	t.Run("false", func(t *testing.T) {
		photo := Photo{PhotoDescription: "bcss"}
		assert.False(t, photo.NoDescription())
	})
}

func TestPhoto_AllFilesMissing(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		photo := Photo{ID: 6969866}
		assert.True(t, photo.AllFilesMissing())
	})
}

func TestPhoto_Updates(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		photo := Photo{PhotoDescription: "bcss", PhotoName: "InitialName"}
		photo.Save()
		assert.Equal(t, "InitialName", photo.PhotoName)
		assert.Equal(t, "bcss", photo.PhotoDescription)

		err := photo.Updates(Photo{PhotoName: "UpdatedName", PhotoDescription: "UpdatedDesc"})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "UpdatedName", photo.PhotoName)
		assert.Equal(t, "UpdatedDesc", photo.PhotoDescription)

	})
}

func TestPhoto_SetFavorite(t *testing.T) {
	t.Run("set to true", func(t *testing.T) {
		photo := Photo{PhotoFavorite: true}
		photo.Save()

		err := photo.SetFavorite(false)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, photo.PhotoFavorite)
	})
	t.Run("set to false", func(t *testing.T) {
		photo := Photo{PhotoFavorite: false}
		photo.Save()

		err := photo.SetFavorite(true)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, photo.PhotoFavorite)
	})
}

func TestPhoto_Approve(t *testing.T) {
	t.Run("quality = 4", func(t *testing.T) {
		photo := Photo{PhotoQuality: 4}
		photo.Save()

		err := photo.Approve()
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 4, photo.PhotoQuality)
	})
	t.Run("quality = 1", func(t *testing.T) {
		photo := Photo{PhotoQuality: 1}
		photo.Save()

		err := photo.Approve()
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 3, photo.PhotoQuality)
	})
}

func TestPhoto_Links(t *testing.T) {
	t.Run("1 result", func(t *testing.T) {
		photo := Photo{PhotoUID: "pt9k3pw1wowuy3c3"}
		links := photo.Links()
		assert.Equal(t, "7jxf3jfn2k", links[0].LinkToken)
	})
}
