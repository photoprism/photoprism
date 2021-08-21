package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
