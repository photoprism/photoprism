package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/media"
)

func TestJSON_Motion(t *testing.T) {
	t.Run("GooglePixel2_JPG", func(t *testing.T) {
		data, err := JSON("testdata/motion/google_pixel2.jpg.json", "")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %#v", data)

		assert.Equal(t, "pixel2.jpg", data.FileName)
		assert.Equal(t, media.Live, data.MediaType)
		assert.Equal(t, true, data.EmbeddedThumb)
		assert.Equal(t, true, data.EmbeddedVideo)
		assert.Equal(t, CodecJpeg, data.Codec)
		assert.Equal(t, int64(0), data.Duration.Milliseconds())
		assert.Equal(t, "0s", data.Duration.String())
		assert.Equal(t, "2018-03-18 19:21:15 +0000 UTC", data.TakenAtLocal.String())
		assert.Equal(t, "2018-03-18 23:21:15 +0000 UTC", data.TakenAt.String())
		assert.Equal(t, 796940000, data.TakenNs)
		assert.Equal(t, "America/New_York", data.TimeZone)
		assert.Equal(t, 3024, data.Width)
		assert.Equal(t, 4032, data.Height)
		assert.Equal(t, 3024, data.ActualWidth())
		assert.Equal(t, 4032, data.ActualHeight())
		assert.Equal(t, 1, data.Orientation)
		assert.Equal(t, float32(35.42307), data.Lat)
		assert.Equal(t, float32(-78.65212), data.Lng)
		assert.Equal(t, "Google", data.CameraMake)
		assert.Equal(t, "Pixel 2", data.CameraModel)
		assert.Equal(t, "", data.LensModel)
	})
	t.Run("GooglePixel4a_JPG", func(t *testing.T) {
		data, err := JSON("testdata/motion/google_pixel4a.jpg.json", "")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %#v", data)

		assert.Equal(t, "pixel4a.jpg", data.FileName)
		assert.Equal(t, media.Live, data.MediaType)
		assert.Equal(t, false, data.EmbeddedThumb)
		assert.Equal(t, true, data.EmbeddedVideo)
		assert.Equal(t, CodecJpeg, data.Codec)
		assert.Equal(t, int64(0), data.Duration.Milliseconds())
		assert.Equal(t, "0s", data.Duration.String())
		assert.Equal(t, "2021-09-17 19:31:36 +0000 UTC", data.TakenAtLocal.String())
		assert.Equal(t, "2021-09-17 23:31:36 +0000 UTC", data.TakenAt.String())
		assert.Equal(t, 844000000, data.TakenNs)
		assert.Equal(t, "America/New_York", data.TimeZone)
		assert.Equal(t, 4032, data.Width)
		assert.Equal(t, 3024, data.Height)
		assert.Equal(t, 3024, data.ActualWidth())
		assert.Equal(t, 4032, data.ActualHeight())
		assert.Equal(t, 6, data.Orientation)
		assert.Equal(t, float32(35.778152), data.Lat)
		assert.Equal(t, float32(-78.63687), data.Lng)
		assert.Equal(t, "Google", data.CameraMake)
		assert.Equal(t, "Pixel 4a", data.CameraModel)
		assert.Equal(t, "", data.LensModel)
	})
	t.Run("GooglePixel6_JPG", func(t *testing.T) {
		data, err := JSON("testdata/motion/google_pixel6.jpg.json", "")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %#v", data)

		assert.Equal(t, "PXL_20211227_151322429.MP.jpg", data.FileName)
		assert.Equal(t, media.Live, data.MediaType)
		assert.Equal(t, true, data.EmbeddedThumb)
		assert.Equal(t, true, data.EmbeddedVideo)
		assert.Equal(t, CodecJpeg, data.Codec)
		assert.Equal(t, int64(0), data.Duration.Milliseconds())
		assert.Equal(t, "0s", data.Duration.String())
		assert.Equal(t, "2021-12-27 16:13:22 +0000 UTC", data.TakenAtLocal.String())
		assert.Equal(t, "2021-12-27 15:13:22 +0000 UTC", data.TakenAt.String())
		assert.Equal(t, 429000000, data.TakenNs)
		assert.Equal(t, "Europe/Berlin", data.TimeZone)
		assert.Equal(t, 4032, data.Width)
		assert.Equal(t, 2268, data.Height)
		assert.Equal(t, 4032, data.ActualWidth())
		assert.Equal(t, 2268, data.ActualHeight())
		assert.Equal(t, 1, data.Orientation)
		assert.Equal(t, float32(48.610027), data.Lat)
		assert.Equal(t, float32(8.861558), data.Lng)
		assert.Equal(t, "Google", data.CameraMake)
		assert.Equal(t, "Pixel 6", data.CameraModel)
		assert.Equal(t, "Pixel 6 back camera 6.81mm f/1.85", data.LensModel)
	})
	t.Run("GooglePixel7Pro_JPG", func(t *testing.T) {
		data, err := JSON("testdata/motion/google_pixel7pro.jpg.json", "")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %#v", data)

		assert.Equal(t, "PXL_20230329_144843201.MP.jpg", data.FileName)
		assert.Equal(t, media.Live, data.MediaType)
		assert.Equal(t, true, data.EmbeddedThumb)
		assert.Equal(t, true, data.EmbeddedVideo)
		assert.Equal(t, CodecJpeg, data.Codec)
		assert.Equal(t, int64(0), data.Duration.Milliseconds())
		assert.Equal(t, "0s", data.Duration.String())
		assert.Equal(t, "2023-03-29 15:48:43 +0000 UTC", data.TakenAtLocal.String())
		assert.Equal(t, "2023-03-29 14:48:43 +0000 UTC", data.TakenAt.String())
		assert.Equal(t, 201000000, data.TakenNs)
		assert.Equal(t, "Europe/London", data.TimeZone)
		assert.Equal(t, 4080, data.Width)
		assert.Equal(t, 3072, data.Height)
		assert.Equal(t, 4080, data.ActualWidth())
		assert.Equal(t, 3072, data.ActualHeight())
		assert.Equal(t, 1, data.Orientation)
		assert.Equal(t, float32(52.70636), data.Lat)
		assert.Equal(t, float32(-2.7605944), data.Lng)
		assert.Equal(t, "Google", data.CameraMake)
		assert.Equal(t, "Pixel 7 Pro", data.CameraModel)
		assert.Equal(t, "Pixel 7 Pro back camera 19.0mm f/3.5", data.LensModel)
	})
	t.Run("SamsungGalaxyS20_JPG", func(t *testing.T) {
		data, err := JSON("testdata/motion/samsung_galaxys20.jpg.json", "")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %#v", data)

		assert.Equal(t, "20230822_143803.jpg", data.FileName)
		assert.Equal(t, media.Live, data.MediaType)
		assert.Equal(t, true, data.EmbeddedThumb)
		assert.Equal(t, true, data.EmbeddedVideo)
		assert.Equal(t, CodecJpeg, data.Codec)
		assert.Equal(t, int64(0), data.Duration.Milliseconds())
		assert.Equal(t, "0s", data.Duration.String())
		assert.Equal(t, "2023-08-22 14:38:03 +0000 UTC", data.TakenAtLocal.String())
		assert.Equal(t, "2023-08-22 14:38:03 +0000 UTC", data.TakenAt.String())
		assert.Equal(t, 583000000, data.TakenNs)
		assert.Equal(t, "", data.TimeZone)
		assert.Equal(t, 4032, data.Width)
		assert.Equal(t, 3024, data.Height)
		assert.Equal(t, 3024, data.ActualWidth())
		assert.Equal(t, 4032, data.ActualHeight())
		assert.Equal(t, 6, data.Orientation)
		assert.Equal(t, float32(0), data.Lat)
		assert.Equal(t, float32(0), data.Lng)
		assert.Equal(t, "samsung", data.CameraMake)
		assert.Equal(t, "SM-G780F", data.CameraModel)
		assert.Equal(t, "", data.LensModel)
	})
	t.Run("SamsungGalaxyS20_MP4", func(t *testing.T) {
		data, err := JSON("testdata/motion/samsung_galaxys20.mp4.json", "")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %#v", data)

		assert.Equal(t, "20230822_143803.mp4", data.FileName)
		assert.Equal(t, media.Unknown, data.MediaType)
		assert.Equal(t, false, data.EmbeddedThumb)
		assert.Equal(t, false, data.EmbeddedVideo)
		assert.Equal(t, CodecAvc1, data.Codec)
		assert.Equal(t, int64(2990), data.Duration.Milliseconds())
		assert.Equal(t, "2.99s", data.Duration.String())
		assert.Equal(t, "2023-08-22 14:38:06 +0000 UTC", data.TakenAtLocal.String())
		assert.Equal(t, "2023-08-22 11:38:06 +0000 UTC", data.TakenAt.String())
		assert.Equal(t, 0, data.TakenNs)
		assert.Equal(t, "Europe/Kiev", data.TimeZone)
		assert.Equal(t, 1440, data.Width)
		assert.Equal(t, 1080, data.Height)
		assert.Equal(t, 1080, data.ActualWidth())
		assert.Equal(t, 1440, data.ActualHeight())
		assert.Equal(t, 6, data.Orientation)
		assert.Equal(t, float32(48.4565), data.Lat)
		assert.Equal(t, float32(35.072), data.Lng)
		assert.Equal(t, "", data.CameraMake)
		assert.Equal(t, "", data.CameraModel)
		assert.Equal(t, "", data.LensModel)
	})
	t.Run("SamsungGalaxyS20FE_HEIF", func(t *testing.T) {
		data, err := JSON("testdata/motion/samsung_galaxys20fe.heif.json", "")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %#v", data)

		assert.Equal(t, "20220423_085935.heif", data.FileName)
		assert.Equal(t, media.Live, data.MediaType)
		assert.Equal(t, false, data.EmbeddedThumb)
		assert.Equal(t, true, data.EmbeddedVideo)
		assert.Equal(t, CodecHeic, data.Codec)
		assert.Equal(t, int64(0), data.Duration.Milliseconds())
		assert.Equal(t, "0s", data.Duration.String())
		assert.Equal(t, "2022-04-23 08:59:35 +0000 UTC", data.TakenAtLocal.String())
		assert.Equal(t, "2022-04-23 06:59:35 +0000 UTC", data.TakenAt.String())
		assert.Equal(t, 0, data.TakenNs)
		assert.Equal(t, "Europe/Berlin", data.TimeZone)
		assert.Equal(t, 4032, data.Width)
		assert.Equal(t, 3024, data.Height)
		assert.Equal(t, 3024, data.ActualWidth())
		assert.Equal(t, 4032, data.ActualHeight())
		assert.Equal(t, 6, data.Orientation)
		assert.Equal(t, float32(51.433468), data.Lat)
		assert.Equal(t, float32(12.110732), data.Lng)
		assert.Equal(t, "samsung", data.CameraMake)
		assert.Equal(t, "SM-G781B", data.CameraModel)
		assert.Equal(t, "", data.LensModel)
	})
	t.Run("SamsungGalaxyS21Ultra_JPG", func(t *testing.T) {
		data, err := JSON("testdata/motion/samsung_galaxys21ultra.jpg.json", "")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %#v", data)

		assert.Equal(t, "20211011_113427.jpg", data.FileName)
		assert.Equal(t, media.Live, data.MediaType)
		assert.Equal(t, true, data.EmbeddedThumb)
		assert.Equal(t, true, data.EmbeddedVideo)
		assert.Equal(t, CodecJpeg, data.Codec)
		assert.Equal(t, int64(0), data.Duration.Milliseconds())
		assert.Equal(t, "0s", data.Duration.String())
		assert.Equal(t, "2021-10-11 11:34:27 +0000 UTC", data.TakenAtLocal.String())
		assert.Equal(t, "2021-10-11 11:34:27 +0000 UTC", data.TakenAt.String())
		assert.Equal(t, 0, data.TakenNs)
		assert.Equal(t, "", data.TimeZone)
		assert.Equal(t, 4000, data.Width)
		assert.Equal(t, 2252, data.Height)
		assert.Equal(t, 4000, data.ActualWidth())
		assert.Equal(t, 2252, data.ActualHeight())
		assert.Equal(t, 1, data.Orientation)
		assert.Equal(t, float32(0), data.Lat)
		assert.Equal(t, float32(0), data.Lng)
		assert.Equal(t, "samsung", data.CameraMake)
		assert.Equal(t, "SM-G998B", data.CameraModel)
		assert.Equal(t, "", data.LensModel)
	})
	t.Run("SamsungGalaxyS21Ultra_MP4", func(t *testing.T) {
		data, err := JSON("testdata/motion/samsung_galaxys21ultra.mp4.json", "")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %#v", data)

		assert.Equal(t, "20211011_113427.mp4", data.FileName)
		assert.Equal(t, media.Unknown, data.MediaType)
		assert.Equal(t, false, data.EmbeddedThumb)
		assert.Equal(t, false, data.EmbeddedVideo)
		assert.Equal(t, CodecAvc1, data.Codec)
		assert.Equal(t, int64(2670), data.Duration.Milliseconds())
		assert.Equal(t, "2.67s", data.Duration.String())
		assert.Equal(t, "2021-10-11 09:34:29 +0000 UTC", data.TakenAtLocal.String())
		assert.Equal(t, "2021-10-11 09:34:29 +0000 UTC", data.TakenAt.String())
		assert.Equal(t, 0, data.TakenNs)
		assert.Equal(t, "UTC", data.TimeZone)
		assert.Equal(t, 1920, data.Width)
		assert.Equal(t, 1080, data.Height)
		assert.Equal(t, 1920, data.ActualWidth())
		assert.Equal(t, 1080, data.ActualHeight())
		assert.Equal(t, 1, data.Orientation)
		assert.Equal(t, float32(0), data.Lat)
		assert.Equal(t, float32(0), data.Lng)
		assert.Equal(t, "", data.CameraMake)
		assert.Equal(t, "", data.CameraModel)
		assert.Equal(t, "", data.LensModel)
	})
}
