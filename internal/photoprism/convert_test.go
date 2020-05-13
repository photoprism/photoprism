package photoprism

import (
	"os"
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
		fileName := conf.ExamplesPath() + "/gopher-video.mp4"
		outputName := conf.ExamplesPath() + "/gopher-video.jpg"

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

		metaData, err := jpegFile.MetaData()

		if err != nil {
			t.Log(err)
		} else {
			t.Logf("video metadata: %+v", metaData)
		}

		_ = os.Remove(outputName)
	})

	t.Run("fern_green.jpg", func(t *testing.T) {
		jpegFilename := conf.ImportPath() + "/fern_green.jpg"

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

		infoJpeg, err := imageJpeg.MetaData()

		if err != nil {
			t.Fatalf("%s for %s", err.Error(), imageJpeg.FileName())
		}

		assert.Equal(t, jpegFilename, imageJpeg.fileName)

		assert.Equal(t, "Canon EOS 7D", infoJpeg.CameraModel)

		rawFilename := conf.ImportPath() + "/raw/IMG_2567.CR2"

		t.Logf("Testing RAW to JPEG convert with %s", rawFilename)

		rawMediaFile, err := NewMediaFile(rawFilename)

		if err != nil {
			t.Fatalf("%s for %s", err.Error(), rawFilename)
		}

		imageRaw, err := convert.ToJpeg(rawMediaFile)

		if err != nil {
			t.Fatalf("%s for %s", err.Error(), rawFilename)
		}

		assert.True(t, fs.FileExists(conf.ImportPath()+"/raw/IMG_2567.jpg"), "Jpeg file was not found - is Darktable installed?")

		if imageRaw == nil {
			t.Fatal("imageRaw is nil")
		}

		assert.NotEqual(t, rawFilename, imageRaw.fileName)

		infoRaw, err := imageRaw.MetaData()

		assert.Equal(t, "Canon EOS 6D", infoRaw.CameraModel)
	})
}

func TestConvert_ToJson(t *testing.T) {
	conf := config.TestConfig()
	convert := NewConvert(conf)

	t.Run("gopher-video.mp4", func(t *testing.T) {
		fileName := conf.ExamplesPath() + "/gopher-video.mp4"
		outputName := conf.ExamplesPath() + "/gopher-video.json"

		_ = os.Remove(outputName)

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)
		assert.Falsef(t, fs.FileExists(outputName), "output file must not exist: %s", outputName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		jsonFile, err := convert.ToJson(mf)

		if err != nil {
			t.Fatal(err)
		}

		if jsonFile == nil {
			t.Fatal("jsonFile should not be nil")
		}

		assert.Equal(t, jsonFile.FileName(), outputName)
		assert.Truef(t, fs.FileExists(jsonFile.FileName()), "output file does not exist: %s", jsonFile.FileName())
		assert.False(t, jsonFile.IsJpeg())
		assert.False(t, jsonFile.IsMedia())
		assert.False(t, jsonFile.IsVideo())
		assert.True(t, jsonFile.IsSidecar())

		_ = os.Remove(outputName)
	})

	t.Run("iphone_7.heic", func(t *testing.T) {
		fileName := conf.ExamplesPath() + "/iphone_7.heic"
		outputName := conf.ExamplesPath() + "/iphone_7.json"

		assert.True(t, fs.FileExists(fileName))
		assert.True(t, fs.FileExists(outputName))

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		jsonFile, err := convert.ToJson(mf)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, jsonFile.FileName(), outputName)
		assert.Truef(t, fs.FileExists(jsonFile.FileName()), "output file does not exist: %s", jsonFile.FileName())
		assert.False(t, jsonFile.IsJpeg())
		assert.False(t, jsonFile.IsMedia())
		assert.False(t, jsonFile.IsVideo())
		assert.True(t, jsonFile.IsSidecar())
	})
}

func TestConvert_Start(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	conf := config.TestConfig()

	conf.InitializeTestData(t)

	convert := NewConvert(conf)

	convert.Start(conf.ImportPath())

	jpegFilename := conf.ImportPath() + "/raw/canon_eos_6d.jpg"

	assert.True(t, fs.FileExists(jpegFilename), "Jpeg file was not found - is Darktable installed?")

	image, err := NewMediaFile(jpegFilename)

	if err != nil {
		t.Fatal(err.Error())
	}

	assert.Equal(t, jpegFilename, image.fileName, "FileName must be the same")

	infoRaw, err := image.MetaData()

	assert.Equal(t, "Canon EOS 6D", infoRaw.CameraModel, "UpdateCamera model should be Canon EOS M10")

	existingJpegFilename := conf.ImportPath() + "/raw/IMG_2567.jpg"

	oldHash := fs.Hash(existingJpegFilename)

	os.Remove(existingJpegFilename)

	convert.Start(conf.ImportPath())

	newHash := fs.Hash(existingJpegFilename)

	assert.True(t, fs.FileExists(existingJpegFilename), "Jpeg file was not found - is Darktable installed?")

	assert.NotEqual(t, oldHash, newHash, "Fingerprint of old and new JPEG file must not be the same")
}
