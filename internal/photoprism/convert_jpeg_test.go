package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/stretchr/testify/assert"
)

func TestConvert_ToJpeg(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	conf := config.TestConfig()
	conf.InitializeTestData()
	convert := NewConvert(conf)

	t.Run("Video", func(t *testing.T) {
		fileName := filepath.Join(conf.ExamplesPath(), "gopher-video.mp4")
		outputName := filepath.Join(conf.SidecarPath(), conf.ExamplesPath(), "gopher-video.mp4.jpg")

		_ = os.Remove(outputName)

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		jpegFile, err := convert.ToJpeg(mf, false)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, jpegFile.FileName(), outputName)
		assert.Truef(t, fs.FileExists(jpegFile.FileName()), "output file does not exist: %s", jpegFile.FileName())

		t.Logf("video metadata: %+v", jpegFile.MetaData())

		_ = os.Remove(outputName)
	})

	t.Run("Raw", func(t *testing.T) {
		jpegFilename := filepath.Join(conf.ImportPath(), "fern_green.jpg")

		assert.Truef(t, fs.FileExists(jpegFilename), "file does not exist: %s", jpegFilename)

		t.Logf("Testing RAW to JPEG convert with %s", jpegFilename)

		mf, err := NewMediaFile(jpegFilename)

		if err != nil {
			t.Fatal(err)
		}

		imageJpeg, err := convert.ToJpeg(mf, false)

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

		imageRaw, err := convert.ToJpeg(rawMediaFile, false)

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
