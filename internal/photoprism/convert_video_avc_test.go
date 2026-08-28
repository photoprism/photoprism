package photoprism

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestConvert_ToAvc(t *testing.T) {
	t.Run("GopherVideoMp4", func(t *testing.T) {
		conf := config.TestConfig()
		convert := NewConvert(conf)

		fileName := filepath.Join(conf.SamplesPath(), "gopher-video.mp4")
		outputName := filepath.Join(conf.SidecarPath(), conf.SamplesPath(), "gopher-video.mp4.avc")

		_ = os.Remove(outputName)

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		avcFile, err := convert.ToAvc(mf, encode.SoftwareAvc, false, false)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, avcFile.FileName(), outputName)
		assert.Truef(t, fs.FileExists(avcFile.FileName()), "output file does not exist: %s", avcFile.FileName())

		t.Logf("video metadata: %+v", avcFile.MetaData())

		oldTime := time.Unix(1, 0)
		assert.NoError(t, os.Chtimes(outputName, oldTime, oldTime))

		avcFile, err = convert.ToAvc(mf, encode.SoftwareAvc, false, true)
		if err != nil {
			t.Fatal(err)
		}

		info, err := os.Stat(avcFile.FileName())
		assert.NoError(t, err)
		assert.True(t, info.ModTime().After(oldTime), "forced conversion should replace the sidecar")

		_ = os.Remove(outputName)
	})
	t.Run("Jpg", func(t *testing.T) {
		conf := config.TestConfig()
		convert := NewConvert(conf)

		fileName := filepath.Join(conf.SamplesPath(), "cat_black.jpg")
		outputName := filepath.Join(conf.SidecarPath(), conf.SamplesPath(), "cat_black.jpg.avc")

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

	t.Run("Low", func(t *testing.T) {
		fileName := filepath.Join(conf.SamplesPath(), "gopher-video.mp4")

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "1M", convert.AvcBitrate(mf))
	})
	t.Run("Medium", func(t *testing.T) {
		fileName := filepath.Join(conf.SamplesPath(), "gopher-video.mp4")

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		mf.width = 1280
		mf.height = 1024

		assert.Equal(t, "16M", convert.AvcBitrate(mf))
	})
	t.Run("High", func(t *testing.T) {
		fileName := filepath.Join(conf.SamplesPath(), "gopher-video.mp4")

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		mf.width = 1920
		mf.height = 1080

		assert.Equal(t, "25M", convert.AvcBitrate(mf))
	})
	t.Run("VeryHigh", func(t *testing.T) {
		fileName := filepath.Join(conf.SamplesPath(), "gopher-video.mp4")

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		mf.width = 4096
		mf.height = 2160

		assert.Equal(t, "60M", convert.AvcBitrate(mf))
	})
}

func TestConvert_TranscodeToAvcCmd(t *testing.T) {
	conf := config.TestConfig()
	convert := NewConvert(conf)

	t.Run("MP4", func(t *testing.T) {
		fileName := filepath.Join(conf.SamplesPath(), "gopher-video.mp4")
		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		r, _, err := convert.TranscodeToAvcCmd(mf, "avc1", encode.SoftwareAvc)

		if err != nil {
			t.Fatal(err)
		}

		assert.Contains(t, r.Path, "ffmpeg")
		assert.Contains(t, r.Args, "mp4")
	})
	t.Run("Jpeg", func(t *testing.T) {
		fileName := filepath.Join(conf.SamplesPath(), "cat_black.jpg")
		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		r, useMutex, err := convert.TranscodeToAvcCmd(mf, "avc1", encode.SoftwareAvc)

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

		r, useMutex, err := convert.TranscodeToAvcCmd(mf, avcName, encode.SoftwareAvc)

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, useMutex)
		assert.Contains(t, r.Path, "convert")
		assert.Contains(t, r.Args, webpName)
		assert.Contains(t, r.Args, avcName)
	})
	t.Run("Insv", func(t *testing.T) {
		mf, err := NewMediaFile("testdata/insta360.insv")
		if err != nil {
			t.Fatal(err)
		}
		// A hardware encoder is requested, but .insv must be forced onto the software v360 path.
		r, _, err := convert.TranscodeToAvcCmd(mf, "insta360.avc", encode.Encoder("intel"))
		if err != nil {
			t.Fatal(err)
		}
		args := strings.Join(r.Args, " ")
		assert.Contains(t, r.Path, "ffmpeg")
		assert.Contains(t, args, "v360=input=dfisheye:output=e")
		assert.NotContains(t, args, "roll=180")
		assert.Contains(t, args, "libx264")
	})
	t.Run("OneRSInsv", func(t *testing.T) {
		mf, err := NewMediaFile(oneRSInsvFixture(t, t.TempDir(), "camera.insv"))
		if err != nil {
			t.Fatal(err)
		}
		r, _, err := convert.TranscodeToAvcCmd(mf, "camera.avc", encode.SoftwareAvc)
		if err != nil {
			t.Fatal(err)
		}
		assert.Contains(t, strings.Join(r.Args, " "), "v360=input=dfisheye:output=e:ih_fov=190:iv_fov=190:roll=180")
	})
	t.Run("OneRSSquareInsv", func(t *testing.T) {
		mf, err := NewMediaFile(oneRSInsvFixture(t, t.TempDir(), "camera.insv"))
		if err != nil {
			t.Fatal(err)
		}
		mf.width = 3072
		mf.height = 3072
		r, _, err := convert.TranscodeToAvcCmd(mf, "camera.avc", encode.SoftwareAvc)
		if err != nil {
			t.Fatal(err)
		}
		// Matched on the filter rather than the bare name: the arguments carry the input
		// path, and a temp directory whose random suffix starts with 360 contains "v360".
		assert.NotContains(t, strings.Join(r.Args, " "), "v360=")
	})
	t.Run("Insta360SeparateLensPair", func(t *testing.T) {
		dir := t.TempDir()
		leftName := writeInsta360CaptureFile(t, dir, "VID_20220625_140410_00_008.insv", "testdata/flash.jpg")
		rightName := writeInsta360CaptureFile(t, dir, "VID_20220625_140410_10_008.insv", "testdata/flash.jpg")
		mf, err := NewMediaFile(leftName)
		if err != nil {
			t.Fatal(err)
		}

		r, useMutex, err := convert.TranscodeToAvcCmd(mf, "camera.avc", encode.Encoder("intel"))
		if err != nil {
			t.Fatal(err)
		}

		args := strings.Join(r.Args, " ")
		assert.True(t, useMutex)
		assert.Contains(t, args, "-i "+leftName+" -i "+rightName)
		assert.Contains(t, args, "hstack=inputs=2:shortest=1,v360=input=dfisheye:output=e")
		assert.Contains(t, args, "-map [v] -map 0:a:0?")
		assert.Contains(t, args, "libx264")
	})
	t.Run("Mp4NoV360", func(t *testing.T) {
		mf, err := NewMediaFile(filepath.Join(conf.SamplesPath(), "gopher-video.mp4"))
		if err != nil {
			t.Fatal(err)
		}
		r, _, err := convert.TranscodeToAvcCmd(mf, "gopher.avc", encode.SoftwareAvc)
		if err != nil {
			t.Fatal(err)
		}
		assert.NotContains(t, strings.Join(r.Args, " "), "v360=")
	})
}
