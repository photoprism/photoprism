package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/meta"
)

func TestMediaFile_HEIC(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	c := config.TestConfig()

	t.Run("iphone_7.heic", func(t *testing.T) {
		img, err := NewMediaFile(filepath.Join(conf.ExamplesPath(), "iphone_7.heic"))

		if err != nil {
			t.Fatal(err)
		}

		info := img.MetaData()

		assert.IsType(t, meta.Data{}, info)

		convert := NewConvert(conf)

		// Create JPEG image.
		jpeg, err := convert.ToImage(img, false)

		if err != nil {
			t.Fatal(err)
		}

		// Replace JPEG image.
		jpeg, err = convert.ToImage(img, true)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("JPEG FILENAME: %s", jpeg.FileName())

		assert.Nil(t, err)

		jpegInfo := jpeg.MetaData()

		assert.IsType(t, meta.Data{}, jpegInfo)

		assert.Nil(t, err)

		assert.Equal(t, "", jpegInfo.DocumentID)
		assert.Equal(t, "2018-09-10 03:16:13 +0000 UTC", jpegInfo.TakenAt.String())
		assert.Equal(t, "2018-09-10 12:16:13 +0000 UTC", jpegInfo.TakenAtLocal.String())
		// KNOWN ISSUE: Orientation 6 would be correct instead (or the image should already be rotated),
		// see https://github.com/strukturag/libheif/issues/227#issuecomment-1532842570
		assert.Equal(t, 1, jpegInfo.Orientation)
		assert.Equal(t, "iPhone 7", jpegInfo.CameraModel)
		assert.Equal(t, "Apple", jpegInfo.CameraMake)
		assert.Equal(t, "iPhone 7 back camera 3.99mm f/1.8", jpegInfo.LensModel)
		assert.Equal(t, "Apple", jpegInfo.LensMake)
		assert.Equal(t, "Asia/Tokyo", jpegInfo.TimeZone)
		assert.Equal(t, "", jpegInfo.Artist)
		assert.Equal(t, 74, jpegInfo.FocalLength)
		assert.Equal(t, "1/4000", jpegInfo.Exposure)
		assert.Equal(t, float32(1.696), jpegInfo.Aperture)
		assert.Equal(t, 20, jpegInfo.Iso)
		assert.Equal(t, float32(34.79745), float32(jpegInfo.Lat))
		assert.Equal(t, float32(134.76463), float32(jpegInfo.Lng))
		assert.Equal(t, 0.0, jpegInfo.Altitude)
		assert.Equal(t, 4032, jpegInfo.Width)
		assert.Equal(t, 3024, jpegInfo.Height)
		assert.Equal(t, false, jpegInfo.Flash)
		assert.Equal(t, "", jpegInfo.Description)

		if err = os.Remove(filepath.Join(conf.SidecarPath(), conf.ExamplesPath(), "iphone_7.heic.jpg")); err != nil {
			t.Error(err)
		}
	})
	t.Run("iphone_15_pro.heic", func(t *testing.T) {
		img, err := NewMediaFile(filepath.Join(c.ExamplesPath(), "iphone_15_pro.heic"))

		if err != nil {
			t.Fatal(err)
		}

		info := img.MetaData()

		assert.IsType(t, meta.Data{}, info)

		convert := NewConvert(c)

		// Create JPEG image.
		jpeg, err := convert.ToImage(img, false)

		if err != nil {
			t.Fatal(err)
		}

		// Replace JPEG image.
		jpeg, err = convert.ToImage(img, true)

		if err != nil {
			t.Fatal(err)
		}

		assert.Nil(t, err)

		jpegInfo := jpeg.MetaData()

		assert.IsType(t, meta.Data{}, jpegInfo)

		assert.Nil(t, err)

		assert.Equal(t, "", jpegInfo.DocumentID)
		assert.Equal(t, "2023-10-31 10:44:43 +0000 UTC", jpegInfo.TakenAt.String())
		assert.Equal(t, "2023-10-31 11:44:43 +0000 UTC", jpegInfo.TakenAtLocal.String())
		assert.Equal(t, 1, jpegInfo.Orientation)
		assert.Equal(t, "iPhone 15 Pro", jpegInfo.CameraModel)
		assert.Equal(t, "Apple", jpegInfo.CameraMake)
		assert.Equal(t, "iPhone 15 Pro back triple camera 2.22mm f/2.2", jpegInfo.LensModel)
		assert.Equal(t, "Apple", jpegInfo.LensMake)
		assert.Equal(t, "Europe/Berlin", jpegInfo.TimeZone)
		assert.Equal(t, "", jpegInfo.Artist)
		assert.Equal(t, 14, jpegInfo.FocalLength)
		assert.Equal(t, "1/60", jpegInfo.Exposure)
		assert.Equal(t, float32(2.275), jpegInfo.Aperture)
		assert.Equal(t, 400, jpegInfo.Iso)
		assert.InEpsilon(t, 52.459605, jpegInfo.Lat, 0.0001)
		assert.InEpsilon(t, 13.3218416, jpegInfo.Lng, 0.0001)
		assert.Equal(t, 50.0, jpegInfo.Altitude)
		assert.Equal(t, 3024, jpegInfo.Width)
		assert.Equal(t, 4032, jpegInfo.Height)
		assert.Equal(t, false, jpegInfo.Flash)
		assert.Equal(t, "", jpegInfo.Description)

		if err = os.Remove(filepath.Join(c.SidecarPath(), c.ExamplesPath(), "iphone_15_pro.heic.jpg")); err != nil {
			t.Error(err)
		}
	})
}
