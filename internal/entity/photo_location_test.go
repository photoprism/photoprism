package entity

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/geo"
)

func TestPhoto_SetPosition(t *testing.T) {
	t.Run("SrcAuto", func(t *testing.T) {
		p := Photo{ID: 1, Place: nil, PlaceID: "", CellID: "s2:479a03fda123", PhotoLat: 0, PhotoLng: 0, PlaceSrc: SrcAuto}
		pos := geo.Position{Lat: 1, Lng: -1, Estimate: true}
		assert.Nil(t, p.Place)
		assert.Equal(t, "", p.PlaceID)
		assert.Equal(t, "s2:479a03fda123", p.CellID)
		assert.Equal(t, 0, int(p.PhotoLat))
		assert.Equal(t, 0, int(p.PhotoLng))
		p.SetPosition(pos, SrcEstimate, false)
		assert.Equal(t, "North Atlantic Ocean", p.Place.Label())
		assert.Equal(t, "zz:NjeJTM9IXJSv", p.PlaceID)
		assert.True(t, strings.HasPrefix(p.CellID, "s2:0ffebb"))
		assert.InEpsilon(t, 1, p.PhotoLat, 0.01)
		assert.InEpsilon(t, -1, p.PhotoLng, 0.01)
	})
}

func TestPhoto_AdoptPlace(t *testing.T) {
	place := PlaceFixtures.Get("mexico")
	t.Run("SrcAuto", func(t *testing.T) {
		p := Photo{ID: 1, Place: nil, PlaceID: "", CellID: "s2:479a03fda123", PhotoLat: -1, PhotoLng: 1, PlaceSrc: SrcAuto}
		o := &Photo{ID: 1, Place: &place, PlaceID: place.ID, CellID: "s2:479a03fda18c", PhotoLat: 15, PhotoLng: -11, PlaceSrc: SrcManual}
		assert.Nil(t, p.Place)
		assert.Equal(t, "", p.PlaceID)
		assert.Equal(t, "s2:479a03fda123", p.CellID)
		assert.Equal(t, -1, int(p.PhotoLat))
		assert.Equal(t, 1, int(p.PhotoLng))
		p.AdoptPlace(o, SrcEstimate, false)
		assert.Equal(t, &place, p.Place)
		assert.Equal(t, place.ID, p.PlaceID)
		assert.Equal(t, "zz", p.CellID)
		assert.Equal(t, 0, int(p.PhotoLat))
		assert.Equal(t, 0, int(p.PhotoLng))
	})
	t.Run("SrcManual", func(t *testing.T) {
		p := Photo{ID: 1, Place: nil, PlaceID: "", CellID: "s2:479a03fda123", PhotoLat: 0, PhotoLng: 0, PlaceSrc: SrcManual}
		o := &Photo{ID: 1, Place: &place, PlaceID: place.ID, CellID: "s2:479a03fda18c", PhotoLat: 1, PhotoLng: -1, PlaceSrc: SrcManual}
		assert.Nil(t, p.Place)
		assert.Equal(t, "", p.PlaceID)
		assert.Equal(t, "s2:479a03fda123", p.CellID)
		assert.Equal(t, 0, int(p.PhotoLat))
		assert.Equal(t, 0, int(p.PhotoLng))
		p.AdoptPlace(o, SrcEstimate, false)
		assert.Nil(t, p.Place)
		assert.Equal(t, "", p.PlaceID)
		assert.Equal(t, "s2:479a03fda123", p.CellID)
		assert.Equal(t, 0, int(p.PhotoLat))
		assert.Equal(t, 0, int(p.PhotoLng))
	})
	t.Run("Force", func(t *testing.T) {
		p := Photo{ID: 1, Place: nil, PlaceID: "", CellID: "s2:479a03fda123", PhotoLat: 1, PhotoLng: -1, PlaceSrc: SrcManual}
		o := &Photo{ID: 1, Place: &place, PlaceID: place.ID, CellID: "s2:479a03fda18c", PhotoLat: 0, PhotoLng: 0, PlaceSrc: SrcManual}
		assert.Nil(t, p.Place)
		assert.Equal(t, "", p.PlaceID)
		assert.Equal(t, "s2:479a03fda123", p.CellID)
		assert.Equal(t, 1, int(p.PhotoLat))
		assert.Equal(t, -1, int(p.PhotoLng))
		p.AdoptPlace(o, SrcEstimate, true)
		assert.Equal(t, &place, p.Place)
		assert.Equal(t, place.ID, p.PlaceID)
		assert.Equal(t, "zz", p.CellID)
		assert.Equal(t, 0, int(p.PhotoLat))
		assert.Equal(t, 0, int(p.PhotoLng))
	})
}

func TestPhoto_RemoveLocation(t *testing.T) {
	t.Run("SrcAuto", func(t *testing.T) {
		m := Photo{ID: 1, PlaceID: "zz:NjeJTM9IXJSv", CellID: "s2:479a03fda18c", PhotoLat: 1, PhotoLng: -1, PlaceSrc: SrcAuto}
		assert.NotEmpty(t, m.CellID)
		m.RemoveLocation(SrcEstimate, false)
		assert.Equal(t, "zz", m.CellID)
		assert.Equal(t, "zz", m.PlaceID)
		assert.Empty(t, m.PhotoLat)
		assert.Empty(t, m.PhotoLng)
		assert.Empty(t, m.PlaceSrc)
	})
	t.Run("SrcMeta", func(t *testing.T) {
		m := Photo{ID: 1, PlaceID: "zz:NjeJTM9IXJSv", CellID: "s2:479a03fda18c", PhotoLat: 1, PhotoLng: -1, PlaceSrc: SrcMeta}
		assert.NotEmpty(t, m.CellID)
		m.RemoveLocation(SrcEstimate, false)
		assert.Equal(t, "s2:479a03fda18c", m.CellID)
		assert.Equal(t, "zz:NjeJTM9IXJSv", m.PlaceID)
		assert.NotEmpty(t, m.PhotoLat)
		assert.NotEmpty(t, m.PhotoLng)
		assert.NotEmpty(t, m.PlaceSrc)
	})
}

func TestPhoto_SetAltitude(t *testing.T) {
	t.Run("ViaSetCoordinates", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(1.234), float32(m.PhotoLat))
		assert.Equal(t, float32(4.321), float32(m.PhotoLng))
		assert.Equal(t, 3, m.PhotoAltitude)

		m.SetCoordinates(0, 0, 5, SrcManual)
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(1.234), float32(m.PhotoLat))
		assert.Equal(t, float32(4.321), float32(m.PhotoLng))
		assert.Equal(t, 5, m.PhotoAltitude)
	})
	t.Run("Update", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(1.234), float32(m.PhotoLat))
		assert.Equal(t, float32(4.321), float32(m.PhotoLng))
		assert.Equal(t, 3, m.PhotoAltitude)

		m.SetAltitude(5, SrcManual)
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(1.234), float32(m.PhotoLat))
		assert.Equal(t, float32(4.321), float32(m.PhotoLng))
		assert.Equal(t, 5, m.PhotoAltitude)
	})
	t.Run("SkipUpdate", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(1.234), float32(m.PhotoLat))
		assert.Equal(t, float32(4.321), float32(m.PhotoLng))
		assert.Equal(t, 3, m.PhotoAltitude)

		m.SetAltitude(5, SrcEstimate)
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(1.234), float32(m.PhotoLat))
		assert.Equal(t, float32(4.321), float32(m.PhotoLng))
		assert.Equal(t, 3, m.PhotoAltitude)
	})
	t.Run("UpdateEmptyAltitude", func(t *testing.T) {
		m := Photo{ID: 1, PlaceSrc: SrcMeta, PhotoLat: float64(1.234), PhotoLng: float64(4.321), PhotoAltitude: 0}

		m.SetAltitude(-5, SrcAuto)
		assert.Equal(t, 0, m.PhotoAltitude)

		m.SetAltitude(-5, SrcEstimate)
		assert.Equal(t, 0, m.PhotoAltitude)

		m.SetAltitude(-5, SrcMeta)
		assert.Equal(t, -5, m.PhotoAltitude)
	})
	t.Run("ZeroAltitudeManual", func(t *testing.T) {
		m := Photo{ID: 1, PlaceSrc: SrcManual, PhotoLat: 1.234, PhotoLng: 4.321, PhotoAltitude: 5}

		m.SetAltitude(0, SrcManual)
		assert.Equal(t, 0, m.PhotoAltitude)
	})
}

func TestPhoto_SetCoordinates(t *testing.T) {
	t.Run("empty coordinates", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(1.234), float32(m.PhotoLat))
		assert.Equal(t, float32(4.321), float32(m.PhotoLng))
		assert.Equal(t, 3, m.PhotoAltitude)

		m.SetCoordinates(0, 0, 5, SrcManual)
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(1.234), float32(m.PhotoLat))
		assert.Equal(t, float32(4.321), float32(m.PhotoLng))
		assert.Equal(t, 5, m.PhotoAltitude)
	})
	t.Run("same source new values", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(1.234), float32(m.PhotoLat))
		assert.Equal(t, float32(4.321), float32(m.PhotoLng))
		assert.Equal(t, 3, m.PhotoAltitude)

		m.SetCoordinates(5.555, 5.555, 5, SrcMeta)
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(5.555), float32(m.PhotoLat))
		assert.Equal(t, float32(5.555), float32(m.PhotoLng))
		assert.Equal(t, 5, m.PhotoAltitude)
	})
	t.Run("different source lower priority", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(1.234), float32(m.PhotoLat))
		assert.Equal(t, float32(4.321), float32(m.PhotoLng))
		assert.Equal(t, 3, m.PhotoAltitude)

		m.SetCoordinates(5.555, 5.555, 5, SrcName)
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(1.234), float32(m.PhotoLat))
		assert.Equal(t, float32(4.321), float32(m.PhotoLng))
		assert.Equal(t, 3, m.PhotoAltitude)
	})
	t.Run("different source equal priority", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(1.234), float32(m.PhotoLat))
		assert.Equal(t, float32(4.321), float32(m.PhotoLng))
		assert.Equal(t, 3, m.PhotoAltitude)

		m.SetCoordinates(5.555, 5.555, 5, SrcKeyword)
		assert.Equal(t, float32(5.555), float32(m.PhotoLat))
		assert.Equal(t, float32(5.555), float32(m.PhotoLng))
		assert.Equal(t, 5, m.PhotoAltitude)
	})
	t.Run("different source higher priority", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo21")
		assert.Equal(t, SrcEstimate, m.PlaceSrc)
		assert.Equal(t, 0.0, m.PhotoLat)
		assert.Equal(t, 0.0, m.PhotoLng)
		assert.Equal(t, 0, m.PhotoAltitude)

		m.SetCoordinates(5.555, 5.555, 5, SrcMeta)
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(5.555), float32(m.PhotoLat))
		assert.Equal(t, float32(5.555), float32(m.PhotoLng))
		assert.Equal(t, 5, m.PhotoAltitude)
	})
	t.Run("different source highest priority (manual)", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, SrcMeta, m.PlaceSrc)
		assert.Equal(t, float32(1.234), float32(m.PhotoLat))
		assert.Equal(t, float32(4.321), float32(m.PhotoLng))
		assert.Equal(t, 3, m.PhotoAltitude)

		m.SetCoordinates(5.555, 5.555, 5, SrcManual)
		assert.Equal(t, SrcManual, m.PlaceSrc)
		assert.Equal(t, float32(5.555), float32(m.PhotoLat))
		assert.Equal(t, float32(5.555), float32(m.PhotoLng))
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
	t.Run("19800101_000002_D640C559", func(t *testing.T) {
		m := PhotoFixtures.Get("19800101_000002_D640C559")
		assert.False(t, m.HasLocation())
	})
	t.Run("Photo08", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo08")
		assert.True(t, m.HasLocation())
	})
}

func TestPhoto_HasLatLng(t *testing.T) {
	t.Run("Photo01", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		assert.True(t, m.HasLatLng())
	})
	t.Run("Photo09", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo09")
		assert.True(t, m.HasLatLng())
		m.PhotoLat = 0
		m.PhotoLng = 0
		assert.False(t, m.HasLatLng())
	})
}

func TestPhoto_NoLatLng(t *testing.T) {
	t.Run("Photo01", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		assert.False(t, m.NoLatLng())
	})
	t.Run("Photo09", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo09")
		assert.False(t, m.NoLatLng())
		m.PhotoLat = 0
		m.PhotoLng = 0
		assert.True(t, m.NoLatLng())
	})
}

func TestPhoto_NoPlace(t *testing.T) {
	t.Run("19800101_000002_D640C559", func(t *testing.T) {
		m := PhotoFixtures.Get("19800101_000002_D640C559")
		assert.True(t, m.UnknownPlace())
	})
	t.Run("Photo08", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo08")
		assert.False(t, m.UnknownPlace())
	})
}

func TestPhoto_HasPlace(t *testing.T) {
	t.Run("19800101_000002_D640C559", func(t *testing.T) {
		m := PhotoFixtures.Get("19800101_000002_D640C559")
		assert.False(t, m.HasPlace())
	})
	t.Run("Photo08", func(t *testing.T) {
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
		t.Fatalf("UTC time should be 2020-02-04T10:54:34: %s", utcTime)
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
		assert.Equal(t, 0.0, m.PhotoLat)
		assert.Equal(t, 0.0, m.PhotoLng)
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
		assert.Equal(t, 0.0, m.PhotoLat)
		assert.Equal(t, 0.0, m.PhotoLng)
		assert.Equal(t, UnknownID, m.PlaceID)
		assert.Equal(t, SrcManual, m.PlaceSrc)
	})
}
