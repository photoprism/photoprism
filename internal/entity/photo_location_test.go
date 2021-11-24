package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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

func TestPhoto_UnknownLocation(t *testing.T) {
	t.Run("no_location", func(t *testing.T) {
		m := PhotoFixtures.Get("19800101_000002_D640C559")
		assert.True(t, m.UnknownLocation())
	})

	t.Run("no_lat_lng", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo08")
		m.PhotoLat = 0.0
		m.PhotoLng = 0.0
		// t.Logf("MODEL: %+v", m)
		assert.False(t, m.HasLocation())
		assert.True(t, m.UnknownLocation())
	})

	t.Run("lat_lng_cell_id", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo08")
		// t.Logf("MODEL: %+v", m)
		assert.True(t, m.HasLocation())
		assert.False(t, m.UnknownLocation())
	})
}

func TestPhoto_TrustedLocation(t *testing.T) {
	t.Run("SrcAuto", func(t *testing.T) {
		m := Photo{ID: 1, CellID: "s2:479a03fda18c", PhotoLat: 1, PhotoLng: -1, PlaceSrc: SrcAuto}
		assert.False(t, m.TrustedLocation())
	})
	t.Run("SrcEstimate", func(t *testing.T) {
		m := Photo{ID: 1, CellID: "s2:479a03fda18c", PhotoLat: 1, PhotoLng: -1, PlaceSrc: SrcEstimate}
		assert.False(t, m.TrustedLocation())
	})
	t.Run("SrcMetaTrue", func(t *testing.T) {
		m := Photo{ID: 1, CellID: "s2:479a03fda18c", PhotoLat: 1, PhotoLng: -1, PlaceSrc: SrcMeta}
		assert.True(t, m.TrustedLocation())
	})
	t.Run("SrcMetaFalse", func(t *testing.T) {
		m := Photo{ID: 1, CellID: "s2:479a03fda18c", PhotoLat: 0, PhotoLng: 0, PlaceSrc: SrcMeta}
		assert.False(t, m.TrustedLocation())
	})
}

func TestPhoto_HasLocation(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		m := PhotoFixtures.Get("19800101_000002_D640C559")
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
		m := PhotoFixtures.Get("Photo09")
		assert.False(t, m.HasLatLng())
	})
}

func TestPhoto_NoLatLng(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		assert.False(t, m.NoLatLng())
	})
	t.Run("true", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo09")
		assert.True(t, m.NoLatLng())
	})
}

func TestPhoto_NoPlace(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		m := PhotoFixtures.Get("19800101_000002_D640C559")
		assert.True(t, m.UnknownPlace())
	})
	t.Run("false", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo08")
		assert.False(t, m.UnknownPlace())
	})
}

func TestPhoto_HasPlace(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		m := PhotoFixtures.Get("19800101_000002_D640C559")
		assert.False(t, m.HasPlace())
	})
	t.Run("true", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo08")
		assert.True(t, m.HasPlace())
	})
}

func TestPhoto_GetTimeZone(t *testing.T) {
	m := Photo{}
	m.PhotoLat = 48.533905555
	m.PhotoLng = 9.01

	result := m.GetTimeZone()

	if result != "Europe/Berlin" {
		t.Fatalf("time zone should be Europe/Berlin: %s", result)
	}
}

func TestPhoto_GetTakenAt(t *testing.T) {
	m := Photo{}
	m.PhotoLat = 48.533905555
	m.PhotoLng = 9.01
	m.TakenAt, _ = time.Parse(time.RFC3339, "2020-02-04T11:54:34Z")
	m.TakenAtLocal, _ = time.Parse(time.RFC3339, "2020-02-04T11:54:34Z")
	m.TimeZone = m.GetTimeZone()

	if m.TimeZone != "Europe/Berlin" {
		t.Fatalf("time zone should be Europe/Berlin: %s", m.TimeZone)
	}

	localTime := m.TakenAtLocal.Format("2006-01-02T15:04:05")

	if localTime != "2020-02-04T11:54:34" {
		t.Fatalf("local time should be 2020-02-04T11:54:34: %s", localTime)
	}

	utcTime := m.GetTakenAt().Format("2006-01-02T15:04:05")

	if utcTime != "2020-02-04T10:54:34" {
		t.Fatalf("utc time should be 2020-02-04T10:54:34: %s", utcTime)
	}
}

func TestPhoto_CountryName(t *testing.T) {
	t.Run("Unknown", func(t *testing.T) {
		m := Photo{PhotoCountry: "xx"}
		assert.Equal(t, "Unknown", m.CountryName())
	})
	t.Run("Germany", func(t *testing.T) {
		m := Photo{PhotoCountry: "de"}
		assert.Equal(t, "Germany", m.CountryName())
	})
}

func TestUpdateLocation(t *testing.T) {
	t.Run("estimate", func(t *testing.T) {
		m := Photo{
			PhotoName:    "test_photo_2",
			PhotoCountry: UnknownID,
			PhotoLat:     0.0,
			PhotoLng:     0.0,
			PlaceID:      "mx:VvfNBpFegSCr",
			PlaceSrc:     SrcEstimate,
		}

		assert.Equal(t, "Unknown", m.CountryName())

		m.UpdateLocation()

		assert.Equal(t, "Mexico", m.CountryName())
		assert.Equal(t, "mx", m.PhotoCountry)
		assert.Equal(t, float32(0.0), m.PhotoLat)
		assert.Equal(t, float32(0.0), m.PhotoLng)
		assert.Equal(t, "mx:VvfNBpFegSCr", m.PlaceID)
		assert.Equal(t, SrcEstimate, m.PlaceSrc)
	})

	t.Run("change_estimate", func(t *testing.T) {
		m := Photo{
			PhotoName:    "test_photo_1",
			PhotoCountry: "de",
			PhotoLat:     0.0,
			PhotoLng:     0.0,
			PlaceID:      "de:HFqPHxa2Hsol",
			PlaceSrc:     SrcManual,
		}

		assert.Equal(t, "Germany", m.CountryName())

		m.UpdateLocation()

		assert.Equal(t, "Germany", m.CountryName())
		assert.Equal(t, "de", m.PhotoCountry)
		assert.Equal(t, float32(0.0), m.PhotoLat)
		assert.Equal(t, float32(0.0), m.PhotoLng)
		assert.Equal(t, UnknownID, m.PlaceID)
		assert.Equal(t, SrcManual, m.PlaceSrc)
	})
}
