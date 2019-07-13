package photoprism

import (
	"os"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestNewConverter(t *testing.T) {
	conf := config.TestConfig()

	converter := NewConverter(conf)

	assert.IsType(t, &Converter{}, converter)
}

func TestConverter_ConvertToJpeg(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	conf := config.TestConfig()

	conf.InitializeTestData(t)

	converter := NewConverter(conf)

	jpegFilename := conf.ImportPath() + "/fern_green.jpg"

	assert.Truef(t, util.Exists(jpegFilename), "file does not exist: %s", jpegFilename)

	t.Logf("Testing RAW to JPEG converter with %s", jpegFilename)

	jpegMediaFile, err := NewMediaFile(jpegFilename)

	assert.Nil(t, err)

	imageJpeg, err := converter.ConvertToJpeg(jpegMediaFile)

	assert.Empty(t, err, "ConvertToJpeg() failed")

	infoJpeg, err := imageJpeg.Exif()

	assert.Nilf(t, err, "Exif() failed for "+imageJpeg.Filename())

	if err != nil {
		return
	}

	assert.Equal(t, jpegFilename, imageJpeg.filename)

	assert.False(t, infoJpeg == nil || err != nil, "Could not read EXIF data of JPEG image")

	assert.Equal(t, "Canon EOS 7D", infoJpeg.CameraModel)

	rawFilename := conf.ImportPath() + "/raw/IMG_2567.CR2"

	t.Logf("Testing RAW to JPEG converter with %s", rawFilename)

	rawMediaFile, err := NewMediaFile(rawFilename)

	assert.Nil(t, err)

	imageRaw, _ := converter.ConvertToJpeg(rawMediaFile)

	assert.True(t, util.Exists(conf.ImportPath()+"/raw/IMG_2567.jpg"), "Jpeg file was not found - is Darktable installed?")

	assert.NotEqual(t, rawFilename, imageRaw.filename)

	infoRaw, err := imageRaw.Exif()

	assert.False(t, infoRaw == nil || err != nil, "Could not read EXIF data of RAW image")

	assert.Equal(t, "Canon EOS 6D", infoRaw.CameraModel)
}

func TestConverter_ConvertAll(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	conf := config.TestConfig()

	conf.InitializeTestData(t)

	converter := NewConverter(conf)

	converter.ConvertAll(conf.ImportPath())

	jpegFilename := conf.ImportPath() + "/raw/canon_eos_6d.jpg"

	assert.True(t, util.Exists(jpegFilename), "Jpeg file was not found - is Darktable installed?")

	image, err := NewMediaFile(jpegFilename)

	assert.Nil(t, err)

	assert.Equal(t, jpegFilename, image.filename, "FileName must be the same")

	infoRaw, err := image.Exif()

	assert.False(t, infoRaw == nil || err != nil, "Could not read EXIF data of RAW image")

	assert.Equal(t, "Canon EOS 6D", infoRaw.CameraModel, "Camera model should be Canon EOS M10")

	existingJpegFilename := conf.ImportPath() + "/raw/IMG_2567.jpg"

	oldHash := util.Hash(existingJpegFilename)

	os.Remove(existingJpegFilename)

	converter.ConvertAll(conf.ImportPath())

	newHash := util.Hash(existingJpegFilename)

	assert.True(t, util.Exists(existingJpegFilename), "Jpeg file was not found - is Darktable installed?")

	assert.NotEqual(t, oldHash, newHash, "Fingerprint of old and new JPEG file must not be the same")
}
