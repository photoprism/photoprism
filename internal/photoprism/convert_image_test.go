package photoprism

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/stretchr/testify/assert"
)

func TestConvert_ToImage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cnf := config.TestConfig()
	cnf.InitializeTestData()
	convert := NewConvert(cnf)

	t.Run("Video", func(t *testing.T) {
		fileName := filepath.Join(cnf.ExamplesPath(), "gopher-video.mp4")
		outputName := filepath.Join(cnf.SidecarPath(), cnf.ExamplesPath(), "gopher-video.mp4.jpg")

		_ = os.Remove(outputName)

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		jpegFile, err := convert.ToImage(mf, false)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, jpegFile.FileName(), outputName)
		assert.Truef(t, fs.FileExists(jpegFile.FileName()), "output file does not exist: %s", jpegFile.FileName())

		t.Logf("video metadata: %+v", jpegFile.MetaData())

		_ = os.Remove(outputName)
	})

	t.Run("Raw", func(t *testing.T) {
		jpegFilename := filepath.Join(cnf.ImportPath(), "fern_green.jpg")

		assert.Truef(t, fs.FileExists(jpegFilename), "file does not exist: %s", jpegFilename)

		t.Logf("Testing RAW to JPEG convert with %s", jpegFilename)

		mf, err := NewMediaFile(jpegFilename)

		if err != nil {
			t.Fatal(err)
		}

		imageJpeg, err := convert.ToImage(mf, false)

		if err != nil {
			t.Fatal(err)
		}

		infoJpeg := imageJpeg.MetaData()

		assert.Equal(t, jpegFilename, imageJpeg.fileName)

		assert.Equal(t, "Canon EOS 7D", infoJpeg.CameraModel)

		rawFilename := filepath.Join(cnf.ImportPath(), "raw", "IMG_2567.CR2")
		jpgFilename := filepath.Join(cnf.SidecarPath(), cnf.ImportPath(), "raw/IMG_2567.CR2.jpg")

		t.Logf("Testing RAW to JPEG convert with %s", rawFilename)

		rawMediaFile, err := NewMediaFile(rawFilename)

		if err != nil {
			t.Fatalf("%s for %s", err.Error(), rawFilename)
		}

		imageRaw, err := convert.ToImage(rawMediaFile, false)

		if err != nil {
			t.Fatalf("%s for %s", err.Error(), rawFilename)
		}

		assert.True(t, fs.FileExists(jpgFilename), "Primary file was not found - is Darktable installed?")

		if imageRaw == nil {
			t.Fatal("imageRaw is nil")
		}

		assert.NotEqual(t, rawFilename, imageRaw.fileName)

		infoRaw := imageRaw.MetaData()

		assert.Equal(t, "Canon EOS 6D", infoRaw.CameraModel)

		_ = os.Remove(jpgFilename)
	})

	t.Run("Svg", func(t *testing.T) {
		svgFile := fs.Abs("./testdata/agpl.svg")

		mediaFile, err := NewMediaFile(svgFile)

		t.Logf("svg: %s", mediaFile.FileName())

		if err != nil {
			t.Fatal(err)
		}

		imageFile, err := convert.ToImage(mediaFile, false)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("jpeg: %s", imageFile.FileName())

		_ = imageFile.Remove()
	})
}

func TestConvert_PngConvertCommands(t *testing.T) {
	cnf := config.TestConfig()
	convert := NewConvert(cnf)

	t.Run("SVG", func(t *testing.T) {
		svgFile := fs.Abs("./testdata/agpl.svg")
		pngFile := fs.Abs("./testdata/agpl.png")

		mediaFile, err := NewMediaFile(svgFile)

		t.Logf("svg: %s", mediaFile.FileName())

		if err != nil {
			t.Fatal(err)
		}

		cmds, useMutex, err := convert.PngConvertCommands(mediaFile, pngFile)

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, useMutex)

		assert.NotEmpty(t, cmds)
		assert.True(t, strings.Contains(cmds[0].String(), "rsvg"))

		t.Logf("commands: %#v", cmds)
	})
}
