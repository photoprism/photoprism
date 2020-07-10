package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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
