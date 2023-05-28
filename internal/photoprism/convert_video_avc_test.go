package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/ffmpeg"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestConvert_ToAvc(t *testing.T) {
	t.Run("gopher-video.mp4", func(t *testing.T) {
		conf := config.TestConfig()
		convert := NewConvert(conf)

		fileName := filepath.Join(conf.ExamplesPath(), "gopher-video.mp4")
		outputName := filepath.Join(conf.SidecarPath(), conf.ExamplesPath(), "gopher-video.mp4.avc")

		_ = os.Remove(outputName)

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		avcFile, err := convert.ToAvc(mf, ffmpeg.SoftwareEncoder, false, false)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, avcFile.FileName(), outputName)
		assert.Truef(t, fs.FileExists(avcFile.FileName()), "output file does not exist: %s", avcFile.FileName())

		t.Logf("video metadata: %+v", avcFile.MetaData())

		_ = os.Remove(outputName)
	})

	t.Run("jpg", func(t *testing.T) {
		conf := config.TestConfig()
		convert := NewConvert(conf)

		fileName := filepath.Join(conf.ExamplesPath(), "cat_black.jpg")
		outputName := filepath.Join(conf.SidecarPath(), conf.ExamplesPath(), "cat_black.jpg.avc")

		_ = os.Remove(outputName)

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		avcFile, err := convert.ToAvc(mf, "", false, false)
		assert.Error(t, err)
		assert.Nil(t, avcFile)
	})
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

	t.Run("MP4", func(t *testing.T) {
		fileName := filepath.Join(conf.ExamplesPath(), "gopher-video.mp4")
		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		r, _, err := convert.AvcConvertCommand(mf, "avc1", ffmpeg.SoftwareEncoder)

		if err != nil {
			t.Fatal(err)
		}

		assert.Contains(t, r.Path, "ffmpeg")
		assert.Contains(t, r.Args, "mp4")
	})
	t.Run("JPEG", func(t *testing.T) {
		fileName := filepath.Join(conf.ExamplesPath(), "cat_black.jpg")
		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		r, useMutex, err := convert.AvcConvertCommand(mf, "avc1", ffmpeg.SoftwareEncoder)

		assert.False(t, useMutex)
		assert.Error(t, err)
		assert.Nil(t, r)
	})
	t.Run("WebP", func(t *testing.T) {
		webpName := "testdata/windows95.webp"
		avcName := "windows95.mp4"
		mf, err := NewMediaFile(webpName)

		if err != nil {
			t.Fatal(err)
		}

		r, useMutex, err := convert.AvcConvertCommand(mf, avcName, ffmpeg.SoftwareEncoder)

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, useMutex)
		assert.Contains(t, r.Path, "convert")
		assert.Contains(t, r.Args, webpName)
		assert.Contains(t, r.Args, avcName)
	})
}
