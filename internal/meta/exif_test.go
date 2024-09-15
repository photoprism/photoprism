package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestExif(t *testing.T) {
	t.Run("iptc-2014.jpg", func(t *testing.T) {
		data, err := Exif("testdata/iptc-2014.jpg", fs.ImageJPEG, true)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Creator1 (ref2014)", data.Artist)
		assert.Equal(t, "2011-10-28T12:00:00Z", data.TakenAt.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "2011-10-28T12:00:00Z", data.TakenAtLocal.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, 0, data.TakenNs)
		assert.Equal(t, "The description aka caption (ref2014)", data.Description)
		assert.Equal(t, "Copyright (Notice) 2014 IPTC - www.iptc.org  (ref2014)", data.Copyright)
		assert.Equal(t, "Adobe Photoshop CC 2014 (Windows)", data.Software)
		assert.Equal(t, 1050, data.Height)
		assert.Equal(t, 2100, data.Width)
		assert.Equal(t, 0.0, data.Lat)
		assert.Equal(t, 0.0, data.Lng)
		assert.Equal(t, 0.0, data.Altitude)
		assert.Equal(t, "", data.Exposure)
		assert.Equal(t, "", data.CameraMake)
		assert.Equal(t, "", data.CameraModel)
		assert.Equal(t, "", data.CameraOwner)
		assert.Equal(t, "", data.CameraSerial)
		assert.Equal(t, 0, data.FocalLength)
		assert.Equal(t, 1, data.Orientation)
	})

	t.Run("iptc-2016.jpg", func(t *testing.T) {
		data, err := Exif("testdata/iptc-2016.jpg", fs.ImageJPEG, true)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Creator1 (ref2016)", data.Artist)
		assert.Equal(t, "2011-10-28T12:00:00Z", data.TakenAt.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "2011-10-28T12:00:00Z", data.TakenAtLocal.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, 0, data.TakenNs)
		assert.Equal(t, "The description aka caption (ref2016)", data.Description)
		assert.Equal(t, "Copyright (Notice) 2016 IPTC - www.iptc.org  (ref2016)", data.Copyright)
		assert.Equal(t, "Adobe Photoshop CC 2017 (Windows)", data.Software)
		assert.Equal(t, 1050, data.Height)
		assert.Equal(t, 2100, data.Width)
		assert.Equal(t, 0.0, data.Lat)
		assert.Equal(t, 0.0, data.Lng)
		assert.Equal(t, 0.0, data.Altitude)
		assert.Equal(t, "", data.Exposure)
		assert.Equal(t, "", data.CameraMake)
		assert.Equal(t, "", data.CameraModel)
		assert.Equal(t, "", data.CameraOwner)
		assert.Equal(t, "", data.CameraSerial)
		assert.Equal(t, 0, data.FocalLength)
		assert.Equal(t, 1, data.Orientation)
	})

	t.Run("photoshop.jpg", func(t *testing.T) {
		data, err := Exif("testdata/photoshop.jpg", fs.ImageJPEG, true)

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("all: %+v", data.exif)

		assert.Equal(t, "Michael Mayer", data.Artist)
		assert.Equal(t, "2020-01-01T16:28:23Z", data.TakenAt.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "2020-01-01T17:28:23Z", data.TakenAtLocal.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, 899614000, data.TakenNs)
		assert.Equal(t, "Example file for development", data.Description)
		assert.Equal(t, "This is a legal notice", data.Copyright)
		assert.Equal(t, "Adobe Photoshop 21.0 (Macintosh)", data.Software)
		assert.Equal(t, 540, data.Height)
		assert.Equal(t, 720, data.Width)
		assert.InEpsilon(t, 52.45969, data.Lat, 0.00001)
		assert.InEpsilon(t, 13.321832, data.Lng, 0.00001)
		assert.Equal(t, 0.0, data.Altitude)
		assert.Equal(t, "1/50", data.Exposure)
		assert.Equal(t, "HUAWEI", data.CameraMake)
		assert.Equal(t, "ELE-L29", data.CameraModel)
		assert.Equal(t, "", data.CameraOwner)
		assert.Equal(t, "", data.CameraSerial)
		assert.Equal(t, 27, data.FocalLength)
		assert.Equal(t, 1, data.Orientation)

		// TODO: Values are empty - why?
		// assert.Equal(t, "HUAWEI P30 Rear Main Camera", data.LensModel)
	})

	t.Run("ladybug.jpg", func(t *testing.T) {
		data, err := Exif("testdata/ladybug.jpg", fs.ImageJPEG, true)

		if err != nil {
			t.Fatal(err)
		}

		//  t.Logf("all: %+v", data.exif)

		assert.Equal(t, "Photographer: TMB", data.Artist)
		assert.Equal(t, "2011-07-10T17:34:28Z", data.TakenAt.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "2011-07-10T19:34:28Z", data.TakenAtLocal.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, 150000000, data.TakenNs)
		assert.Equal(t, "", data.Title)             // Should be "Ladybug"
		assert.Equal(t, "", data.Keywords.String()) // Should be "Ladybug"
		assert.Equal(t, "", data.Description)
		assert.Equal(t, "", data.Copyright)
		assert.Equal(t, 540, data.Height)
		assert.Equal(t, 720, data.Width)
		assert.InEpsilon(t, 51.254852, data.Lat, 0.00001)
		assert.InEpsilon(t, 7.389468, data.Lng, 0.00001)
		assert.Equal(t, 0.0, data.Altitude)
		assert.Equal(t, "1/125", data.Exposure)
		assert.Equal(t, "Canon", data.CameraMake)
		assert.Equal(t, "Canon EOS 50D", data.CameraModel)
		assert.Equal(t, "Thomas Meyer-Boudnik", data.CameraOwner)
		assert.Equal(t, "2260716910", data.CameraSerial)
		assert.Equal(t, "", data.LensMake)
		assert.Equal(t, "EF100mm f/2.8 Macro USM", data.LensModel)
		assert.Equal(t, 100, data.FocalLength)
		assert.Equal(t, 1, data.Orientation)
	})

	t.Run("gopro_hd2.jpg", func(t *testing.T) {
		data, err := Exif("testdata/gopro_hd2.jpg", fs.ImageJPEG, true)

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("all: %+v", data.exif)

		assert.Equal(t, "", data.Artist)
		assert.Equal(t, "2017-12-21T05:17:28Z", data.TakenAt.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "2017-12-21T05:17:28Z", data.TakenAtLocal.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "", data.Title)
		assert.Equal(t, "", data.Keywords.String())
		assert.Equal(t, "", data.Description)
		assert.Equal(t, "", data.Copyright)
		assert.Equal(t, 180, data.Height)
		assert.Equal(t, 240, data.Width)
		assert.Equal(t, 0.0, data.Lat)
		assert.Equal(t, 0.0, data.Lng)
		assert.Equal(t, 0.0, data.Altitude)
		assert.Equal(t, "1/2462", data.Exposure)
		assert.Equal(t, "GoPro", data.CameraMake)
		assert.Equal(t, "HD2", data.CameraModel)
		assert.Equal(t, "", data.CameraOwner)
		assert.Equal(t, "", data.CameraSerial)
		assert.Equal(t, 16, data.FocalLength)
		assert.Equal(t, 1, data.Orientation)
	})

	t.Run("tweethog.png", func(t *testing.T) {
		_, err := Exif("testdata/tweethog.png", fs.ImagePNG, true)

		if err == nil {
			t.Fatal("err should NOT be nil")
		}

		assert.Equal(t, "found no exif header", err.Error())
	})

	t.Run("iphone_7.heic", func(t *testing.T) {
		data, err := Exif("testdata/iphone_7.heic", fs.ImageHEIC, true)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "2018-09-10T03:16:13Z", data.TakenAt.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "2018-09-10T12:16:13Z", data.TakenAtLocal.Format("2006-01-02T15:04:05Z"))
		assert.InEpsilon(t, 34.79745, data.Lat, 0.00001)
		assert.InEpsilon(t, 134.76463, data.Lng, 0.00001)
		assert.Equal(t, 0.0, data.Altitude)
		assert.Equal(t, "1/4000", data.Exposure)
		assert.Equal(t, "Apple", data.CameraMake)
		assert.Equal(t, "iPhone 7", data.CameraModel)
		assert.Equal(t, 74, data.FocalLength)
		assert.Equal(t, 6, data.Orientation)
		assert.Equal(t, "Apple", data.LensMake)
		assert.Equal(t, "iPhone 7 back camera 3.99mm f/1.8", data.LensModel)
		assert.Equal(t, "", data.ColorProfile)
	})

	t.Run("gps-2000.jpg", func(t *testing.T) {
		data, err := Exif("testdata/gps-2000.jpg", fs.ImageJPEG, true)

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("GPS 2000: %+v", data.exif)

		assert.Equal(t, "", data.Artist)
		assert.True(t, data.TakenAt.IsZero())
		assert.True(t, data.TakenAtLocal.IsZero())
		assert.Equal(t, "", data.Description)
		assert.Equal(t, "", data.Copyright)
		assert.Equal(t, 0, data.Height)
		assert.Equal(t, 0, data.Width)
		assert.InEpsilon(t, -38.405193, data.Lat, 0.00001)
		assert.InEpsilon(t, 144.18896, data.Lng, 0.00001)
		assert.Equal(t, 0.0, data.Altitude)
		assert.Equal(t, "", data.Exposure)
		assert.Equal(t, "", data.CameraMake)
		assert.Equal(t, "", data.CameraModel)
		assert.Equal(t, "", data.CameraOwner)
		assert.Equal(t, "", data.CameraSerial)
		assert.Equal(t, 0, data.FocalLength)
		assert.Equal(t, 1, data.Orientation)
	})

	t.Run("image-2011.jpg", func(t *testing.T) {
		data, err := Exif("testdata/image-2011.jpg", fs.ImageJPEG, true)

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("ALL: %+v", data.exif)

		/*
		  Exiftool date information:

		  File Modification Date/Time     : 2020:05:15 08:25:46+00:00
		  File Access Date/Time           : 2020:05:15 08:25:47+00:00
		  File Inode Change Date/Time     : 2020:05:15 08:25:46+00:00
		  Modify Date                     : 2020:05:15 10:25:45
		  Create Date                     : 2011:07:19 11:36:38
		  Metadata Date                   : 2020:05:15 10:25:45+02:00

		*/

		t.Logf("TakenGps: %s", data.TakenGps)

		assert.Equal(t, "2020-05-15T10:25:45Z", data.TakenAt.Format("2006-01-02T15:04:05Z"))      // TODO
		assert.Equal(t, "2020-05-15T10:25:45Z", data.TakenAtLocal.Format("2006-01-02T15:04:05Z")) // TODO
		assert.Equal(t, 0.0, data.Lat)
		assert.Equal(t, 0.0, data.Lng)
		assert.Equal(t, 0.0, data.Altitude)
		assert.Equal(t, "1/1100", data.Exposure)
		assert.Equal(t, "SAMSUNG", data.CameraMake)
		assert.Equal(t, "GT-I9000", data.CameraModel)
		assert.Equal(t, 3, data.FocalLength)
		assert.Equal(t, 1, data.Orientation)
		assert.Equal(t, "", data.LensMake)
		assert.Equal(t, "", data.LensModel)
	})

	t.Run("ship.jpg", func(t *testing.T) {
		data, err := Exif("testdata/ship.jpg", fs.ImageJPEG, true)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "2019-05-12T15:13:53Z", data.TakenAt.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "2019-05-12T17:13:53Z", data.TakenAtLocal.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, 955677000, data.TakenNs)
		assert.InEpsilon(t, 53.12349, data.Lat, 0.00001)
		assert.InEpsilon(t, 18.00152, data.Lng, 0.00001)
		assert.Equal(t, 63, clean.Altitude(data.Altitude))
		assert.Equal(t, "1/100", data.Exposure)
		assert.Equal(t, "Xiaomi", data.CameraMake)
		assert.Equal(t, "Mi A1", data.CameraModel)
		assert.Equal(t, 52, data.FocalLength)
		assert.Equal(t, 1, data.Orientation)
		assert.Equal(t, "", data.LensMake)
		assert.Equal(t, "", data.LensModel)
	})

	t.Run("no-exif-data.jpg", func(t *testing.T) {
		_, err := Exif("testdata/no-exif-data.jpg", fs.ImageJPEG, false)

		if err == nil {
			t.Fatal("err should NOT be nil")
		}

		assert.Equal(t, "found no exif header", err.Error())
	})

	t.Run("no-exif-data.jpg/BruteForce", func(t *testing.T) {
		_, err := Exif("testdata/no-exif-data.jpg", fs.ImageJPEG, true)

		if err == nil {
			t.Fatal("err should NOT be nil")
		}

		assert.Equal(t, "found no exif data", err.Error())
	})

	t.Run("screenshot.png", func(t *testing.T) {
		data, err := Exif("testdata/screenshot.png", fs.ImagePNG, true)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "721", data.exif["PixelXDimension"])
		assert.Equal(t, "332", data.exif["PixelYDimension"])
	})

	t.Run("orientation.jpg", func(t *testing.T) {
		data, err := Exif("testdata/orientation.jpg", fs.ImageJPEG, true)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "3264", data.exif["PixelXDimension"])
		assert.Equal(t, "1836", data.exif["PixelYDimension"])
		assert.Equal(t, 3264, data.Width)
		assert.Equal(t, 1836, data.Height)
		assert.Equal(t, 1, data.Orientation) // TODO: Should be 1

		if err := data.JSON("testdata/orientation.json", "orientation.jpg"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 326, data.Width)
		assert.Equal(t, 184, data.Height)
		assert.Equal(t, 1, data.Orientation)

		if err := data.JSON("testdata/orientation.json", "foo.jpg"); err != nil {
			assert.EqualError(t, err, "metadata: original name foo.jpg does not match orientation.jpg (exiftool)")
		} else {
			t.Error("error expected when providing wrong original name")
		}
	})

	t.Run("gopher-preview.jpg", func(t *testing.T) {
		_, err := Exif("testdata/gopher-preview.jpg", fs.ImageJPEG, false)

		assert.EqualError(t, err, "found no exif header")
	})

	t.Run("gopher-preview.jpg/BruteForce", func(t *testing.T) {
		_, err := Exif("testdata/gopher-preview.jpg", fs.ImageJPEG, true)

		assert.EqualError(t, err, "found no exif data")
	})

	t.Run("huawei-gps-error.jpg", func(t *testing.T) {
		data, err := Exif("testdata/huawei-gps-error.jpg", fs.ImageJPEG, true)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "2020-06-16T16:52:46Z", data.TakenAt.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "2020-06-16T18:52:46Z", data.TakenAtLocal.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, 695326000, data.TakenNs)
		assert.InEpsilon(t, 48.302776, data.Lat, 0.00001)
		assert.InEpsilon(t, 8.9275, data.Lng, 0.00001)
		assert.Equal(t, 0.0, data.Altitude)
		assert.Equal(t, "1/110", data.Exposure)
		assert.Equal(t, "HUAWEI", data.CameraMake)
		assert.Equal(t, "ELE-L29", data.CameraModel)
		assert.Equal(t, 27, data.FocalLength)
		assert.Equal(t, 0, data.Orientation)
		assert.Equal(t, "", data.LensMake)
		assert.Equal(t, "", data.LensModel)
	})

	t.Run("panorama360.jpg", func(t *testing.T) {
		data, err := Exif("testdata/panorama360.jpg", fs.ImageJPEG, true)

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("all: %+v", data.exif)

		assert.Equal(t, "", data.Artist)
		assert.Equal(t, "2020-05-24T08:55:21Z", data.TakenAt.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "2020-05-24T11:55:21Z", data.TakenAtLocal.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "", data.Title)
		assert.Equal(t, "", data.Keywords.String())
		assert.Equal(t, "", data.Description)
		assert.Equal(t, "", data.Copyright)
		assert.Equal(t, 3600, data.Height)
		assert.Equal(t, 7200, data.Width)
		assert.InEpsilon(t, 59.84083, data.Lat, 0.00001)
		assert.InEpsilon(t, 30.51, data.Lng, 0.00001)
		assert.Equal(t, 0.0, data.Altitude)
		assert.Equal(t, "1/1250", data.Exposure)
		assert.Equal(t, "SAMSUNG", data.CameraMake)
		assert.Equal(t, "SM-C200", data.CameraModel)
		assert.Equal(t, "", data.CameraOwner)
		assert.Equal(t, "", data.CameraSerial)
		assert.Equal(t, 6, data.FocalLength)
		assert.Equal(t, 1, data.Orientation)
		assert.Equal(t, "", data.Projection)
		assert.Equal(t, "", data.ColorProfile)
	})

	t.Run("exif-example.tiff", func(t *testing.T) {
		data, err := Exif("testdata/exif-example.tiff", fs.ImageTIFF, true)

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("all: %+v", data.exif)

		assert.Equal(t, "", data.Artist)
		assert.Equal(t, "0001-01-01T00:00:00Z", data.TakenAt.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "0001-01-01T00:00:00Z", data.TakenAtLocal.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "", data.Title)
		assert.Equal(t, "", data.Keywords.String())
		assert.Equal(t, "", data.Description)
		assert.Equal(t, "", data.Copyright)
		assert.Equal(t, 43, data.Height)
		assert.Equal(t, 65, data.Width)
		assert.Equal(t, 0.0, data.Lat)
		assert.Equal(t, 0.0, data.Lng)
		assert.Equal(t, 0.0, data.Altitude)
		assert.Equal(t, "", data.Exposure)
		assert.Equal(t, "", data.CameraMake)
		assert.Equal(t, "", data.CameraModel)
		assert.Equal(t, "", data.CameraOwner)
		assert.Equal(t, "", data.CameraSerial)
		assert.Equal(t, 0, data.FocalLength)
		assert.Equal(t, 1, data.Orientation)
		assert.Equal(t, "", data.Projection)
		assert.Equal(t, "", data.ColorProfile)
	})

	t.Run("out-of-range-500.jpg", func(t *testing.T) {
		data, err := Exif("testdata/out-of-range-500.jpg", fs.ImageJPEG, true)

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("all: %+v", data.exif)

		assert.Equal(t, "", data.Artist)
		assert.Equal(t, "2017-04-09T18:33:44Z", data.TakenAt.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "2017-04-09T18:33:44Z", data.TakenAtLocal.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "", data.Title)
		assert.Equal(t, "", data.Keywords.String())
		assert.Equal(t, "", data.Description)
		assert.Equal(t, "", data.Copyright)
		assert.Equal(t, 2448, data.Height)
		assert.Equal(t, 3264, data.Width)
		assert.Equal(t, 0.0, data.Lat)
		assert.Equal(t, 0.0, data.Lng)
		assert.Equal(t, 0.0, data.Altitude)
		assert.Equal(t, "1/387", data.Exposure)
		assert.Equal(t, "Apple", data.CameraMake)
		assert.Equal(t, "iPhone 5s", data.CameraModel)
		assert.Equal(t, "", data.CameraOwner)
		assert.Equal(t, "", data.CameraSerial)
		assert.Equal(t, 29, data.FocalLength)
		assert.Equal(t, 3, data.Orientation)
		assert.Equal(t, "", data.Projection)
		assert.Equal(t, "", data.ColorProfile)
	})

	t.Run("digikam.jpg", func(t *testing.T) {
		data, err := Exif("testdata/digikam.jpg", fs.ImageJPEG, true)

		if err != nil {
			t.Fatal(err)
		}

		//  t.Logf("all: %+v", data.exif)

		assert.Equal(t, "", data.Codec)
		assert.Equal(t, "", data.Artist)
		assert.Equal(t, "2020-10-17T15:48:24Z", data.TakenAt.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "2020-10-17T17:48:24Z", data.TakenAtLocal.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "Europe/Berlin", data.TimeZone)
		assert.Equal(t, "", data.Title)
		assert.Equal(t, "", data.Keywords.String())
		assert.Equal(t, "", data.Description)
		assert.Equal(t, "", data.Copyright)
		assert.Equal(t, 2736, data.Height)
		assert.Equal(t, 3648, data.Width)
		assert.InEpsilon(t, 52.46052, data.Lat, 0.00001)
		assert.InEpsilon(t, 13.331402, data.Lng, 0.00001)
		assert.Equal(t, 84, clean.Altitude(data.Altitude))
		assert.Equal(t, "1/50", data.Exposure)
		assert.Equal(t, "HUAWEI", data.CameraMake)
		assert.Equal(t, "ELE-L29", data.CameraModel)
		assert.Equal(t, "", data.CameraOwner)
		assert.Equal(t, "", data.CameraSerial)
		assert.Equal(t, "", data.LensMake)
		assert.Equal(t, "", data.LensModel)
		assert.Equal(t, 27, data.FocalLength)
		assert.Equal(t, 0, data.Orientation)
		assert.Equal(t, "", data.ColorProfile)
	})

	t.Run("notebook.jpg", func(t *testing.T) {
		data, err := Exif("testdata/notebook.jpg", fs.ImageJPEG, true)

		if err != nil {
			t.Fatal(err)
		}

		//  t.Logf("all: %+v", data.exif)

		assert.Equal(t, 3000, data.Height)
		assert.Equal(t, 4000, data.Width)
		assert.Equal(t, 0.0, data.Lat)
		assert.Equal(t, 0.0, data.Lng)
		assert.Equal(t, 0.0, data.Altitude)
		assert.Equal(t, "1/24", data.Exposure)
		assert.Equal(t, "HMD Global", data.CameraMake)
		assert.Equal(t, "Nokia X71", data.CameraModel)
		assert.Equal(t, 26, data.FocalLength)
		assert.Equal(t, 1, data.Orientation)
	})

	t.Run("snow.jpg", func(t *testing.T) {
		data, err := Exif("testdata/snow.jpg", fs.ImageJPEG, true)

		if err != nil {
			t.Fatal(err)
		}

		//  t.Logf("all: %+v", data.exif)

		assert.Equal(t, 3072, data.Height)
		assert.Equal(t, 4608, data.Width)
		assert.Equal(t, 0.0, data.Lat)
		assert.Equal(t, 0.0, data.Lng)
		assert.Equal(t, 0.0, data.Altitude)
		assert.Equal(t, "1/1600", data.Exposure)
		assert.Equal(t, "OLYMPUS IMAGING CORP.", data.CameraMake)
		assert.Equal(t, "TG-830", data.CameraModel)
		assert.Equal(t, 28, data.FocalLength)
		assert.Equal(t, 1, data.Orientation)
	})

	t.Run("keywords.jpg", func(t *testing.T) {
		data, err := Exif("testdata/keywords.jpg", fs.ImageJPEG, true)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, CodecUnknown, data.Codec)
		assert.Equal(t, "", data.Title)
		assert.Equal(t, "", data.Artist)
		assert.Equal(t, Keywords{"flash"}, data.Keywords)
		assert.Equal(t, "", data.Description)
		assert.Equal(t, "", data.Copyright)
		assert.Equal(t, "Canon", data.CameraMake)
		assert.Equal(t, "Canon EOS 7D", data.CameraModel)
		assert.Equal(t, "", data.LensMake)
		assert.Equal(t, "EF70-200mm f/4L IS USM", data.LensModel)
		assert.Equal(t, 1, data.Orientation)
	})

	t.Run("Iceland-P3.jpg", func(t *testing.T) {
		data, err := Exif("testdata/Iceland-P3.jpg", fs.ImageJPEG, true)

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("all: %+v", data.exif)

		assert.Equal(t, "Nicolas Cornet", data.Artist)
		assert.Equal(t, "2012-08-08T22:07:18Z", data.TakenAt.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "2012-08-08T22:07:18Z", data.TakenAtLocal.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "", data.Title)
		assert.Equal(t, "", data.Keywords.String())
		assert.Equal(t, "", data.Description)
		assert.Equal(t, "Nicolas Cornet", data.Copyright)
		assert.Equal(t, 400, data.Height)
		assert.Equal(t, 600, data.Width)
		assert.InEpsilon(t, 65.05558, data.Lat, 0.00001)
		assert.InEpsilon(t, -16.625702, data.Lng, 0.00001)
		assert.Equal(t, 0.0, data.Altitude)
		assert.Equal(t, "1/8", data.Exposure)
		assert.Equal(t, "NIKON CORPORATION", data.CameraMake)
		assert.Equal(t, "NIKON D800E", data.CameraModel)
		assert.Equal(t, "", data.CameraOwner)
		assert.Equal(t, "", data.CameraSerial)
		assert.Equal(t, 16, data.FocalLength)
		assert.Equal(t, 1, data.Orientation)
		assert.Equal(t, "", data.Projection)
		assert.Equal(t, "", data.ColorProfile)
	})

	t.Run("Iceland-sRGB.jpg", func(t *testing.T) {
		data, err := Exif("testdata/Iceland-sRGB.jpg", fs.ImageJPEG, true)

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("all: %+v", data.exif)

		assert.Equal(t, "Nicolas Cornet", data.Artist)
		assert.Equal(t, "2012-08-08T22:07:18Z", data.TakenAt.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "2012-08-08T22:07:18Z", data.TakenAtLocal.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "", data.Title)
		assert.Equal(t, "", data.Keywords.String())
		assert.Equal(t, "", data.Description)
		assert.Equal(t, "Nicolas Cornet", data.Copyright)
		assert.Equal(t, 400, data.Height)
		assert.Equal(t, 600, data.Width)
		assert.InEpsilon(t, 65.05558, data.Lat, 0.00001)
		assert.InEpsilon(t, -16.625702, data.Lng, 0.00001)
		assert.Equal(t, 0.0, data.Altitude)
		assert.Equal(t, "1/8", data.Exposure)
		assert.Equal(t, "NIKON CORPORATION", data.CameraMake)
		assert.Equal(t, "NIKON D800E", data.CameraModel)
		assert.Equal(t, "", data.CameraOwner)
		assert.Equal(t, "", data.CameraSerial)
		assert.Equal(t, 16, data.FocalLength)
		assert.Equal(t, 1, data.Orientation)
		assert.Equal(t, "", data.Projection)
		assert.Equal(t, "", data.ColorProfile)
	})

	t.Run("animated.gif", func(t *testing.T) {
		_, err := Exif("testdata/animated.gif", fs.ImageGIF, true)

		if err == nil {
			t.Fatal("error expected")
		} else {
			assert.Equal(t, "found no exif data", err.Error())
		}
	})

	t.Run("aurora.jpg", func(t *testing.T) {
		data, err := Exif("testdata/aurora.jpg", fs.ImageJPEG, false)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "2021-10-29 13:42:00 +0000 UTC", data.TakenAtLocal.String())
		assert.Equal(t, "2021-10-29 13:42:00 +0000 UTC", data.TakenAt.String())
		assert.Equal(t, "", data.TimeZone) // Local Time
		assert.Equal(t, 1, data.Orientation)
		assert.Equal(t, 0.0, data.Lat)
		assert.Equal(t, 0.0, data.Lng)
	})

	t.Run("buggy_panorama.jpg", func(t *testing.T) {
		data, err := Exif("testdata/buggy_panorama.jpg", fs.ImageJPEG, false)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "2022-04-24 10:35:53 +0000 UTC", data.TakenAtLocal.String())
		assert.Equal(t, "2022-04-24 02:35:53 +0000 UTC", data.TakenAt.String())
		assert.Equal(t, "Asia/Shanghai", data.TimeZone) // Local Time
		assert.Equal(t, 1, data.Orientation)
		assert.InEpsilon(t, 33.640007, data.Lat, 0.00001)
		assert.InEpsilon(t, 103.48, data.Lng, 0.00001)
		assert.Equal(t, 0.0, data.Altitude)
	})

	t.Run("altitude.jpg", func(t *testing.T) {
		data, err := Exif("testdata/altitude.jpg", fs.ImageJPEG, false)

		if err != nil {
			t.Fatal(err)
		}

		assert.InEpsilon(t, 45.75285, data.Lat, 0.00001)
		assert.InEpsilon(t, 33.221977, data.Lng, 0.00001)
		assert.InEpsilon(t, 4294967284, data.Altitude, 1000)
		assert.Equal(t, 0, clean.Altitude(data.Altitude))
	})
}
