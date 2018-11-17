// +build slow

package photoprism

import (
	"os"
	"testing"

	"github.com/photoprism/photoprism/internal/fsutil"
	"github.com/photoprism/photoprism/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ConvertToJpeg(t *testing.T) {
	conf := test.NewConfig()

	conf.InitializeTestData(t)

	converter := NewConverter(conf.GetDarktableCli())

	jpegFilename := conf.GetImportPath() + "/iphone/IMG_6788.JPG"

	assert.Truef(t, fsutil.Exists(jpegFilename), "file does not exist: %s", jpegFilename)

	t.Logf("Testing RAW to JPEG converter with %s", jpegFilename)

	jpegMediaFile, err := NewMediaFile(jpegFilename)

	assert.Nil(t, err)

	imageJpeg, err := converter.ConvertToJpeg(jpegMediaFile)

	assert.Empty(t, err, "ConvertToJpeg() failed")

	infoJpeg, err := imageJpeg.GetExifData()

	assert.Emptyf(t, err, "GetExifData() failed")

	assert.Equal(t, jpegFilename, imageJpeg.filename)

	assert.False(t, infoJpeg == nil || err != nil, "Could not read EXIF data of JPEG image")

	assert.Equal(t, "iPhone SE", infoJpeg.CameraModel)

	rawFilemame := conf.GetImportPath() + "/raw/IMG_1435.CR2"

	t.Logf("Testing RAW to JPEG converter with %s", rawFilemame)

	rawMediaFile, err := NewMediaFile(rawFilemame)

	assert.Nil(t, err)

	imageRaw, _ := converter.ConvertToJpeg(rawMediaFile)

	assert.True(t, fsutil.Exists(conf.GetImportPath()+"/raw/IMG_1435.jpg"), "Jpeg file was not found - is Darktable installed?")

	assert.NotEqual(t, rawFilemame, imageRaw.filename)

	infoRaw, err := imageRaw.GetExifData()

	assert.False(t, infoRaw == nil || err != nil, "Could not read EXIF data of RAW image")

	assert.Equal(t, "Canon EOS M10", infoRaw.CameraModel)
}

func TestConverter_ConvertAll(t *testing.T) {
	conf := test.NewConfig()

	conf.InitializeTestData(t)

	converter := NewConverter(conf.GetDarktableCli())

	converter.ConvertAll(conf.GetImportPath())

	jpegFilename := conf.GetImportPath() + "/raw/IMG_1435.jpg"

	assert.True(t, fsutil.Exists(jpegFilename), "Jpeg file was not found - is Darktable installed?")

	image, err := NewMediaFile(jpegFilename)

	assert.Nil(t, err)

	assert.Equal(t, jpegFilename, image.filename, "FileName must be the same")

	infoRaw, err := image.GetExifData()

	assert.False(t, infoRaw == nil || err != nil, "Could not read EXIF data of RAW image")

	assert.Equal(t, "Canon EOS M10", infoRaw.CameraModel, "Camera model should be Canon EOS M10")

	existingJpegFilename := conf.GetImportPath() + "/raw/20140717_154212_1EC48F8489.jpg"

	oldHash := fsutil.Hash(existingJpegFilename)

	os.Remove(existingJpegFilename)

	converter.ConvertAll(conf.GetImportPath())

	newHash := fsutil.Hash(existingJpegFilename)

	assert.True(t, fsutil.Exists(existingJpegFilename), "Jpeg file was not found - is Darktable installed?")

	assert.NotEqual(t, oldHash, newHash, "Fingerprint of old and new JPEG file must not be the same")
}
