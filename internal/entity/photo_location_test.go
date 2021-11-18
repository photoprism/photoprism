package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
			PlaceID:      "s2:85d1ea7d3278",
			PlaceSrc:     SrcEstimate,
		}

		assert.Equal(t, "Unknown", m.CountryName())

		m.UpdateLocation()

		assert.Equal(t, "Mexico", m.CountryName())
		assert.Equal(t, "mx", m.PhotoCountry)
		assert.Equal(t, float32(0.0), m.PhotoLat)
		assert.Equal(t, float32(0.0), m.PhotoLng)
		assert.Equal(t, "s2:85d1ea7d3278", m.PlaceID)
		assert.Equal(t, SrcEstimate, m.PlaceSrc)
	})

	t.Run("change_estimate", func(t *testing.T) {
		m := Photo{
			PhotoName:    "test_photo_1",
			PhotoCountry: "de",
			PhotoLat:     0.0,
			PhotoLng:     0.0,
			PlaceID:      "s2:85d1ea7d3278",
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
