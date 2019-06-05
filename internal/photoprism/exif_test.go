package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestMediaFile_Exif(t *testing.T) {
	conf := config.TestConfig()

	conf.InitializeTestData(t)

	image1, err := NewMediaFile(conf.ImportPath() + "/iphone/IMG_6788.JPG")

	assert.Nil(t, err)

	info, err := image1.Exif()

	assert.Empty(t, err)

	assert.IsType(t, &Exif{}, info)

	assert.Equal(t, "iPhone SE", info.CameraModel)
}

func TestMediaFile_Exif_Slow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	conf := config.TestConfig()

	conf.InitializeTestData(t)

	image2, err := NewMediaFile(conf.ImportPath() + "/raw/IMG_1435.CR2")

	assert.Nil(t, err)

	info, err := image2.Exif()

	assert.Empty(t, err)

	assert.IsType(t, &Exif{}, info)

	assert.Equal(t, "Canon EOS M10", info.CameraModel)
}
