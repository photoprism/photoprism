package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExif(t *testing.T) {
	t.Run("jpg file with exif data", func(t *testing.T) {
		data, err := Exif("testdata/photoshop.jpg")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Michael Mayer", data.Artist)
		assert.Equal(t, "2020-01-01T16:28:23Z", data.TakenAt.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "2020-01-01T17:28:23Z", data.TakenAtLocal.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "Example file for development", data.Description)
		assert.Equal(t, "This is a legal notice", data.Copyright)
		assert.Equal(t, 2736, data.Height)
		assert.Equal(t, 3648, data.Width)
		assert.Equal(t, 52.459690093888895, data.Lat)
		assert.Equal(t, 13.321831703055555, data.Lng)
		assert.Equal(t, 0, data.Altitude)
		assert.Equal(t, "HUAWEI", data.CameraMake)
		assert.Equal(t, "ELE-L29", data.CameraModel)
		assert.Equal(t, "", data.CameraOwner)
		assert.Equal(t, "", data.CameraSerial)
		assert.Equal(t, 27, data.FocalLength)
		assert.Equal(t, 1, int(data.Orientation))

		// TODO: Values are empty - why?
		// assert.Equal(t, "HUAWEI P30 Rear Main Camera", data.LensModel)
	})

	t.Run("ladybug jpg file", func(t *testing.T) {
		data, err := Exif("testdata/ladybug.jpg")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("all: %+v", data.All)

		assert.Equal(t, "Photographer: TMB", data.Artist)
		assert.Equal(t, "2011-07-10T17:34:28Z", data.TakenAt.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "2011-07-10T19:34:28Z", data.TakenAtLocal.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "", data.Title)    // Should be "Ladybug"
		assert.Equal(t, "", data.Keywords) // Should be "Ladybug"
		assert.Equal(t, "", data.Description)
		assert.Equal(t, "", data.Copyright)
		assert.Equal(t, 540, data.Height)
		assert.Equal(t, 720, data.Width)
		assert.Equal(t, 51.25485166666667, data.Lat)
		assert.Equal(t, 7.389468333333333, data.Lng)
		assert.Equal(t, 0, data.Altitude)
		assert.Equal(t, "Canon", data.CameraMake)
		assert.Equal(t, "Canon EOS 50D", data.CameraModel)
		assert.Equal(t, "Thomas Meyer-Boudnik", data.CameraOwner)
		assert.Equal(t, "2260716910", data.CameraSerial)
		assert.Equal(t, 100, data.FocalLength)
		assert.Equal(t, 1, int(data.Orientation))
	})

	t.Run("GoPro HD2 jpg file", func(t *testing.T) {
		data, err := Exif("testdata/gopro_hd2.jpg")

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("all: %+v", data.All)

		assert.Equal(t, "", data.Artist)
		assert.Equal(t, "2017-12-21T05:17:28Z", data.TakenAt.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "2017-12-21T05:17:28Z", data.TakenAtLocal.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "", data.Title)    // Should be "Ladybug"
		assert.Equal(t, "", data.Keywords) // Should be "Ladybug"
		assert.Equal(t, "DCIM\\100GOPRO", data.Description)
		assert.Equal(t, "", data.Copyright)
		assert.Equal(t, 2880, data.Height)
		assert.Equal(t, 3840, data.Width)
		assert.Equal(t, 0.0, data.Lng)
		assert.Equal(t, 0, data.Altitude)
		assert.Equal(t, "GoPro", data.CameraMake)
		assert.Equal(t, "HD2", data.CameraModel)
		assert.Equal(t, "", data.CameraOwner)
		assert.Equal(t, "", data.CameraSerial)
		assert.Equal(t, 16, data.FocalLength)
		assert.Equal(t, 1, int(data.Orientation))
	})

	t.Run("png file without exif", func(t *testing.T) {
		_, err := Exif("testdata/tweethog.png")

		assert.Error(t, err, "file does not have EXIF")
		// TODO: png with exif data
	})

	t.Run("heic file with exif data", func(t *testing.T) {
		data, err := Exif("testdata/iphone_7.heic")
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "2018-09-10T03:16:13Z", data.TakenAt.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "2018-09-10T12:16:13Z", data.TakenAtLocal.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, 34.79745, data.Lat)
		assert.Equal(t, 134.76463333333334, data.Lng)
		assert.Equal(t, 0, data.Altitude)
		assert.Equal(t, "Apple", data.CameraMake)
		assert.Equal(t, "iPhone 7", data.CameraModel)
		assert.Equal(t, 74, data.FocalLength)
		assert.Equal(t, 6, int(data.Orientation))
		assert.Equal(t, "Apple", data.LensMake)
		assert.Equal(t, "iPhone 7 back camera 3.99mm f/1.8", data.LensModel)

	})
}
