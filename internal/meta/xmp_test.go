package meta

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestXMP(t *testing.T) {
	t.Run("AppleXmpTwo", func(t *testing.T) {
		data, err := XMP("testdata/apple-test-2.xmp")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Botanischer Garten", data.Title)
		assert.Equal(t, time.Date(2021, 3, 24, 13, 07, 29, 0, time.FixedZone("", +3600)).UTC(), data.TakenAt.UTC())
		assert.Equal(t, "Tulpen am See", data.Caption)
		assert.Equal(t, Keywords{"blume", "krokus", "schöne", "wiese"}, data.Keywords)
	})
	t.Run("Photoshop", func(t *testing.T) {
		data, err := XMP("testdata/photoshop.xmp")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Night Shift / Berlin / 2020", data.Title)
		assert.Equal(t, time.Date(2020, 1, 1, 17, 28, 25, 729626112, time.UTC), data.TakenAt)
		assert.Equal(t, "Michael Mayer", data.Artist)
		assert.Equal(t, "Example file for development", data.Caption)
		assert.Equal(t, "This is an (edited) legal notice", data.Copyright)
		assert.Equal(t, "HUAWEI", data.CameraMake)
		assert.Equal(t, "ELE-L29", data.CameraModel)
		assert.Equal(t, "HUAWEI P30 Rear Main Camera", data.LensModel)
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

		assert.Equal(t, time.Date(2022, 9, 4, 0, 48, 26, 0, time.UTC), data.TakenAt.UTC())
		assert.True(t, data.TakenAtLocal.IsZero())
		assert.Equal(t, "UTC", data.TimeZone)
	})
	t.Run("XmpFaceRegions", func(t *testing.T) {
		data, err := XMP("testdata/xmp-face-regions.xmp")

		require.NoError(t, err)
		assert.Equal(t, "Family Photo", data.Title)
		assert.Equal(t, "Canon", data.CameraMake)
		assert.Equal(t, "Canon EOS 5D Mark IV", data.CameraModel)

		// Expect exactly two Face regions (the Pet region should be excluded).
		require.Len(t, data.FaceRegions, 2, "expected 2 face regions (Pet type should be excluded)")

		// Region 1: John Doe
		// XMP center coords: x=0.4, y=0.3, w=0.2, h=0.4
		// Converted top-left: x=0.4-0.1=0.3, y=0.3-0.2=0.1
		r0 := data.FaceRegions[0]
		assert.Equal(t, "John Doe", r0.Name)
		assert.InDelta(t, 0.3, float64(r0.X), 0.001, "Region 0 X (top-left)")
		assert.InDelta(t, 0.1, float64(r0.Y), 0.001, "Region 0 Y (top-left)")
		assert.InDelta(t, 0.2, float64(r0.W), 0.001, "Region 0 W")
		assert.InDelta(t, 0.4, float64(r0.H), 0.001, "Region 0 H")

		// Region 2: Jane Smith
		// XMP center coords: x=0.75, y=0.35, w=0.15, h=0.3
		// Converted top-left: x=0.75-0.075=0.675, y=0.35-0.15=0.2
		r1 := data.FaceRegions[1]
		assert.Equal(t, "Jane Smith", r1.Name)
		assert.InDelta(t, 0.675, float64(r1.X), 0.001, "Region 1 X (top-left)")
		assert.InDelta(t, 0.2, float64(r1.Y), 0.001, "Region 1 Y (top-left)")
		assert.InDelta(t, 0.15, float64(r1.W), 0.001, "Region 1 W")
		assert.InDelta(t, 0.3, float64(r1.H), 0.001, "Region 1 H")
	})
	t.Run("XmpNoFaceRegions", func(t *testing.T) {
		// Verify that XMP files without face regions return an empty FaceRegions slice.
		data, err := XMP("testdata/photoshop.xmp")

		require.NoError(t, err)
		assert.Empty(t, data.FaceRegions, "photoshop.xmp should have no face regions")
	})
}
