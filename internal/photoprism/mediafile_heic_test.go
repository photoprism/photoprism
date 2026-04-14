package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/meta"
)

func TestMediaFile_Heic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	c := config.TestConfig()

	t.Run("IphoneSevenHeic", func(t *testing.T) {
		prevDisableHeifConvert := conf.Options().DisableHeifConvert
		prevDisableImageMagick := conf.Options().DisableImageMagick
		conf.Options().DisableHeifConvert = true
		conf.Options().DisableImageMagick = true
		t.Cleanup(func() {
			conf.Options().DisableHeifConvert = prevDisableHeifConvert
			conf.Options().DisableImageMagick = prevDisableImageMagick
		})

		img, err := NewMediaFile(filepath.Join(conf.SamplesPath(), "iphone_7.heic"))

		if err != nil {
			t.Fatal(err)
		}

		info := img.MetaData()

		assert.IsType(t, meta.Data{}, info)

		convert := NewConvert(conf)

		// Create JPEG image.
		if _, err = convert.ToImage(img, false); err != nil {
			t.Fatal(err)
		}

		// Replace JPEG image.
		jpeg, err := convert.ToImage(img, true)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("JPEG FILENAME: %s", jpeg.FileName())

		assert.Nil(t, err)

		jpegInfo := jpeg.MetaData()

		assert.IsType(t, meta.Data{}, jpegInfo)

		assert.Nil(t, err)

		assert.Equal(t, "", jpegInfo.DocumentID)
		assert.Equal(t, "2018-09-10 03:16:13.023 +0000 UTC", jpegInfo.TakenAt.String())
		assert.Equal(t, "2018-09-10 12:16:13.023 +0000 UTC", jpegInfo.TakenAtLocal.String())
		// The native libvips/libheif path does not apply EXIF orientation for HEIF files
		// because the HEIF spec treats it as informational only (see strukturag/libheif#227).
		// This iPhone 7 file lacks an irot box and carries EXIF orientation 6, which is
		// technically non-conformant.  The output retains the raw sensor dimensions.
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
		assert.Equal(t, "", jpegInfo.Caption)

		if err = os.Remove(filepath.Join(conf.SidecarPath(), conf.SamplesPath(), "iphone_7.heic.jpg")); err != nil {
			t.Error(err)
		}
	})
	t.Run("IphoneFifteenProHeic", func(t *testing.T) {
		prevDisableHeifConvert := c.Options().DisableHeifConvert
		prevDisableImageMagick := c.Options().DisableImageMagick
		c.Options().DisableHeifConvert = true
		c.Options().DisableImageMagick = true
		t.Cleanup(func() {
			c.Options().DisableHeifConvert = prevDisableHeifConvert
			c.Options().DisableImageMagick = prevDisableImageMagick
		})

		img, err := NewMediaFile(filepath.Join(c.SamplesPath(), "iphone_15_pro.heic"))

		if err != nil {
			t.Fatal(err)
		}

		info := img.MetaData()

		assert.IsType(t, meta.Data{}, info)

		convert := NewConvert(c)

		// Create JPEG image.
		if _, err = convert.ToImage(img, false); err != nil {
			t.Fatal(err)
		}

		// Replace JPEG image.
		jpeg, err := convert.ToImage(img, true)

		if err != nil {
			t.Fatal(err)
		}

		assert.Nil(t, err)

		jpegInfo := jpeg.MetaData()

		assert.IsType(t, meta.Data{}, jpegInfo)

		assert.Nil(t, err)

		assert.Equal(t, "", jpegInfo.DocumentID)
		assert.Equal(t, "2023-10-31 10:44:43.432 +0000 UTC", jpegInfo.TakenAt.String())
		assert.Equal(t, "2023-10-31 11:44:43.432 +0000 UTC", jpegInfo.TakenAtLocal.String())
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
		assert.Equal(t, "", jpegInfo.Caption)

		if err = os.Remove(filepath.Join(c.SidecarPath(), c.SamplesPath(), "iphone_15_pro.heic.jpg")); err != nil {
			t.Error(err)
		}
	})
	t.Run("FoxProfileAvif", func(t *testing.T) {
		prevDisableHeifConvert := c.Options().DisableHeifConvert
		prevDisableImageMagick := c.Options().DisableImageMagick
		c.Options().DisableHeifConvert = true
		c.Options().DisableImageMagick = true
		t.Cleanup(func() {
			c.Options().DisableHeifConvert = prevDisableHeifConvert
			c.Options().DisableImageMagick = prevDisableImageMagick
		})

		img, err := NewMediaFile(filepath.Join(c.SamplesPath(), "fox.profile0.8bpc.yuv420.avif"))

		if err != nil {
			t.Fatal(err)
		}

		convert := NewConvert(c)

		jpeg, err := convert.ToImage(img, true)

		if err != nil {
			t.Fatal(err)
		}

		assert.NotNil(t, jpeg)
		assert.True(t, jpeg.IsJpeg())
		assert.True(t, jpeg.Exists())
		assert.Greater(t, jpeg.Width(), 0)
		assert.Greater(t, jpeg.Height(), 0)

		if err = os.Remove(filepath.Join(c.SidecarPath(), c.SamplesPath(), "fox.profile0.8bpc.yuv420.avif.jpg")); err != nil {
			t.Error(err)
		}
	})
}
