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

	jpegFilename := conf.ImportPath() + "/fern_green.jpg"

	assert.Truef(t, fs.FileExists(jpegFilename), "file does not exist: %s", jpegFilename)

	t.Logf("Testing RAW to JPEG convert with %s", jpegFilename)

	jpegMediaFile, err := NewMediaFile(jpegFilename)

	assert.Nil(t, err)

	imageJpeg, err := convert.ToJpeg(jpegMediaFile)

	assert.Empty(t, err, "ToJpeg() failed")

	infoJpeg, err := imageJpeg.MetaData()

	assert.Nilf(t, err, "UpdateExif() failed for "+imageJpeg.Filename())

	if err != nil {
		t.Fatalf("%s for %s", err.Error(), imageJpeg.Filename())
	}

	assert.Equal(t, jpegFilename, imageJpeg.filename)

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

	assert.NotEqual(t, rawFilename, imageRaw.filename)

	infoRaw, err := imageRaw.MetaData()

	assert.Equal(t, "Canon EOS 6D", infoRaw.CameraModel)
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

	assert.Equal(t, jpegFilename, image.filename, "FileName must be the same")

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
