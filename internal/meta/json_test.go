package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSON(t *testing.T) {
	t.Run("iphone-mov.json", func(t *testing.T) {
		data, err := JSON("testdata/iphone-mov.json", "")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %+v", data)

		assert.Equal(t, CodecAvc1, data.Codec)
		assert.Equal(t, "3s", data.Duration.String())
		assert.Equal(t, "2018-09-08 17:20:14 +0000 UTC", data.TakenAtLocal.String())
		assert.Equal(t, "2018-09-08 15:20:14 +0000 UTC", data.TakenAt.String())
		assert.Equal(t, "Europe/Berlin", data.TimeZone)
		assert.Equal(t, 1920, data.Width)
		assert.Equal(t, 1080, data.Height)
		assert.Equal(t, 1080, data.ActualWidth())
		assert.Equal(t, 1920, data.ActualHeight())
		assert.Equal(t, 6, data.Orientation)
		assert.Equal(t, float32(52.4587), data.Lat)
		assert.Equal(t, float32(13.4593), data.Lng)
		assert.Equal(t, "Apple", data.CameraMake)
		assert.Equal(t, "iPhone SE", data.CameraModel)
		assert.Equal(t, "", data.LensModel)
	})

	t.Run("gopher-telegram.json", func(t *testing.T) {
		data, err := JSON("testdata/gopher-telegram.json", "")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %+v", data)

		assert.Equal(t, CodecAvc1, data.Codec)
		assert.Equal(t, "2s", data.Duration.String())
		assert.Equal(t, "2020-05-11 14:18:35 +0000 UTC", data.TakenAtLocal.String())
		assert.Equal(t, "2020-05-11 14:18:35 +0000 UTC", data.TakenAt.String())
		assert.Equal(t, "", data.TimeZone)
		assert.Equal(t, 270, data.Width)
		assert.Equal(t, 480, data.Height)
		assert.Equal(t, 270, data.ActualWidth())
		assert.Equal(t, 480, data.ActualHeight())
		assert.Equal(t, 1, data.Orientation)
		assert.Equal(t, float32(0), data.Lat)
		assert.Equal(t, float32(0), data.Lng)
		assert.Equal(t, "", data.CameraMake)
		assert.Equal(t, "", data.CameraModel)
		assert.Equal(t, "", data.LensModel)
	})

	t.Run("gopher-original.json", func(t *testing.T) {
		data, err := JSON("testdata/gopher-original.json", "")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %+v", data)

		assert.Equal(t, CodecAvc1, data.Codec)
		assert.Equal(t, "2s", data.Duration.String())
		assert.Equal(t, "2020-05-11 14:16:48 +0000 UTC", data.TakenAtLocal.String())
		assert.Equal(t, "2020-05-11 12:16:48 +0000 UTC", data.TakenAt.String())
		assert.Equal(t, "Europe/Berlin", data.TimeZone)
		assert.Equal(t, 1920, data.Width)
		assert.Equal(t, 1080, data.Height)
		assert.Equal(t, 1080, data.ActualWidth())
		assert.Equal(t, 1920, data.ActualHeight())
		assert.Equal(t, float32(0.5625), data.AspectRatio())
		assert.Equal(t, 6, data.Orientation)
		assert.Equal(t, float32(52.4596), data.Lat)
		assert.Equal(t, float32(13.3218), data.Lng)
		assert.Equal(t, "", data.CameraMake)
		assert.Equal(t, "", data.CameraModel)
		assert.Equal(t, "", data.LensModel)
	})

	t.Run("berlin-landscape.json", func(t *testing.T) {
		data, err := JSON("testdata/berlin-landscape.json", "")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %+v", data)

		assert.Equal(t, CodecAvc1, data.Codec)
		assert.Equal(t, "4s", data.Duration.String())
		assert.Equal(t, "2020-05-14 11:34:41 +0000 UTC", data.TakenAtLocal.String())
		assert.Equal(t, "2020-05-14 09:34:41 +0000 UTC", data.TakenAt.String())
		assert.Equal(t, "Europe/Berlin", data.TimeZone)
		assert.Equal(t, 1920, data.Width)
		assert.Equal(t, 1080, data.Height)
		assert.Equal(t, 1, data.Orientation)
		assert.Equal(t, float32(52.4649), data.Lat)
		assert.Equal(t, float32(13.3148), data.Lng)
		assert.Equal(t, "", data.CameraMake)
		assert.Equal(t, "", data.CameraModel)
		assert.Equal(t, "", data.LensModel)
	})

	t.Run("mp4.json", func(t *testing.T) {
		data, err := JSON("testdata/mp4.json", "")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %+v", data)

		assert.Equal(t, CodecAvc1, data.Codec)
		assert.Equal(t, "4m25s", data.Duration.String())
		assert.Equal(t, "2019-11-23 13:51:49 +0000 UTC", data.TakenAtLocal.String())
		assert.Equal(t, 848, data.Width)
		assert.Equal(t, 480, data.Height)
		assert.Equal(t, 1, data.Orientation)
		assert.Equal(t, "", data.Copyright)
		assert.Equal(t, "", data.CameraMake)
		assert.Equal(t, "", data.CameraModel)
		assert.Equal(t, "", data.LensModel)
	})

	t.Run("photoshop.json", func(t *testing.T) {
		data, err := JSON("testdata/photoshop.json", "")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %+v", data)

		assert.Equal(t, CodecXMP, data.Codec)
		assert.Equal(t, "0s", data.Duration.String())
		assert.Equal(t, float32(52.45969), data.Lat)
		assert.Equal(t, float32(13.321831), data.Lng)
		assert.Equal(t, "2020-01-01 16:28:23 +0000 UTC", data.TakenAt.String())
		assert.Equal(t, "2020-01-01 17:28:23 +0000 UTC", data.TakenAtLocal.String())
		assert.Equal(t, "Europe/Berlin", data.TimeZone)
		assert.Equal(t, "Night Shift / Berlin / 2020", data.Title)
		assert.Equal(t, "Michael Mayer", data.Artist)
		assert.Equal(t, "Example file for development", data.Description)
		assert.Equal(t, "This is an (edited) legal notice", data.Copyright)
		assert.Equal(t, "HUAWEI", data.CameraMake)
		assert.Equal(t, "ELE-L29", data.CameraModel)
		assert.Equal(t, "HUAWEI P30 Rear Main Camera", data.LensModel)
		assert.Equal(t, 1, data.Orientation)
	})

	t.Run("canon_eos_6d.json", func(t *testing.T) {
		data, err := JSON("testdata/canon_eos_6d.json", "")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %+v", data)

		assert.Equal(t, CodecJpeg, data.Codec)
		assert.Equal(t, "", data.Title)
		assert.Equal(t, "", data.Artist)
		assert.Equal(t, "", data.Description)
		assert.Equal(t, "", data.Copyright)
		assert.Equal(t, "Canon", data.CameraMake)
		assert.Equal(t, "Canon EOS 6D", data.CameraModel)
		assert.Equal(t, "EF24-105mm f/4L IS USM", data.LensModel)
		assert.Equal(t, 1, data.Orientation)
	})

	t.Run("gps-2000.json", func(t *testing.T) {
		data, err := JSON("testdata/gps-2000.json", "")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %+v", data)

		assert.Equal(t, CodecUnknown, data.Codec)
		assert.Equal(t, "", data.Title)
		assert.Equal(t, "", data.Artist)
		assert.Equal(t, "", data.Description)
		assert.Equal(t, "", data.Copyright)
		assert.Equal(t, "", data.CameraMake)
		assert.Equal(t, "", data.CameraModel)
		assert.Equal(t, "", data.LensMake)
		assert.Equal(t, "", data.LensModel)
		assert.Equal(t, 1, data.Orientation)
	})

	t.Run("ladybug.json", func(t *testing.T) {
		data, err := JSON("testdata/ladybug.json", "")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %+v", data)

		assert.Equal(t, CodecUnknown, data.Codec)
		assert.Equal(t, "", data.Title)
		assert.Equal(t, "", data.Artist)
		assert.Equal(t, "", data.Description)
		assert.Equal(t, "", data.Copyright)
		assert.Equal(t, "", data.CameraMake)
		assert.Equal(t, "", data.CameraModel)
		assert.Equal(t, "", data.LensMake)
		assert.Equal(t, "", data.LensModel)
		assert.Equal(t, 1, data.Orientation)
	})

	t.Run("iphone_7.json", func(t *testing.T) {
		data, err := JSON("testdata/iphone_7.json", "")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %+v", data)

		assert.Equal(t, CodecHeic, data.Codec)
		assert.Equal(t, "", data.Title)
		assert.Equal(t, "", data.Artist)
		assert.Equal(t, "", data.Description)
		assert.Equal(t, "", data.Copyright)
		assert.Equal(t, 6, data.Orientation)
		assert.Equal(t, "Apple", data.CameraMake)
		assert.Equal(t, "iPhone 7", data.CameraModel)
		assert.Equal(t, "Apple", data.LensMake)
		assert.Equal(t, "iPhone 7 back camera 3.99mm f/1.8", data.LensModel)
	})

	t.Run("uuid-original.json", func(t *testing.T) {
		data, err := JSON("testdata/uuid-original.json", "")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %+v", data)

		assert.Equal(t, "9bafc58c-6c82-4e66-a45f-c13f915f99c5", data.DocumentID)
		assert.Equal(t, "", data.InstanceID)
		assert.Equal(t, CodecJpeg, data.Codec)
		assert.Equal(t, "0s", data.Duration.String())
		assert.Equal(t, "2018-12-06 12:32:26 +0000 UTC", data.TakenAtLocal.String())
		assert.Equal(t, "2018-12-06 11:32:26 +0000 UTC", data.TakenAt.String())
		assert.Equal(t, "Europe/Berlin", data.TimeZone)
		assert.Equal(t, 3024, data.Width)
		assert.Equal(t, 4032, data.Height)
		assert.Equal(t, 1, data.Orientation)
		assert.Equal(t, float32(48.300003), data.Lat)
		assert.Equal(t, float32(8.929067), data.Lng)
		assert.Equal(t, "Apple", data.CameraMake)
		assert.Equal(t, "iPhone SE", data.CameraModel)
		assert.Equal(t, "iPhone SE back camera 4.15mm f/2.2", data.LensModel)
	})

	t.Run("uuid-copy.json", func(t *testing.T) {
		data, err := JSON("testdata/uuid-copy.json", "")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %+v", data)

		assert.Equal(t, "4b1fef2d1cf4a5be38b263e0637edead", data.DocumentID)
		assert.Equal(t, "dafbfeb8-a129-4e7c-9cf0-e7996a701cdb", data.InstanceID)
		assert.Equal(t, CodecJpeg, data.Codec)
		assert.Equal(t, "0s", data.Duration.String())
		assert.Equal(t, "2018-12-06 12:32:26 +0000 UTC", data.TakenAtLocal.String())
		assert.Equal(t, "2018-12-06 11:32:26 +0000 UTC", data.TakenAt.String())
		assert.Equal(t, "Europe/Berlin", data.TimeZone)
		assert.Equal(t, 1024, data.Width)
		assert.Equal(t, 1365, data.Height)
		assert.Equal(t, 1, data.Orientation)
		assert.Equal(t, float32(48.300003), data.Lat)
		assert.Equal(t, float32(8.929067), data.Lng)
		assert.Equal(t, "Apple", data.CameraMake)
		assert.Equal(t, "iPhone SE", data.CameraModel)
		assert.Equal(t, "iPhone SE back camera 4.15mm f/2.2", data.LensModel)
	})

	t.Run("uuid-imagemagick.json", func(t *testing.T) {
		data, err := JSON("testdata/uuid-imagemagick.json", "")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %+v", data)

		assert.Equal(t, "9bafc58c-6c82-4e66-a45f-c13f915f99c5", data.DocumentID)
		assert.Equal(t, "", data.InstanceID)
		assert.Equal(t, CodecJpeg, data.Codec)
		assert.Equal(t, "0s", data.Duration.String())
		assert.Equal(t, "2018-12-06 12:32:26 +0000 UTC", data.TakenAtLocal.String())
		assert.Equal(t, "2018-12-06 11:32:26 +0000 UTC", data.TakenAt.String())
		assert.Equal(t, "Europe/Berlin", data.TimeZone)
		assert.Equal(t, 1125, data.Width)
		assert.Equal(t, 1500, data.Height)
		assert.Equal(t, 1, data.Orientation)
		assert.Equal(t, float32(48.300003), data.Lat)
		assert.Equal(t, float32(8.929067), data.Lng)
		assert.Equal(t, "Apple", data.CameraMake)
		assert.Equal(t, "iPhone SE", data.CameraModel)
		assert.Equal(t, "iPhone SE back camera 4.15mm f/2.2", data.LensModel)
	})

	t.Run("orientation.json", func(t *testing.T) {
		data, err := JSON("testdata/orientation.json", "")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 326, data.Width)
		assert.Equal(t, 184, data.Height)
		assert.Equal(t, 1, data.Orientation)
	})
}
