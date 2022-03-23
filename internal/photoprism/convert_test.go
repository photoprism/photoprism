package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/stretchr/testify/assert"
)

func TestNewConvert(t *testing.T) {
	conf := config.TestConfig()

	convert := NewConvert(conf)

	assert.IsType(t, &Convert{}, convert)
}

func TestConvert_ToJpeg(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	conf := config.TestConfig()
	conf.InitializeTestData(t)
	convert := NewConvert(conf)

	t.Run("gopher-video.mp4", func(t *testing.T) {
		fileName := filepath.Join(conf.ExamplesPath(), "gopher-video.mp4")
		outputName := filepath.Join(conf.SidecarPath(), conf.ExamplesPath(), "gopher-video.mp4.jpg")

		_ = os.Remove(outputName)

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		jpegFile, err := convert.ToJpeg(mf)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, jpegFile.FileName(), outputName)
		assert.Truef(t, fs.FileExists(jpegFile.FileName()), "output file does not exist: %s", jpegFile.FileName())

		t.Logf("video metadata: %+v", jpegFile.MetaData())

		_ = os.Remove(outputName)
	})

	t.Run("fern_green.jpg", func(t *testing.T) {
		jpegFilename := filepath.Join(conf.ImportPath(), "fern_green.jpg")

		assert.Truef(t, fs.FileExists(jpegFilename), "file does not exist: %s", jpegFilename)

		t.Logf("Testing RAW to JPEG convert with %s", jpegFilename)

		mf, err := NewMediaFile(jpegFilename)

		if err != nil {
			t.Fatal(err)
		}

		imageJpeg, err := convert.ToJpeg(mf)

		if err != nil {
			t.Fatal(err)
		}

		infoJpeg := imageJpeg.MetaData()

		assert.Equal(t, jpegFilename, imageJpeg.fileName)

		assert.Equal(t, "Canon EOS 7D", infoJpeg.CameraModel)

		rawFilename := filepath.Join(conf.ImportPath(), "raw", "IMG_2567.CR2")
		jpgFilename := filepath.Join(conf.SidecarPath(), conf.ImportPath(), "raw/IMG_2567.CR2.jpg")

		t.Logf("Testing RAW to JPEG convert with %s", rawFilename)

		rawMediaFile, err := NewMediaFile(rawFilename)

		if err != nil {
			t.Fatalf("%s for %s", err.Error(), rawFilename)
		}

		imageRaw, err := convert.ToJpeg(rawMediaFile)

		if err != nil {
			t.Fatalf("%s for %s", err.Error(), rawFilename)
		}

		assert.True(t, fs.FileExists(jpgFilename), "Jpeg file was not found - is Darktable installed?")

		if imageRaw == nil {
			t.Fatal("imageRaw is nil")
		}

		assert.NotEqual(t, rawFilename, imageRaw.fileName)

		infoRaw := imageRaw.MetaData()

		assert.Equal(t, "Canon EOS 6D", infoRaw.CameraModel)

		_ = os.Remove(jpgFilename)
	})
}

func TestConvert_ToJson(t *testing.T) {
	conf := config.TestConfig()
	convert := NewConvert(conf)

	t.Run("gopher-video.mp4", func(t *testing.T) {
		fileName := filepath.Join(conf.ExamplesPath(), "gopher-video.mp4")

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		jsonName, err := convert.ToJson(mf)

		if err != nil {
			t.Fatal(err)
		}

		if jsonName == "" {
			t.Fatal("json file name should not be empty")
		}

		assert.FileExists(t, jsonName)

		_ = os.Remove(jsonName)
	})

	t.Run("IMG_4120.JPG", func(t *testing.T) {
		fileName := filepath.Join(conf.ExamplesPath(), "IMG_4120.JPG")
		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		jsonName, err := convert.ToJson(mf)

		if err != nil {
			t.Fatal(err)
		}

		if jsonName == "" {
			t.Fatal("json file name should not be empty")
		}

		assert.FileExists(t, jsonName)

		_ = os.Remove(jsonName)
	})

	t.Run("iphone_7.heic", func(t *testing.T) {
		fileName := conf.ExamplesPath() + "/iphone_7.heic"

		assert.True(t, fs.FileExists(fileName))

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		jsonName, err := convert.ToJson(mf)

		if err != nil {
			t.Fatal(err)
		}

		if jsonName == "" {
			t.Fatal("json file name should not be empty")
		}

		assert.FileExists(t, jsonName)

		_ = os.Remove(jsonName)
	})
}

func TestConvert_Start(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	conf := config.TestConfig()

	conf.InitializeTestData(t)

	convert := NewConvert(conf)

	err := convert.Start(conf.ImportPath())

	if err != nil {
		t.Fatal(err)
	}

	jpegFilename := filepath.Join(conf.SidecarPath(), conf.ImportPath(), "raw/canon_eos_6d.dng.jpg")

	assert.True(t, fs.FileExists(jpegFilename), "Jpeg file was not found - is Darktable installed?")

	image, err := NewMediaFile(jpegFilename)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, jpegFilename, image.fileName, "FileName must be the same")

	infoRaw := image.MetaData()

	assert.Equal(t, "Canon EOS 6D", infoRaw.CameraModel, "UpdateCamera model should be Canon EOS M10")

	existingJpegFilename := filepath.Join(conf.SidecarPath(), conf.ImportPath(), "/raw/IMG_2567.CR2.jpg")

	oldHash := fs.Hash(existingJpegFilename)

	_ = os.Remove(existingJpegFilename)

	if err := convert.Start(conf.ImportPath()); err != nil {
		t.Fatal(err)
	}

	newHash := fs.Hash(existingJpegFilename)

	assert.True(t, fs.FileExists(existingJpegFilename), "Jpeg file was not found - is Darktable installed?")

	assert.NotEqual(t, oldHash, newHash, "Fingerprint of old and new JPEG file must not be the same")
}

func TestConvert_AvcBitrate(t *testing.T) {
	conf := config.TestConfig()
	convert := NewConvert(conf)

	t.Run("low", func(t *testing.T) {
		fileName := filepath.Join(conf.ExamplesPath(), "gopher-video.mp4")

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "1M", convert.AvcBitrate(mf))
	})

	t.Run("medium", func(t *testing.T) {
		fileName := filepath.Join(conf.ExamplesPath(), "gopher-video.mp4")

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		mf.width = 1280
		mf.height = 1024

		assert.Equal(t, "16M", convert.AvcBitrate(mf))
	})

	t.Run("high", func(t *testing.T) {
		fileName := filepath.Join(conf.ExamplesPath(), "gopher-video.mp4")

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		mf.width = 1920
		mf.height = 1080

		assert.Equal(t, "25M", convert.AvcBitrate(mf))
	})

	t.Run("very_high", func(t *testing.T) {
		fileName := filepath.Join(conf.ExamplesPath(), "gopher-video.mp4")

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		mf.width = 4096
		mf.height = 2160

		assert.Equal(t, "50M", convert.AvcBitrate(mf))
	})
}

func TestConvert_AvcConvertCommand(t *testing.T) {
	conf := config.TestConfig()
	convert := NewConvert(conf)

	t.Run(".mp4", func(t *testing.T) {
		fileName := filepath.Join(conf.ExamplesPath(), "gopher-video.mp4")
		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		r, _, err := convert.AvcConvertCommand(mf, "avc1", "")

		if err != nil {
			t.Fatal(err)
		}
		assert.Contains(t, r.Path, "ffmpeg")
		assert.Contains(t, r.Args, "mp4")
	})
	t.Run(".jpg", func(t *testing.T) {
		fileName := filepath.Join(conf.ExamplesPath(), "cat_black.jpg")
		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		r, _, err := convert.AvcConvertCommand(mf, "avc1", "")
		assert.Error(t, err)
		assert.Nil(t, r)
	})
}
