package meta

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestData_ResolveTimeZone exercises each branch of (*Data).ResolveTimeZone in
// isolation so a regression in either the EXIF/ExifTool path or the XMP path
// surfaces as a focused unit-test failure rather than as a downstream entity
// mismatch.
func TestData_ResolveTimeZone(t *testing.T) {
	t.Run("CreatedAtOverridesTakenAt", func(t *testing.T) {
		// Mirrors the Media Create Date branch (resolver line ~215): a
		// non-zero CreatedAt wins over the parsed TakenAt for videos and
		// other containers where the "create" tag is authoritative.
		created := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		other := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
		data := Data{CreatedAt: created, TakenAt: other, TakenAtLocal: other}

		data.ResolveTimeZone("test.jpg")

		assert.Equal(t, created, data.TakenAt)
	})
	t.Run("GpsUtcFallback", func(t *testing.T) {
		// No TakenAt / TakenAtLocal, but GPSDateTime is present — the
		// resolver promotes the GPS UTC timestamp and clears any stale
		// local time.
		gps := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		data := Data{TakenGps: gps}

		data.ResolveTimeZone("test.jpg")

		assert.Equal(t, gps, data.TakenAt)
		assert.Equal(t, "UTC", data.TimeZone)
		assert.True(t, data.TakenAtLocal.IsZero() || data.TakenAtLocal.Equal(gps))
	})
	t.Run("PlausibilityCheckFork", func(t *testing.T) {
		// Local and UTC time differ by >27h → the resolver assumes the
		// local timestamp is bogus and forces it to mirror the UTC value.
		utc := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		local := time.Date(2024, 6, 17, 12, 0, 0, 0, time.UTC)
		data := Data{TakenAt: utc, TakenAtLocal: local}

		data.ResolveTimeZone("test.jpg")

		assert.Equal(t, data.TakenAt.UTC(), data.TakenAtLocal.UTC())
	})
	t.Run("Mp4DefaultsToUtc", func(t *testing.T) {
		// MP4 containers conventionally store UTC timestamps with no
		// explicit zone, so the resolver flips to UTC when MimeType
		// matches and no local time is set.
		taken := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		data := Data{TakenAt: taken, MimeType: MimeVideoMp4}

		data.ResolveTimeZone("test.mp4")

		assert.Equal(t, "UTC", data.TimeZone)
	})
	t.Run("QuicktimeDefaultsToUtc", func(t *testing.T) {
		taken := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		data := Data{TakenAt: taken, MimeType: MimeQuicktime}

		data.ResolveTimeZone("test.mov")

		assert.Equal(t, "UTC", data.TimeZone)
	})
	t.Run("GpsResolvesZoneAndLocal", func(t *testing.T) {
		// Berlin GPS + Berlin wall-clock. Resolver picks the IANA zone
		// from coordinates and derives the UTC instant from the local
		// time (CET is +01:00 in January).
		berlin := time.FixedZone("CET", 3600)
		local := time.Date(2024, 1, 15, 17, 28, 25, 0, berlin)
		data := Data{
			TakenAt:      local,
			TakenAtLocal: local,
			Lat:          52.5,
			Lng:          13.4,
		}

		data.ResolveTimeZone("test.jpg")

		assert.Equal(t, "Europe/Berlin", data.TimeZone)
		assert.Equal(t, time.Date(2024, 1, 15, 16, 28, 25, 0, time.UTC), data.TakenAt.UTC())
		assert.Equal(t, "2024-01-15 17:28:25", data.TakenAtLocal.Format("2006-01-02 15:04:05"))
	})
	t.Run("OffsetResolvesZone", func(t *testing.T) {
		// No GPS, no wall-clock zone — the resolver falls back to the
		// TimeOffset string ("+02:00") to derive a fixed-offset zone.
		taken := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		data := Data{
			TakenAt:      taken,
			TakenAtLocal: taken,
			TimeOffset:   "+02:00",
		}

		data.ResolveTimeZone("test.jpg")

		assert.Equal(t, "UTC+2", data.TimeZone)
	})
	t.Run("FallbackLocalFromUtc", func(t *testing.T) {
		// TakenAtLocal zero, TakenAt set, no GPS/offset — resolver fills
		// in TakenAtLocal so downstream consumers always see a value.
		taken := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		data := Data{TakenAt: taken}

		data.ResolveTimeZone("test.jpg")

		assert.False(t, data.TakenAtLocal.IsZero())
	})
	t.Run("NanosAppliedToBoth", func(t *testing.T) {
		// Sub-second precision lives in TakenNs (mirroring SubSecTimeOriginal).
		// The resolver applies it to both TakenAt and TakenAtLocal when the
		// truncated UTC time has no nanoseconds yet, preserving entity-layer
		// parity across the two timestamp columns.
		taken := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		data := Data{
			TakenAt:      taken,
			TakenAtLocal: taken,
			TakenNs:      123456789,
		}

		data.ResolveTimeZone("test.jpg")

		assert.Equal(t, 123456789, data.TakenAt.Nanosecond())
		assert.Equal(t, 123456789, data.TakenAtLocal.Nanosecond())
	})
	t.Run("NoopOnEmpty", func(t *testing.T) {
		// All time fields zero — the resolver must not panic and must
		// leave the receiver in a coherent (still empty) state.
		data := Data{}

		assert.NotPanics(t, func() { data.ResolveTimeZone("test.jpg") })

		assert.True(t, data.TakenAt.IsZero())
		assert.True(t, data.TakenAtLocal.IsZero())
	})
	t.Run("PreservesTakenAtIfNonZero", func(t *testing.T) {
		// CreatedAt empty, TakenAt populated — the resolver must not blank
		// out a non-zero TakenAt just because CreatedAt is missing.
		taken := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		data := Data{TakenAt: taken, TakenAtLocal: taken}

		data.ResolveTimeZone("test.jpg")

		assert.False(t, data.TakenAt.IsZero())
		assert.Equal(t, taken.UTC().Format("2006-01-02 15:04:05"), data.TakenAt.UTC().Format("2006-01-02 15:04:05"))
	})
}
