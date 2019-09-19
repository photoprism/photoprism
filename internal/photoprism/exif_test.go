package photoprism

import (
	"os"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestMediaFile_Exif_JPEG(t *testing.T) {
	conf := config.TestConfig()

	img, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")

	assert.Nil(t, err)

	info, err := img.Exif()

	assert.Empty(t, err)

	assert.IsType(t, &Exif{}, info)

	assert.Equal(t, "Canon EOS 6D", info.CameraModel)
	assert.Equal(t, "Africa/Johannesburg", info.TimeZone)
	t.Logf("UTC: %s", info.TakenAt.String())
	t.Logf("Local: %s", info.TakenAtLocal.String())
}

func TestMediaFile_Exif_DNG(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	conf := config.TestConfig()

	img, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.dng")

	assert.Nil(t, err)

	info, err := img.Exif()

	assert.Empty(t, err)

	assert.IsType(t, &Exif{}, info)

	assert.Equal(t, "Canon EOS 6D", info.CameraModel)
}

func TestMediaFile_Exif_HEIF(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	conf := config.TestConfig()

	img, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")

	assert.Nil(t, err)

	info, err := img.Exif()

	assert.IsType(t, &Exif{}, info)

	assert.Nil(t, err)

	converter := NewConverter(conf)

	jpeg, err := converter.ConvertToJpeg(img)

	assert.Nil(t, err)

	jpegInfo, err := jpeg.Exif()

	assert.IsType(t, &Exif{}, jpegInfo)

	assert.Nil(t, err)

	assert.Equal(t, "iPhone 7", jpegInfo.CameraModel)

	if err := os.Remove(conf.ExamplesPath() + "/iphone_7.jpg"); err != nil {
		t.Error(err)
	}
}
