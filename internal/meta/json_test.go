package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSON(t *testing.T) {
	t.Run("iphone-mov.json", func(t *testing.T) {
		data, err := JSON("testdata/iphone-mov.json")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %+v", data)

		assert.Equal(t, CodecAvc1, data.Codec)
		assert.Equal(t, "3s", data.Duration.String())
		assert.Equal(t, "2018-09-08 17:20:14 +0000 UTC", data.TakenAtLocal.String())
		assert.Equal(t, "2018-09-08 15:20:14 +0000 UTC", data.TakenAt.String())
		assert.Equal(t, "Europe/Berlin", data.TimeZone)
		assert.Equal(t, 1080, data.Width)
		assert.Equal(t, 1920, data.Height)
		assert.Equal(t, float32(52.4587), data.Lat)
		assert.Equal(t, float32(13.4593), data.Lng)
		assert.Equal(t, "Apple", data.CameraMake)
		assert.Equal(t, "iPhone SE", data.CameraModel)
		assert.Equal(t, "", data.LensModel)
	})

	t.Run("gopher-telegram.json", func(t *testing.T) {
		data, err := JSON("testdata/gopher-telegram.json")

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
		assert.Equal(t, float32(0), data.Lat)
		assert.Equal(t, float32(0), data.Lng)
		assert.Equal(t, "", data.CameraMake)
		assert.Equal(t, "", data.CameraModel)
		assert.Equal(t, "", data.LensModel)
	})

	t.Run("gopher-original.json", func(t *testing.T) {
		data, err := JSON("testdata/gopher-original.json")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %+v", data)

		assert.Equal(t, CodecAvc1, data.Codec)
		assert.Equal(t, "2s", data.Duration.String())
		assert.Equal(t, "2020-05-11 14:16:48 +0000 UTC", data.TakenAtLocal.String())
		assert.Equal(t, "2020-05-11 12:16:48 +0000 UTC", data.TakenAt.String())
		assert.Equal(t, "Europe/Berlin", data.TimeZone)
		assert.Equal(t, 1080, data.Width)
		assert.Equal(t, 1920, data.Height)
		assert.Equal(t, float32(52.4596), data.Lat)
		assert.Equal(t, float32(13.3218), data.Lng)
		assert.Equal(t, "", data.CameraMake)
		assert.Equal(t, "", data.CameraModel)
		assert.Equal(t, "", data.LensModel)
	})

	t.Run("berlin-landscape.json", func(t *testing.T) {
		data, err := JSON("testdata/berlin-landscape.json")

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
		assert.Equal(t, float32(52.4649), data.Lat)
		assert.Equal(t, float32(13.3148), data.Lng)
		assert.Equal(t, "", data.CameraMake)
		assert.Equal(t, "", data.CameraModel)
		assert.Equal(t, "", data.LensModel)
	})

	t.Run("mp4.json", func(t *testing.T) {
		data, err := JSON("testdata/mp4.json")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %+v", data)

		assert.Equal(t, CodecAvc1, data.Codec)
		assert.Equal(t, "4m25s", data.Duration.String())
		assert.Equal(t, "2019-11-23 13:51:49 +0000 UTC", data.TakenAtLocal.String())
		assert.Equal(t, 848, data.Width)
		assert.Equal(t, 480, data.Height)
		assert.Equal(t, "", data.Copyright)
		assert.Equal(t, "", data.CameraMake)
		assert.Equal(t, "", data.CameraModel)
		assert.Equal(t, "", data.LensModel)
	})

	t.Run("photoshop.json", func(t *testing.T) {
		data, err := JSON("testdata/photoshop.json")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %+v", data)

		assert.Equal(t, CodecUnknown, data.Codec)
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
	})

	t.Run("canon_eos_6d.json", func(t *testing.T) {
		data, err := JSON("testdata/canon_eos_6d.json")

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
	})

	t.Run("gps-2000.json", func(t *testing.T) {
		data, err := JSON("testdata/gps-2000.json")

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
	})

	t.Run("ladybug.json", func(t *testing.T) {
		data, err := JSON("testdata/ladybug.json")

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
	})

	t.Run("iphone_7.json", func(t *testing.T) {
		data, err := JSON("testdata/iphone_7.json")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("DATA: %+v", data)

		assert.Equal(t, CodecUnknown, data.Codec)
		assert.Equal(t, "", data.Title)
		assert.Equal(t, "", data.Artist)
		assert.Equal(t, "", data.Description)
		assert.Equal(t, "", data.Copyright)
		assert.Equal(t, "Apple", data.CameraMake)
		assert.Equal(t, "iPhone 7", data.CameraModel)
		assert.Equal(t, "Apple", data.LensMake)
		assert.Equal(t, "iPhone 7 back camera 3.99mm f/1.8", data.LensModel)
	})

}
