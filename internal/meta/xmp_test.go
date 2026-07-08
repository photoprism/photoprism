package meta

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestXMP(t *testing.T) {
	t.Run("AppleXmpTwo", func(t *testing.T) {
		data, err := XMP("testdata/apple-test-2.xmp")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Botanischer Garten", data.Title)
		assert.Equal(t, time.Date(2021, 3, 24, 13, 07, 29, 0, time.FixedZone("", +3600)).UTC(), data.TakenAt.UTC())
		// GPS resolves to Europe/Berlin; March 24 2021 is still CET (DST starts March 28),
		// so wall-clock 13:07:29 +01:00 = 12:07:29 UTC.
		assert.Equal(t, "Europe/Berlin", data.TimeZone)
		assert.Equal(t, "2021-03-24 13:07:29", data.TakenAtLocal.Format("2006-01-02 15:04:05"))
		assert.Equal(t, "Tulpen am See", data.Caption)
		// dc:subject feeds the descriptive Subject field, not Keywords, and
		// multi-word entries keep their spaces ("Schöne Wiese" stays intact).
		assert.Equal(t, "Krokus, Blume, Schöne Wiese", data.Subject)
		assert.Empty(t, data.Keywords)
		// Apple GPS — pure-decimal value with separate *Ref.
		assert.InDelta(t, 52.525082, data.Lat, 1e-4)
		assert.InDelta(t, 13.369367, data.Lng, 1e-4)
	})
	t.Run("Photoshop", func(t *testing.T) {
		data, err := XMP("testdata/photoshop.xmp")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Night Shift / Berlin / 2020", data.Title)
		// GPS resolves to Europe/Berlin. Wall-clock from photoshop:DateCreated is
		// "2020-01-01T17:28:25.729626112" — the resolver re-parses it in Berlin
		// (CET / +01:00) and re-joins the sub-second fraction from
		// exif:SubSecTimeOriginal (899614 → 899614000ns), matching the EXIF flow
		// where SubSecTimeOriginal is authoritative for the nanosecond component.
		assert.Equal(t, time.Date(2020, 1, 1, 16, 28, 25, 899614000, time.UTC), data.TakenAt)
		assert.Equal(t, "Europe/Berlin", data.TimeZone)
		assert.Equal(t, "2020-01-01 17:28:25", data.TakenAtLocal.Format("2006-01-02 15:04:05"))
		assert.Equal(t, "Michael Mayer", data.Artist)
		assert.Equal(t, "Example file for development", data.Caption)
		assert.Equal(t, "This is an (edited) legal notice", data.Copyright)
		// dc:subject feeds the descriptive Subject field only (matching the
		// embedded/ExifTool path's data.Subject source), never Keywords.
		assert.Equal(t, "desk, coffee, computer", data.Subject)
		assert.Empty(t, data.Keywords)
		assert.Equal(t, "HUAWEI", data.CameraMake)
		assert.Equal(t, "ELE-L29", data.CameraModel)
		assert.Equal(t, "HUAWEI P30 Rear Main Camera", data.LensModel)
		// Adobe 2-component GPS form (degrees + decimal-minutes + cardinal).
		assert.InDelta(t, 52.459690, data.Lat, 1e-4)
		assert.InDelta(t, 13.321832, data.Lng, 1e-4)
	})
	t.Run("CanonEosSixD", func(t *testing.T) {
		data, err := XMP("testdata/canon_eos_6d.xmp")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "", data.Title)
		assert.Equal(t, "", data.Artist)
		assert.Equal(t, "", data.Caption)
		assert.Equal(t, "", data.Copyright)
		assert.Equal(t, "Canon", data.CameraMake)
		assert.Equal(t, "Canon EOS 6D", data.CameraModel)
		assert.Equal(t, "EF24-105mm f/4L IS USM", data.LensModel)
	})
	t.Run("IphoneSeven", func(t *testing.T) {
		data, err := XMP("testdata/iphone_7.xmp")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "iPhone 7 / September 2018", data.Title)
		assert.Equal(t, "", data.Artist)
		assert.Equal(t, "", data.Caption)
		assert.Equal(t, "", data.Copyright)
		assert.Equal(t, "Apple", data.CameraMake)
		assert.Equal(t, "iPhone 7", data.CameraModel)
		assert.Equal(t, "iPhone 7 back camera 3.99mm f/1.8", data.LensModel)
		assert.Equal(t, false, data.Favorite)
		// iPhone 7 sidecar uses Adobe 2-component GPS form.
		assert.InDelta(t, 34.797450, data.Lat, 1e-4)
		assert.InDelta(t, 134.764633, data.Lng, 1e-4)
	})
	t.Run("Fstop", func(t *testing.T) {
		data, err := XMP("testdata/fstop-favorite.xmp")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, true, data.Favorite)
	})
	t.Run("DateHeic", func(t *testing.T) {
		data, err := XMP("testdata/date.heic.xmp")

		if err != nil {
			t.Fatal(err)
		}

		// photoshop:DateCreated = "2022-09-03T17:48:26-07:00" → 00:48:26 UTC the
		// next day. GPS resolves to America/Los_Angeles, so TakenAtLocal carries
		// the Seattle-area wall-clock (17:48:26 on Sept 3).
		assert.Equal(t, time.Date(2022, 9, 4, 0, 48, 26, 0, time.UTC), data.TakenAt.UTC())
		assert.Equal(t, "America/Los_Angeles", data.TimeZone)
		assert.Equal(t, "2022-09-03 17:48:26", data.TakenAtLocal.Format("2006-01-02 15:04:05"))
		// Apple HEIC: pure-decimal GPS with W cardinal → negative Lng.
		assert.InDelta(t, 47.675403, data.Lat, 1e-4)
		assert.InDelta(t, -122.317392, data.Lng, 1e-4)
		assert.InDelta(t, 63.63, data.Altitude, 0.01)
	})
	t.Run("SyntheticTimeOffsetsSubsec", func(t *testing.T) {
		// No GPS — resolver derives the time zone from exif:OffsetTimeOriginal
		// ("+02:00") and applies exif:SubSecTimeOriginal (123456 → 123456000ns).
		data, err := XMP("testdata/xmp/synthetic/time-offsets-subsec.xmp")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "+02:00", data.TimeOffset)
		assert.Equal(t, "UTC+2", data.TimeZone)
		// photoshop:DateCreated already carries an inline .123456 fraction; the
		// resolver preserves it because TakenAt.Nanosecond() != 0 short-circuits
		// the TakenNs re-application.
		assert.Equal(t, 123456000, data.TakenAt.Nanosecond())
		// The +02:00 offset is applied to the offset-less capture timestamp, so
		// the wall-clock 15:42:18 local resolves to 13:42:18 UTC.
		assert.Equal(t, "2026-05-06 13:42:18", data.TakenAt.UTC().Format("2006-01-02 15:04:05"))
		assert.Equal(t, "2026-05-06 15:42:18", data.TakenAtLocal.Format("2006-01-02 15:04:05"))
	})
	t.Run("SyntheticDateCreatedPriority", func(t *testing.T) {
		// photoshop:DateCreated (capture time) must win over a later, disagreeing
		// xmp:CreateDate (file write). No GPS, so nothing re-derives TakenAt — a
		// regression guard for CreatedAt clobbering the higher-priority capture
		// instant in the shared resolver.
		data, err := XMP("testdata/xmp/synthetic/datecreated-priority.xmp")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "2026-01-15 10:00:00", data.TakenAt.Format("2006-01-02 15:04:05"))
		assert.Equal(t, "2026-01-15 10:00:00", data.TakenAtLocal.Format("2006-01-02 15:04:05"))
	})
	t.Run("SyntheticGpsTimeCombined", func(t *testing.T) {
		// GPS coordinates + combined GPS timestamp ("2026-05-06T15:42:18Z"), no
		// other capture timestamp. Resolver falls back to GPS UTC time first,
		// then promotes the IANA zone from coordinates.
		data, err := XMP("testdata/xmp/synthetic/gps-time-combined.xmp")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Europe/Berlin", data.TimeZone)
		assert.Equal(t, time.Date(2026, 5, 6, 15, 42, 18, 0, time.UTC), data.TakenAt.UTC())
		// May 6 is CEST in Berlin (+02:00); 15:42:18 UTC → 17:42:18 wall-clock.
		assert.Equal(t, "2026-05-06 17:42:18", data.TakenAtLocal.Format("2006-01-02 15:04:05"))
	})
	t.Run("SyntheticGpsTimeSplit", func(t *testing.T) {
		// Same payload as Combined but expressed as split GPSDateStamp/GPSTimeStamp.
		// Doc reader must reassemble them and the resolver must produce identical
		// entity state to the combined case.
		data, err := XMP("testdata/xmp/synthetic/gps-time-split.xmp")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Europe/Berlin", data.TimeZone)
		assert.Equal(t, time.Date(2026, 5, 6, 15, 42, 18, 0, time.UTC), data.TakenAt.UTC())
		assert.Equal(t, "2026-05-06 17:42:18", data.TakenAtLocal.Format("2006-01-02 15:04:05"))
	})
	t.Run("SyntheticPanoramaKeyword", func(t *testing.T) {
		// GPano:ProjectionType=equirectangular must auto-add the "panorama"
		// keyword for parity with the EXIF/ExifTool paths
		// (exif.go:327, json_exiftool.go:282-284).
		data, err := XMP("testdata/xmp/synthetic/gpano-360.xmp")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "equirectangular", data.Projection)
		assert.Contains(t, data.Keywords.String(), "panorama")
	})
	t.Run("SyntheticAutoKeywordsFromCaption", func(t *testing.T) {
		// dc:description containing "HDR" must trigger AutoAddKeywords —
		// the "hdr" keyword is added and data.ImageType is set to
		// ImageTypeHDR for parity with json_exiftool.go:287.
		body := `<?xml version="1.0" encoding="UTF-8"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/">
   <dc:description><rdf:Alt><rdf:li xml:lang="x-default">HDR sunset</rdf:li></rdf:Alt></dc:description>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`
		tmp := filepath.Join(t.TempDir(), "caption-hdr.xmp")
		if err := os.WriteFile(tmp, []byte(body), fs.ModeFile); err != nil {
			t.Fatal(err)
		}

		data, err := XMP(tmp)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "HDR sunset", data.Caption)
		assert.Contains(t, data.Keywords.String(), "hdr")
		assert.Equal(t, ImageTypeHDR, data.ImageType)
	})
}

func TestApplyTimeOffset(t *testing.T) {
	base := time.Date(2026, 5, 6, 15, 42, 18, 123456000, time.UTC)

	t.Run("PositiveOffset", func(t *testing.T) {
		got := applyTimeOffset(base, "+02:00")
		assert.Equal(t, "2026-05-06 15:42:18", got.Format("2006-01-02 15:04:05"))
		assert.Equal(t, "2026-05-06 13:42:18", got.UTC().Format("2006-01-02 15:04:05"))
		assert.Equal(t, 123456000, got.Nanosecond())
	})

	t.Run("NegativeOffset", func(t *testing.T) {
		got := applyTimeOffset(base, "-05:30")
		assert.Equal(t, "2026-05-06 15:42:18", got.Format("2006-01-02 15:04:05"))
		assert.Equal(t, "2026-05-06 21:12:18", got.UTC().Format("2006-01-02 15:04:05"))
	})

	t.Run("InvalidOffsetReturnsInput", func(t *testing.T) {
		got := applyTimeOffset(base, "not-an-offset")
		assert.Equal(t, base, got)
	})
}
