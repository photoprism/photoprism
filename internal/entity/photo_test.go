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
			TitleSrc:         SrcManual,
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
			CameraSrc:        SrcMeta,
			LensID:           uint(6),
			CellID:           "1234",
			PlaceSrc:         SrcManual,
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

		m := PhotoFixtures.Get("Photo08")

		if err := SavePhotoForm(m, f); err != nil {
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

func TestPhoto_AddLabels(t *testing.T) {
	t.Run("add label", func(t *testing.T) {
		m := PhotoFixtures.Get("19800101_000002_D640C559")
		classifyLabels := classify.Labels{{Name: "cactus", Uncertainty: 30, Source: SrcManual, Priority: 5, Categories: []string{"plant"}}}
		len1 := len(m.Labels)
		m.AddLabels(classifyLabels)
		assert.Greater(t, len(m.Labels), len1)
	})
	t.Run("update label", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		classifyLabels := classify.Labels{{Name: "landscape", Uncertainty: 10, Source: SrcManual, Priority: 5, Categories: []string{"plant"}}}
		assert.Equal(t, 20, m.Labels[0].Uncertainty)
		assert.Equal(t, SrcImage, m.Labels[0].LabelSrc)
		len1 := len(m.Labels)
		m.AddLabels(classifyLabels)
		assert.Equal(t, len(m.Labels), len1)
		assert.Equal(t, 10, m.Labels[0].Uncertainty)
		assert.Equal(t, SrcManual, m.Labels[0].LabelSrc)
	})
}

func TestPhoto_SetDescription(t *testing.T) {
	t.Run("empty description", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, "photo description blacklist", m.PhotoDescription)
		m.SetDescription("", SrcManual)
		assert.Equal(t, "photo description blacklist", m.PhotoDescription)
	})
	t.Run("description not from the same source", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, "photo description blacklist", m.PhotoDescription)
		m.SetDescription("new photo description", SrcName)
		assert.Equal(t, "photo description blacklist", m.PhotoDescription)
	})
	t.Run("success", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, "photo description blacklist", m.PhotoDescription)
		m.SetDescription("new photo description", SrcMeta)
		assert.Equal(t, "new photo description", m.PhotoDescription)
	})
}

func TestPhoto_SetTakenAt(t *testing.T) {
	t.Run("empty taken", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC), m.TakenAt)
		m.SetTakenAt(time.Time{}, time.Time{}, "", SrcManual)
		assert.Equal(t, time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC), m.TakenAt)
	})
	t.Run("taken not from the same source", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC), m.TakenAt)
		m.SetTakenAt(time.Date(2019, 12, 11, 9, 7, 18, 0, time.UTC),
			time.Date(2019, 12, 11, 9, 7, 18, 0, time.UTC), "", SrcAuto)
		assert.Equal(t, time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC), m.TakenAt)
	})
	t.Run("from name", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		m.TimeZone = ""
		m.TakenSrc = SrcAuto

		assert.Equal(t, time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC), m.TakenAt)
		assert.Equal(t, time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC), m.TakenAtLocal)
		assert.Equal(t, "", m.TimeZone)
		assert.Equal(t, SrcAuto, m.TakenSrc)

		m.SetTakenAt(time.Date(2011, 12, 11, 9, 7, 18, 0, time.UTC),
			time.Date(2019, 11, 11, 10, 7, 18, 0, time.UTC), "America/New_York", SrcName)

		assert.Equal(t, "", m.TimeZone)
		assert.Equal(t, SrcName, m.TakenSrc)

		assert.Equal(t, time.Date(2011, 12, 11, 9, 7, 18, 0, time.UTC), m.TakenAt)
		assert.Equal(t, time.Date(2019, 11, 11, 10, 7, 18, 0, time.UTC), m.TakenAtLocal)
	})
	t.Run("success", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC), m.TakenAt)
		assert.Equal(t, time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC), m.TakenAtLocal)

		m.SetTakenAt(time.Date(2019, 12, 11, 9, 7, 18, 0, time.UTC),
			time.Date(2019, 12, 11, 10, 7, 18, 0, time.UTC), "", SrcMeta)

		assert.Equal(t, time.Date(2019, 12, 11, 9, 7, 18, 0, time.UTC), m.TakenAt)
		assert.Equal(t, time.Date(2019, 12, 11, 10, 7, 18, 0, time.UTC), m.TakenAtLocal)
	})
	t.Run("fallback", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, time.Date(2013, time.November, 11, 9, 7, 18, 0, time.UTC), m.TakenAt)
		assert.Equal(t, time.Date(2013, time.November, 11, 9, 7, 18, 0, time.UTC), m.TakenAtLocal)

		t.Logf("SRC, ZONE, UTC, LOCAL: %s / %s / %s /%s", m.TakenSrc, m.TimeZone, m.TakenAt, m.TakenAtLocal)

		m.SetTakenAt(time.Date(2019, time.December, 11, 9, 7, 18, 0, time.UTC),
			time.Date(2019, time.December, 11, 10, 7, 18, 0, time.UTC), "", SrcAuto)

		t.Logf("SRC, ZONE, UTC, LOCAL: %s / %s / %s /%s", m.TakenSrc, m.TimeZone, m.TakenAt, m.TakenAtLocal)

		assert.Equal(t, time.Date(2013, time.November, 11, 9, 7, 18, 0, time.UTC), m.TakenAt)
		assert.Equal(t, time.Date(2013, time.November, 11, 9, 7, 18, 0, time.UTC), m.TakenAtLocal)

		newTime := time.Date(2013, time.November, 11, 9, 7, 18, 0, time.UTC)

		expected := time.Date(2013, time.November, 11, 8, 7, 18, 0, time.UTC)

		m.TimeZone = "Europe/Berlin"

		m.SetTakenAt(newTime, newTime, "", SrcName)

		assert.Equal(t, expected, m.TakenAt)
		assert.Equal(t, m.GetTakenAtLocal(), m.TakenAtLocal)
	})
	t.Run("time zone", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")

		zone := "Europe/Berlin"

		loc, _ := time.LoadLocation(zone)

		newTime := time.Date(2013, 11, 11, 9, 7, 18, 0, loc)

		m.SetTakenAt(newTime, newTime, zone, SrcName)

		assert.Equal(t, newTime.UTC(), m.TakenAt)
		assert.Equal(t, newTime, m.TakenAtLocal)
	})
	t.Run("time > max year", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC), m.TakenAt)
		assert.Equal(t, time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC), m.TakenAtLocal)
		m.SetTakenAt(time.Date(2123, 12, 11, 9, 7, 18, 0, time.UTC),
			time.Date(2123, 12, 11, 10, 7, 18, 0, time.UTC), "", SrcManual)
		assert.Equal(t, time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC), m.TakenAt)
		assert.Equal(t, time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC), m.TakenAtLocal)
	})
	t.Run("success with empty takenAtLocal", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC), m.TakenAt)
		assert.Equal(t, time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC), m.TakenAtLocal)
		m.SetTakenAt(time.Date(2019, 12, 11, 9, 7, 18, 0, time.UTC),
			time.Time{}, "test", SrcXmp)
		assert.Equal(t, time.Date(2019, 12, 11, 9, 7, 18, 0, time.UTC), m.TakenAt)
		assert.Equal(t, time.Date(2019, 12, 11, 9, 7, 18, 0, time.UTC), m.TakenAtLocal)
	})
	t.Run("don't update older date", func(t *testing.T) {
		photo := &Photo{TakenAt: time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC)}
		photo.SetTakenAt(time.Date(2014, 12, 11, 9, 7, 18, 0, time.UTC),
			time.Date(2014, 12, 11, 10, 7, 18, 0, time.UTC), "", SrcAuto)
		assert.Equal(t, time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC), photo.TakenAt)
	})
	t.Run("set local time from utc", func(t *testing.T) {
		photo := &Photo{TakenAt: time.Date(2015, 11, 11, 9, 7, 18, 0, time.UTC), TimeZone: "Europe/Berlin"}
		photo.SetTakenAt(time.Date(2014, 12, 11, 9, 7, 18, 0, time.UTC),
			time.Date(2014, 12, 11, 10, 7, 18, 0, time.UTC), time.UTC.String(), SrcManual)
		assert.Equal(t, time.Date(2014, 12, 11, 9, 7, 18, 0, time.UTC), photo.TakenAt)
		assert.Equal(t, time.Date(2014, 12, 11, 10, 07, 18, 0, time.UTC), photo.TakenAtLocal)
	})
	t.Run("local is UTC", func(t *testing.T) {
		photo := &Photo{TakenAt: time.Date(2015, 11, 11, 9, 7, 18, 0, time.UTC), TimeZone: time.UTC.String()}
		photo.SetTakenAt(time.Date(2014, 12, 11, 9, 7, 18, 0, time.UTC),
			time.Date(2014, 12, 11, 10, 7, 18, 0, time.UTC), "", SrcManual)
		assert.Equal(t, time.Date(2014, 12, 11, 9, 7, 18, 0, time.UTC), photo.TakenAt)
		assert.Equal(t, time.Date(2014, 12, 11, 9, 07, 18, 0, time.UTC), photo.TakenAtLocal)
	})
}

func TestPhoto_UpdateTimeZone(t *testing.T) {
	t.Run("PhotoTimeZone", func(t *testing.T) {
		m := PhotoFixtures.Get("PhotoTimeZone")

		takenLocal := time.Date(2015, time.May, 17, 23, 2, 46, 0, time.UTC)
		takenJerusalemUtc := time.Date(2015, time.May, 17, 20, 2, 46, 0, time.UTC)
		takenShanghaiUtc := time.Date(2015, time.May, 17, 15, 2, 46, 0, time.UTC)

		assert.Equal(t, "", m.TimeZone)
		assert.Equal(t, takenLocal, m.TakenAt)
		assert.Equal(t, takenLocal, m.TakenAtLocal)

		zone1 := "Asia/Jerusalem"

		m.UpdateTimeZone(zone1)

		assert.Equal(t, zone1, m.TimeZone)
		assert.Equal(t, takenJerusalemUtc, m.TakenAt)
		assert.Equal(t, takenLocal, m.TakenAtLocal)

		m.UpdateTimeZone(zone1)

		assert.Equal(t, zone1, m.TimeZone)
		assert.Equal(t, takenJerusalemUtc, m.TakenAt)
		assert.Equal(t, takenLocal, m.TakenAtLocal)

		zone2 := "Asia/Shanghai"

		m.UpdateTimeZone(zone2)

		assert.Equal(t, zone2, m.TimeZone)
		assert.Equal(t, takenShanghaiUtc, m.TakenAt)
		assert.Equal(t, takenLocal, m.TakenAtLocal)

		zone3 := "UTC"

		m.UpdateTimeZone(zone3)

		assert.Equal(t, zone2, m.TimeZone)
		assert.Equal(t, takenShanghaiUtc, m.TakenAt)
		assert.Equal(t, takenLocal, m.TakenAtLocal)
	})
	t.Run("VideoTimeZone", func(t *testing.T) {
		m := PhotoFixtures.Get("VideoTimeZone")

		takenUtc := time.Date(2015, 5, 17, 17, 48, 46, 0, time.UTC)
		takenJerusalem := time.Date(2015, time.May, 17, 20, 48, 46, 0, time.UTC)
		takenShanghaiUtc := time.Date(2015, time.May, 17, 12, 48, 46, 0, time.UTC)

		assert.Equal(t, "UTC", m.TimeZone)
		assert.Equal(t, takenUtc, m.TakenAt)
		assert.Equal(t, takenUtc, m.TakenAtLocal)

		zone1 := "Asia/Jerusalem"

		m.UpdateTimeZone(zone1)

		assert.Equal(t, zone1, m.TimeZone)
		assert.Equal(t, takenUtc, m.TakenAt)
		assert.Equal(t, takenJerusalem, m.TakenAtLocal)

		m.UpdateTimeZone(zone1)

		assert.Equal(t, zone1, m.TimeZone)
		assert.Equal(t, takenUtc, m.TakenAt)
		assert.Equal(t, takenJerusalem, m.TakenAtLocal)

		zone2 := "Asia/Shanghai"

		m.UpdateTimeZone(zone2)

		assert.Equal(t, zone2, m.TimeZone)
		assert.Equal(t, takenShanghaiUtc, m.TakenAt)
		assert.Equal(t, takenJerusalem, m.TakenAtLocal)

		zone3 := "UTC"

		m.UpdateTimeZone(zone3)

		assert.Equal(t, zone2, m.TimeZone)
		assert.Equal(t, takenShanghaiUtc, m.TakenAt)
		assert.Equal(t, takenJerusalem, m.TakenAtLocal)
	})
	t.Run("UTC", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo12")
		m.TimeZone = "UTC"

		zone := "Europe/Berlin"

		takenAt := m.TakenAt
		takenAtLocal := m.TakenAtLocal

		assert.Equal(t, takenAt, m.TakenAt)
		assert.Equal(t, takenAtLocal, m.TakenAtLocal)

		m.UpdateTimeZone(zone)

		assert.Equal(t, takenAt, m.TakenAt)
		assert.Equal(t, m.GetTakenAtLocal(), m.TakenAtLocal)
	})

	t.Run("Europe/Berlin", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo12")

		zone := "Europe/Berlin"

		takenAt := m.TakenAt
		takenAtLocal := m.TakenAtLocal

		assert.Equal(t, takenAt, m.TakenAt)
		assert.Equal(t, takenAtLocal, m.TakenAtLocal)
		assert.Equal(t, "", m.TimeZone)

		m.UpdateTimeZone(zone)

		assert.Equal(t, m.GetTakenAt(), m.TakenAt)
		assert.Equal(t, takenAtLocal, m.TakenAtLocal)
	})

	t.Run("America/New_York", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo12")
		m.TimeZone = "Europe/Berlin"
		m.TakenAt = m.GetTakenAt()

		zone := "America/New_York"

		takenAt := m.TakenAt
		takenAtLocal := m.TakenAtLocal

		assert.Equal(t, takenAt, m.TakenAt)
		assert.Equal(t, takenAtLocal, m.TakenAtLocal)
		assert.Equal(t, "Europe/Berlin", m.TimeZone)

		m.UpdateTimeZone(zone)

		assert.Equal(t, m.GetTakenAt(), m.TakenAt)
		assert.Equal(t, takenAtLocal, m.TakenAtLocal)
	})

	t.Run("manual", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo12")
		m.TimeZone = "Europe/Berlin"
		m.TakenAt = m.GetTakenAt()
		m.TakenSrc = SrcManual

		zone := "America/New_York"

		takenAt := m.TakenAt
		takenAtLocal := m.TakenAtLocal

		assert.Equal(t, takenAt, m.TakenAt)
		assert.Equal(t, takenAtLocal, m.TakenAtLocal)
		assert.Equal(t, "Europe/Berlin", m.TimeZone)

		m.UpdateTimeZone(zone)

		assert.Equal(t, takenAt, m.TakenAt)
		assert.Equal(t, takenAtLocal, m.TakenAtLocal)
		assert.Equal(t, "Europe/Berlin", m.TimeZone)
	})
	t.Run("zone = UTC", func(t *testing.T) {
		photo := &Photo{TakenAt: time.Date(2015, 11, 11, 9, 7, 18, 0, time.UTC), TimeZone: "Europe/Berlin"}
		photo.UpdateTimeZone("")
		assert.Equal(t, time.Date(2015, 11, 11, 9, 7, 18, 0, time.UTC), photo.TakenAt)
		assert.Equal(t, "Europe/Berlin", photo.TimeZone)
	})
}

func TestPhoto_SetAltitude(t *testing.T) {
	t.Run("ViaSetCoordinates", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(1.234), m.PhotoLat)
		assert.Equal(t, float32(4.321), m.PhotoLng)
		assert.Equal(t, 3, m.PhotoAltitude)

		m.SetCoordinates(0, 0, 5, SrcManual)
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(1.234), m.PhotoLat)
		assert.Equal(t, float32(4.321), m.PhotoLng)
		assert.Equal(t, 5, m.PhotoAltitude)
	})
	t.Run("Update", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(1.234), m.PhotoLat)
		assert.Equal(t, float32(4.321), m.PhotoLng)
		assert.Equal(t, 3, m.PhotoAltitude)

		m.SetAltitude(5, SrcManual)
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(1.234), m.PhotoLat)
		assert.Equal(t, float32(4.321), m.PhotoLng)
		assert.Equal(t, 5, m.PhotoAltitude)
	})
	t.Run("SkipUpdate", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(1.234), m.PhotoLat)
		assert.Equal(t, float32(4.321), m.PhotoLng)
		assert.Equal(t, 3, m.PhotoAltitude)

		m.SetAltitude(5, SrcEstimate)
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(1.234), m.PhotoLat)
		assert.Equal(t, float32(4.321), m.PhotoLng)
		assert.Equal(t, 3, m.PhotoAltitude)
	})
	t.Run("UpdateEmptyAltitude", func(t *testing.T) {
		m := Photo{ID: 1, PlaceSrc: SrcMeta, PhotoLat: float32(1.234), PhotoLng: float32(4.321), PhotoAltitude: 0}

		m.SetAltitude(-5, SrcAuto)
		assert.Equal(t, 0, m.PhotoAltitude)

		m.SetAltitude(-5, SrcEstimate)
		assert.Equal(t, 0, m.PhotoAltitude)

		m.SetAltitude(-5, SrcMeta)
		assert.Equal(t, -5, m.PhotoAltitude)
	})
	t.Run("ZeroAltitudeManual", func(t *testing.T) {
		m := Photo{ID: 1, PlaceSrc: SrcManual, PhotoLat: float32(1.234), PhotoLng: float32(4.321), PhotoAltitude: 5}

		m.SetAltitude(0, SrcManual)
		assert.Equal(t, 0, m.PhotoAltitude)
	})
}

func TestPhoto_SetCoordinates(t *testing.T) {
	t.Run("empty coordinates", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(1.234), m.PhotoLat)
		assert.Equal(t, float32(4.321), m.PhotoLng)
		assert.Equal(t, 3, m.PhotoAltitude)

		m.SetCoordinates(0, 0, 5, SrcManual)
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(1.234), m.PhotoLat)
		assert.Equal(t, float32(4.321), m.PhotoLng)
		assert.Equal(t, 5, m.PhotoAltitude)
	})
	t.Run("same source new values", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(1.234), m.PhotoLat)
		assert.Equal(t, float32(4.321), m.PhotoLng)
		assert.Equal(t, 3, m.PhotoAltitude)

		m.SetCoordinates(5.555, 5.555, 5, SrcMeta)
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(5.555), m.PhotoLat)
		assert.Equal(t, float32(5.555), m.PhotoLng)
		assert.Equal(t, 5, m.PhotoAltitude)
	})
	t.Run("different source lower priority", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(1.234), m.PhotoLat)
		assert.Equal(t, float32(4.321), m.PhotoLng)
		assert.Equal(t, 3, m.PhotoAltitude)

		m.SetCoordinates(5.555, 5.555, 5, SrcName)
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(1.234), m.PhotoLat)
		assert.Equal(t, float32(4.321), m.PhotoLng)
		assert.Equal(t, 3, m.PhotoAltitude)
	})
	t.Run("different source equal priority", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(1.234), m.PhotoLat)
		assert.Equal(t, float32(4.321), m.PhotoLng)
		assert.Equal(t, 3, m.PhotoAltitude)

		m.SetCoordinates(5.555, 5.555, 5, SrcKeyword)
		assert.Equal(t, float32(5.555), m.PhotoLat)
		assert.Equal(t, float32(5.555), m.PhotoLng)
		assert.Equal(t, 5, m.PhotoAltitude)
	})
	t.Run("different source higher priority", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo21")
		assert.Equal(t, SrcEstimate, m.PlaceSrc)
		assert.Equal(t, float32(0), m.PhotoLat)
		assert.Equal(t, float32(0), m.PhotoLng)
		assert.Equal(t, 0, m.PhotoAltitude)

		m.SetCoordinates(5.555, 5.555, 5, SrcMeta)
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(5.555), m.PhotoLat)
		assert.Equal(t, float32(5.555), m.PhotoLng)
		assert.Equal(t, 5, m.PhotoAltitude)
	})
	t.Run("different source highest priority (manual)", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(1.234), m.PhotoLat)
		assert.Equal(t, float32(4.321), m.PhotoLng)
		assert.Equal(t, 3, m.PhotoAltitude)

		m.SetCoordinates(5.555, 5.555, 5, SrcManual)
		assert.Equal(t, SrcManual, m.PlaceSrc)
		assert.Equal(t, float32(5.555), m.PhotoLat)
		assert.Equal(t, float32(5.555), m.PhotoLng)
		assert.Equal(t, 5, m.PhotoAltitude)
	})
}

func TestPhoto_Delete(t *testing.T) {
	t.Run("not permanent", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo16")
		files, err := m.Delete(false)
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, files, 1)
	})
	t.Run("permanent", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo16")
		files, err := m.Delete(true)
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, files, 1)
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
		photo := Photo{PhotoUID: "pt9atdre2lvl0yhx", PhotoName: "Holiday", OriginalName: "holidayOriginal2"}
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
	t.Run("known location", func(t *testing.T) {
		location := CellFixtures.Pointer("mexico")
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

		if err := photo.Save(); err != nil {
			t.Fatal(err)
		}

		if err := photo.Approve(); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 4, photo.PhotoQuality)
	})
	t.Run("quality = 1", func(t *testing.T) {
		photo := Photo{PhotoQuality: 1}

		if err := photo.Save(); err != nil {
			t.Fatal(err)
		}

		if err := photo.Approve(); err != nil {
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

func TestPhoto_SetPrimary(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := PhotoFixtures.Get("19800101_000002_D640C559")

		if err := m.SetPrimary(""); err != nil {
			t.Fatal(err)
		}
	})
}

func TestMapKey(t *testing.T) {
	assert.Equal(t, "ogh006/abc236", MapKey(time.Date(2016, 11, 11, 9, 7, 18, 0, time.UTC), "abc236"))
}

func TestNewPhoto(t *testing.T) {
	t.Run("stackable", func(t *testing.T) {
		m := NewPhoto(true)
		assert.Equal(t, IsStackable, m.PhotoStack)
	})
	t.Run("not stackable", func(t *testing.T) {
		m := NewPhoto(false)
		assert.Equal(t, IsUnstacked, m.PhotoStack)
	})
}

func TestPhoto_FirstOrCreate(t *testing.T) {
	t.Run("photo already existing", func(t *testing.T) {
		photo := Photo{PhotoUID: "567454", PhotoName: "Light", OriginalName: "lightBlub.jpg"}
		err := photo.Save()
		if err != nil {
			t.Fatal(err)
		}
		err2 := photo.Find()
		if err2 != nil {
			t.Fatal(err2)
		}
		err3 := photo.FirstOrCreate()
		if err3 != nil {
			t.Fatal(err3)
		}
	})
	t.Run("photo not yet existing", func(t *testing.T) {
		photo := Photo{PhotoUID: "567459", PhotoName: "Light2", OriginalName: "lightBlub2.jpg"}
		err3 := photo.FirstOrCreate()
		if err3 != nil {
			t.Fatal(err3)
		}
		err2 := photo.Find()
		if err2 != nil {
			t.Fatal(err2)
		}
	})
}

func TestPhoto_UnknownCamera(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		photo := Photo{}
		assert.True(t, photo.UnknownCamera())
	})
	t.Run("false", func(t *testing.T) {
		photo := Photo{CameraID: 100000}
		assert.False(t, photo.UnknownCamera())
	})
}

func TestPhoto_UnknownLens(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		photo := Photo{}
		assert.True(t, photo.UnknownLens())
	})
	t.Run("false", func(t *testing.T) {
		photo := Photo{LensID: 100000}
		assert.False(t, photo.UnknownLens())
	})
}

func TestPhoto_UpdateDateFields(t *testing.T) {
	t.Run("year < 1000", func(t *testing.T) {
		photo := &Photo{TakenAt: time.Date(900, 11, 11, 9, 7, 18, 0, time.UTC)}
		photo.UpdateDateFields()
		assert.Equal(t, time.Date(900, 11, 11, 9, 7, 18, 0, time.UTC), photo.TakenAt)
		assert.Empty(t, photo.TakenAtLocal)
	})
	t.Run("set to unknown", func(t *testing.T) {
		photo := &Photo{TakenAt: time.Date(1900, 11, 11, 9, 7, 18, 0, time.UTC), TakenSrc: SrcAuto, CreatedAt: time.Date(1900, 11, 11, 5, 7, 18, 0, time.UTC)}
		photo.UpdateDateFields()
		assert.Equal(t, UnknownYear, photo.PhotoYear)
	})
}

func TestPhoto_SetCamera(t *testing.T) {
	t.Run("camera nil", func(t *testing.T) {
		photo := &Photo{}
		photo.SetCamera(nil, SrcAuto)
		assert.Empty(t, photo.Camera)
	})
	t.Run("camera unknown", func(t *testing.T) {
		photo := &Photo{}
		camera := &Camera{CameraSlug: ""}
		photo.SetCamera(camera, SrcAuto)
		assert.Empty(t, photo.Camera)
	})
	t.Run("do not overwrite manual changes", func(t *testing.T) {
		cameraOld := &Camera{CameraSlug: "OldCamera", ID: 10000000111}
		photo := &Photo{CameraSrc: SrcManual, Camera: cameraOld, CameraID: 10000000111}
		assert.Equal(t, "OldCamera", photo.Camera.CameraSlug)
		assert.Equal(t, SrcManual, photo.CameraSrc)
		assert.False(t, photo.UnknownCamera())
		camera := &Camera{CameraSlug: "NewCamera"}
		photo.SetCamera(camera, SrcAuto)
		assert.Equal(t, "OldCamera", photo.Camera.CameraSlug)
	})
	t.Run("set new camera", func(t *testing.T) {
		cameraOld := &Camera{CameraSlug: "OldCamera", ID: 10000000111}
		photo := &Photo{CameraSrc: SrcAuto, Camera: cameraOld, CameraID: 10000000111}
		assert.Equal(t, "OldCamera", photo.Camera.CameraSlug)
		camera := &Camera{CameraSlug: "NewCamera"}
		photo.SetCamera(camera, SrcMeta)
		assert.Equal(t, "NewCamera", photo.Camera.CameraSlug)
	})
}

func TestPhoto_SetLens(t *testing.T) {
	t.Run("lens nil", func(t *testing.T) {
		photo := &Photo{}
		photo.SetLens(nil, SrcAuto)
		assert.Empty(t, photo.Lens)
	})
	t.Run("lens unknown", func(t *testing.T) {
		photo := &Photo{}
		lens := &Lens{LensSlug: ""}
		photo.SetLens(lens, SrcAuto)
		assert.Empty(t, photo.Lens)
	})
	t.Run("do not overwrite manual changes", func(t *testing.T) {
		lensOld := &Lens{LensSlug: "OldLens", ID: 10000000111}
		photo := &Photo{CameraSrc: SrcManual, Lens: lensOld, LensID: 10000000111}
		assert.Equal(t, "OldLens", photo.Lens.LensSlug)
		lens := &Lens{LensSlug: "NewLens"}
		photo.SetLens(lens, SrcAuto)
		assert.Equal(t, "OldLens", photo.Lens.LensSlug)
	})
	t.Run("set new camera", func(t *testing.T) {
		lensOld := &Lens{LensSlug: "OldLens", ID: 10000000111}
		photo := &Photo{CameraSrc: SrcAuto, Lens: lensOld, LensID: 10000000111}
		assert.Equal(t, "OldLens", photo.Lens.LensSlug)
		lens := &Lens{LensSlug: "NewLens"}
		photo.SetLens(lens, SrcMeta)
		assert.Equal(t, "NewLens", photo.Lens.LensSlug)
	})
}

func TestPhoto_SetExposure(t *testing.T) {
	t.Run("changes have priority", func(t *testing.T) {
		photo := &Photo{PhotoFocalLength: 5, PhotoFNumber: 3, PhotoIso: 300, PhotoExposure: "45", CameraSrc: SrcMeta}
		photo.SetExposure(8, 9, 500, "66", SrcManual)
		assert.Equal(t, 8, photo.PhotoFocalLength)
		assert.Equal(t, float32(9), photo.PhotoFNumber)
		assert.Equal(t, 500, photo.PhotoIso)
		assert.Equal(t, "66", photo.PhotoExposure)
	})
	t.Run("changes have no priority", func(t *testing.T) {
		photo := &Photo{PhotoFocalLength: 5, PhotoFNumber: 3, PhotoIso: 300, PhotoExposure: "45", CameraSrc: SrcManual}
		photo.SetExposure(8, 9, 500, "66", SrcMeta)
		assert.Equal(t, 5, photo.PhotoFocalLength)
		assert.Equal(t, float32(3), photo.PhotoFNumber)
		assert.Equal(t, 300, photo.PhotoIso)
		assert.Equal(t, "45", photo.PhotoExposure)
	})
}

func TestPhoto_AllFiles(t *testing.T) {
	t.Run("photo with files", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		files := m.AllFiles()
		assert.Equal(t, 2, len(files))
	})
	t.Run("photo without files", func(t *testing.T) {
		m := &Photo{ID: 100000023456}
		files := m.AllFiles()
		assert.Equal(t, 0, len(files))
	})
}

func TestPhoto_Archive(t *testing.T) {
	t.Run("archive not yet archived photo", func(t *testing.T) {
		m := &Photo{PhotoTitle: "HappyLilly"}
		assert.Empty(t, m.DeletedAt)
		err := m.Archive()
		if err != nil {
			t.Fatal(err)
		}
		assert.NotEmpty(t, m.DeletedAt)
		err = m.Restore()
		if err != nil {
			t.Fatal(err)
		}
		assert.Empty(t, m.DeletedAt)
	})
}

func TestPhoto_SetCameraSerial(t *testing.T) {
	m := &Photo{}
	assert.Empty(t, m.CameraSerial)
	m.SetCameraSerial("abcCamera")
	assert.Equal(t, "abcCamera", m.CameraSerial)
}

func TestPhoto_MapKey(t *testing.T) {
	m := &Photo{TakenAt: time.Date(2016, 11, 11, 9, 7, 18, 0, time.UTC), CellID: "abc236"}
	assert.Equal(t, "ogh006/abc236", m.MapKey())
}
