package photoprism

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/projection"
	"github.com/photoprism/photoprism/pkg/video"
)

func TestMediaFile_HasSidecarJson(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/beach_wood.jpg")

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, mediaFile.HasSidecarJson())
	})
	t.Run("true", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/blue-go-video.mp4")

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, mediaFile.HasSidecarJson())
	})
	t.Run("true", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/blue-go-video.mp4.json")

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, mediaFile.HasSidecarJson())
	})
}

func TestMediaFile_SidecarJsonName(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/beach_sand.jpg")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "", mediaFile.SidecarJsonName())
	})
	t.Run("true", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/blue-go-video.mp4")

		if err != nil {
			t.Fatal(err)
		}

		assert.Contains(t, mediaFile.SidecarJsonName(), "blue-go-video.mp4.json")
	})
}

func TestMediaFile_NeedsExifToolJson(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/beach_sand.jpg")

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, mediaFile.NeedsExifToolJson())
	})
	t.Run("true", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/blue-go-video.mp4")

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, mediaFile.NeedsExifToolJson())
	})
	t.Run("true", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/blue-go-video.mp4.json")

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, mediaFile.NeedsExifToolJson())
	})
}

func TestMediaFile_CreateExifToolJson(t *testing.T) {
	conf := config.TestConfig()

	t.Run("gopher-video.mp4", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/gopher-video.mp4")

		if err != nil {
			t.Fatal(err)
		}

		jsonName, err := mediaFile.ExifToolJsonName()

		if fs.FileExists(jsonName) {
			if err = os.Remove(jsonName); err != nil {
				t.Error(err)
			}
		}

		assert.True(t, mediaFile.NeedsExifToolJson())

		err = mediaFile.CreateExifToolJson(NewConvert(conf))

		if err != nil {
			t.Fatal(err)
		}

		data := mediaFile.MetaData()

		assert.Empty(t, err)

		assert.IsType(t, meta.Data{}, data)

		assert.Equal(t, "2020-05-11 14:18:35 +0000 UTC", data.TakenAt.String())
		assert.Equal(t, "2020-05-11 14:18:35 +0000 UTC", data.TakenAtLocal.String())
		assert.Equal(t, time.Duration(2410000000), data.Duration)
		assert.Equal(t, meta.CodecAvc1, data.Codec)
		assert.Equal(t, 270, data.Width)
		assert.Equal(t, 480, data.Height)
		assert.Equal(t, false, data.Flash)
		assert.Equal(t, "", data.Description)

		if err = os.Remove(jsonName); err != nil {
			t.Error(err)
		}
	})
}

func TestMediaFile_Exif_JPEG(t *testing.T) {
	conf := config.TestConfig()

	t.Run("elephants.jpg", func(t *testing.T) {
		img, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")

		if err != nil {
			t.Fatal(err)
		}

		data := img.MetaData()

		assert.Empty(t, err)

		assert.IsType(t, meta.Data{}, data)

		assert.Equal(t, "", data.DocumentID)
		assert.Equal(t, "2013-11-26 13:53:55 +0000 UTC", data.TakenAt.String())
		assert.Equal(t, "2013-11-26 15:53:55 +0000 UTC", data.TakenAtLocal.String())
		assert.Equal(t, 1, data.Orientation)
		assert.Equal(t, "Canon EOS 6D", data.CameraModel)
		assert.Equal(t, "Canon", data.CameraMake)
		assert.Equal(t, "EF70-200mm f/4L IS USM", data.LensModel)
		assert.Equal(t, "", data.LensMake)
		assert.Equal(t, "Africa/Johannesburg", data.TimeZone)
		assert.Equal(t, "", data.Artist)
		assert.Equal(t, 111, data.FocalLength)
		assert.Equal(t, "1/640", data.Exposure)
		assert.Equal(t, float32(6.644), data.Aperture)
		assert.Equal(t, float32(10), data.FNumber)
		assert.Equal(t, 200, data.Iso)
		assert.Equal(t, float32(-33.45347), data.Lat)
		assert.Equal(t, float32(25.764645), data.Lng)
		assert.Equal(t, 190.0, data.Altitude)
		assert.Equal(t, 497, data.Width)
		assert.Equal(t, 331, data.Height)
		assert.Equal(t, false, data.Flash)
		assert.Equal(t, "", data.Description)
		t.Logf("UTC: %s", data.TakenAt.String())
		t.Logf("Local: %s", data.TakenAtLocal.String())
	})

	t.Run("fern_green.jpg", func(t *testing.T) {
		img, err := NewMediaFile(conf.ExamplesPath() + "/fern_green.jpg")

		if err != nil {
			t.Fatal(err)
		}

		info := img.MetaData()

		assert.Empty(t, err)

		assert.IsType(t, meta.Data{}, info)

		assert.Equal(t, "", info.DocumentID)
		assert.Equal(t, 1, info.Orientation)
		assert.Equal(t, "Canon EOS 7D", info.CameraModel)
		assert.Equal(t, "Canon", info.CameraMake)
		assert.Equal(t, "EF100mm f/2.8L Macro IS USM", info.LensModel)
		assert.Equal(t, "", info.LensMake)
		assert.Equal(t, "", info.TimeZone)
		assert.Equal(t, "", info.Artist)
		assert.Equal(t, 100, info.FocalLength)
		assert.Equal(t, "1/250", info.Exposure)
		assert.Equal(t, float32(6.644), info.Aperture)
		assert.Equal(t, float32(10), info.FNumber)
		assert.Equal(t, 200, info.Iso)
		assert.Equal(t, 0.0, info.Altitude)
		assert.Equal(t, 331, info.Width)
		assert.Equal(t, 331, info.Height)
		assert.Equal(t, true, info.Flash)
		assert.Equal(t, "", info.Description)
		t.Logf("UTC: %s", info.TakenAt.String())
		t.Logf("Local: %s", info.TakenAtLocal.String())
	})

	t.Run("blue-go-video.mp4", func(t *testing.T) {
		img, err := NewMediaFile(conf.ExamplesPath() + "/blue-go-video.mp4")

		if err != nil {
			t.Fatal(err)
		}

		info := img.MetaData()

		assert.Empty(t, err)

		assert.IsType(t, meta.Data{}, info)
	})

	t.Run("panorama360.jpg", func(t *testing.T) {
		img, err := NewMediaFile("testdata/panorama360.jpg")

		if err != nil {
			t.Fatal(err)
		}

		data := img.MetaData()

		assert.Empty(t, err)

		assert.IsType(t, meta.Data{}, data)

		assert.Equal(t, "", data.Artist)
		assert.Equal(t, "2020-05-24T08:55:21Z", data.TakenAt.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "2020-05-24T11:55:21Z", data.TakenAtLocal.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "", data.Title)
		assert.Equal(t, "panorama", data.Keywords.String())
		assert.Equal(t, "", data.Description)
		assert.Equal(t, "", data.Copyright)
		assert.Equal(t, 3600, data.Height)
		assert.Equal(t, 7200, data.Width)
		assert.Equal(t, float32(59.84083), data.Lat)
		assert.Equal(t, float32(30.51), data.Lng)
		assert.Equal(t, 0.0, data.Altitude)
		assert.Equal(t, "1/1250", data.Exposure)
		assert.Equal(t, "SAMSUNG", data.CameraMake)
		assert.Equal(t, "SM-C200", data.CameraModel)
		assert.Equal(t, "", data.CameraOwner)
		assert.Equal(t, "", data.CameraSerial)
		assert.Equal(t, 6, data.FocalLength)
		assert.Equal(t, 1, data.Orientation)
		assert.Equal(t, projection.Equirectangular.String(), data.Projection)
	})

	t.Run("digikam.jpg", func(t *testing.T) {
		img, err := NewMediaFile("testdata/digikam.jpg")

		if err != nil {
			t.Fatal(err)
		}

		data := img.MetaData()

		assert.Empty(t, err)

		assert.IsType(t, meta.Data{}, data)

		assert.Equal(t, "jpeg", data.Codec)
		assert.Equal(t, "", data.Artist)
		assert.Equal(t, "2020-10-17T15:48:24Z", data.TakenAt.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "2020-10-17T17:48:24Z", data.TakenAtLocal.Format("2006-01-02T15:04:05Z"))
		assert.Equal(t, "Europe/Berlin", data.TimeZone)
		assert.Equal(t, "", data.Title)
		assert.Equal(t, "berlin, shop", data.Keywords.String())
		assert.Equal(t, "", data.Description)
		assert.Equal(t, "", data.Copyright)
		assert.Equal(t, 375, data.Height)
		assert.Equal(t, 500, data.Width)
		assert.Equal(t, float32(52.46052), data.Lat)
		assert.Equal(t, float32(13.331402), data.Lng)
		assert.Equal(t, 84.0, data.Altitude)
		assert.Equal(t, "1/50", data.Exposure)
		assert.Equal(t, "HUAWEI", data.CameraMake)
		assert.Equal(t, "ELE-L29", data.CameraModel)
		assert.Equal(t, "", data.CameraOwner)
		assert.Equal(t, "", data.CameraSerial)
		assert.Equal(t, "", data.LensMake)
		assert.Equal(t, "", data.LensModel)
		assert.Equal(t, 27, data.FocalLength)
		assert.Equal(t, 1, int(data.Orientation))
	})
}

func TestMediaFile_Exif_DNG(t *testing.T) {
	c := config.TestConfig()

	img, err := NewMediaFile(c.ExamplesPath() + "/canon_eos_6d.dng")

	assert.Nil(t, err)

	assert.True(t, img.Ok())
	assert.False(t, img.Empty())

	info := img.MetaData()

	assert.Empty(t, err)

	assert.IsType(t, meta.Data{}, info)

	assert.Equal(t, "", info.DocumentID)
	assert.Equal(t, "2019-06-06 07:29:51 +0000 UTC", info.TakenAt.String())
	assert.Equal(t, "2019-06-06 07:29:51 +0000 UTC", info.TakenAtLocal.String())
	assert.Equal(t, 1, info.Orientation)
	assert.Equal(t, "Canon EOS 6D", info.CameraModel)
	assert.Equal(t, "Canon", info.CameraMake)
	assert.Equal(t, "EF24-105mm f/4L IS USM", info.LensModel)
	assert.Equal(t, "", info.Artist)
	assert.Equal(t, 65, info.FocalLength)
	assert.Equal(t, "1/60", info.Exposure)
	assert.Equal(t, float32(4.971), info.Aperture)
	assert.Equal(t, 1000, info.Iso)
	assert.Equal(t, float32(0), info.Lat)
	assert.Equal(t, float32(0), info.Lng)
	assert.Equal(t, 0.0, info.Altitude)
	assert.Equal(t, false, info.Flash)
	assert.Equal(t, "", info.Description)

	// TODO: Unstable results, depending on test order!
	// assert.Equal(t, 1224, info.Width)
	// assert.Equal(t, 816, info.Height)
	t.Logf("canon_eos_6d.dng width x height: %d x %d", info.Width, info.Height)
	// Workaround, remove when fixed:
	assert.NotEmpty(t, info.Width)
	assert.NotEmpty(t, info.Height)
}

func TestMediaFile_Exif_HEIC(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	conf := config.TestConfig()

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
	assert.Equal(t, 6, jpegInfo.Orientation)
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
	assert.Equal(t, float32(34.79745), jpegInfo.Lat)
	assert.Equal(t, float32(134.76463), jpegInfo.Lng)
	assert.Equal(t, 0.0, jpegInfo.Altitude)
	assert.Equal(t, 4032, jpegInfo.Width)
	assert.Equal(t, 3024, jpegInfo.Height)
	assert.Equal(t, false, jpegInfo.Flash)
	assert.Equal(t, "", jpegInfo.Description)

	if err := os.Remove(filepath.Join(conf.SidecarPath(), conf.ExamplesPath(), "iphone_7.heic.jpg")); err != nil {
		t.Error(err)
	}
}

func TestMediaFile_VideoInfo(t *testing.T) {
	c := config.TestConfig()
	t.Run(
		"samsung-motion-photo.jpg", func(t *testing.T) {
			fileName := filepath.Join(c.ExamplesPath(), "samsung-motion-photo.jpg")

			mf, err := NewMediaFile(fileName)
			if err != nil {
				t.Fatal(err)
			}

			info := mf.VideoInfo()

			assert.Equal(t, video.MP4, info.VideoType)
			assert.Equal(t, video.CodecAVC, info.VideoCodec)
			assert.Equal(t, 1440, info.VideoWidth)
			assert.Equal(t, 1080, info.VideoHeight)
			assert.Equal(t, int64(2685814), info.VideoOffset)
			assert.Equal(t, int64(0), info.ThumbOffset)
			assert.Equal(t, "2.933s", info.Duration.String())
			assert.Equal(t, fs.ImageJPEG, info.FileType)
			assert.Equal(t, media.Live, info.MediaType)
		},
	)

	t.Run(
		"beach_sand.jpg", func(t *testing.T) {
			fileName := filepath.Join(conf.ExamplesPath(), "beach_sand.jpg")

			mf, err := NewMediaFile(fileName)
			if err != nil {
				t.Fatal(err)
			}

			info := mf.VideoInfo()

			assert.Equal(t, video.Unknown, info.VideoType)
			assert.Equal(t, video.CodecUnknown, info.VideoCodec)
			assert.Equal(t, 0, info.VideoWidth)
			assert.Equal(t, 0, info.VideoHeight)
			assert.Equal(t, int64(-1), info.VideoOffset)
			assert.Equal(t, int64(-1), info.ThumbOffset)
			assert.Equal(t, time.Duration(0), info.Duration)
			assert.Equal(t, fs.ImageJPEG, info.FileType)
			assert.Equal(t, media.Image, info.MediaType)
		},
	)
}
