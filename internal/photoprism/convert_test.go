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

func TestConvert_Start(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	c := config.TestConfig()

	c.InitializeTestData()

	convert := NewConvert(c)

	err := convert.Start(c.ImportPath(), nil, false)

	if err != nil {
		t.Fatal(err)
	}

	jpegFilename := filepath.Join(c.SidecarPath(), c.ImportPath(), "raw/canon_eos_6d.dng.jpg")

	assert.True(t, fs.FileExists(jpegFilename), "Primary file was not found - is Darktable installed?")

	image, err := NewMediaFile(jpegFilename)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, jpegFilename, image.fileName, "FileName must be the same")

	infoRaw := image.MetaData()

	assert.Equal(t, "Canon EOS 6D", infoRaw.CameraModel, "UpdateCamera model should be Canon EOS M10")

	existingJpegFilename := filepath.Join(c.SidecarPath(), c.ImportPath(), "/raw/IMG_2567.CR2.jpg")

	oldHash := fs.Hash(existingJpegFilename)

	_ = os.Remove(existingJpegFilename)

	if err = convert.Start(c.ImportPath(), nil, false); err != nil {
		t.Fatal(err)
	}

	newHash := fs.Hash(existingJpegFilename)

	assert.True(t, fs.FileExists(existingJpegFilename), "Primary file was not found - is Darktable installed?")

	assert.NotEqual(t, oldHash, newHash, "Fingerprint of old and new JPEG file must not be the same")
}
