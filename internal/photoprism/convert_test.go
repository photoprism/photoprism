package photoprism

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/ffmpeg"
	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media/video"
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

	initErr := c.InitializeTestData()
	assert.NoError(t, initErr)

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

// fakeMediaFile returns a MediaFile with no backing file but a pinned codec
// and filename, so MediaFile.Ok() reports false and MetaData()/FileType()
// return the pinned values without invoking ExifTool or stat'ing the disk.
func fakeMediaFile(codec, fileName string) *MediaFile {
	return &MediaFile{
		fileName: fileName,
		metaData: meta.Data{Codec: codec},
	}
}

func TestConvert_FFmpegAllowed(t *testing.T) {
	cnf := config.TestConfig()
	convert := NewConvert(cnf)

	// NewConvert captures the current ffmpeg.Exclude(). Pin the field directly
	// so the test stays deterministic regardless of what the global is set to.
	convert.ffmpegExclude = video.NewFormats("magicyuv, avi")

	t.Run("NilFile", func(t *testing.T) {
		assert.True(t, convert.FFmpegAllowed(nil))
	})
	t.Run("EmptyCodecAndType", func(t *testing.T) {
		assert.True(t, convert.FFmpegAllowed(fakeMediaFile("", "")))
	})
	t.Run("AllowedCodec", func(t *testing.T) {
		assert.True(t, convert.FFmpegAllowed(fakeMediaFile("avc1", "/tmp/clip.mp4")))
	})
	t.Run("ExcludedCodec", func(t *testing.T) {
		assert.False(t, convert.FFmpegAllowed(fakeMediaFile("magicyuv", "/tmp/clip.mp4")))
	})
	t.Run("ExcludedCodecMixedCase", func(t *testing.T) {
		assert.False(t, convert.FFmpegAllowed(fakeMediaFile("MagicYUV", "/tmp/clip.mp4")))
	})
	t.Run("ExcludedContainerOnly", func(t *testing.T) {
		// Codec is unknown, but the container extension is on the list.
		assert.False(t, convert.FFmpegAllowed(fakeMediaFile("", "/tmp/clip.avi")))
	})
	t.Run("AllowedContainer", func(t *testing.T) {
		assert.True(t, convert.FFmpegAllowed(fakeMediaFile("", "/tmp/clip.mp4")))
	})
	t.Run("ExcludedByVideoInfo", func(t *testing.T) {
		// The metadata codec is unknown, but the built-in video probe identifies it.
		mf := &MediaFile{fileName: "/tmp/clip.mov", videoInfo: video.Info{VideoCodec: video.CodecMagicYUV}}
		assert.False(t, convert.FFmpegAllowed(mf))
	})
	t.Run("EmptyExcludeList", func(t *testing.T) {
		convert.ffmpegExclude = video.NewFormats("")
		assert.True(t, convert.FFmpegAllowed(fakeMediaFile("magicyuv", "/tmp/clip.avi")))
	})
}

func TestNewConvert_CapturesFFmpegExclude(t *testing.T) {
	// Snapshot and restore the package state so other tests aren't affected.
	saved := ffmpeg.Exclude()
	t.Cleanup(func() { ffmpeg.SetExclude(saved) })

	ffmpeg.SetExclude(video.NewFormats("magicyuv"))

	convert := NewConvert(config.TestConfig())
	assert.False(t, convert.FFmpegAllowed(fakeMediaFile("magicyuv", "/tmp/clip.mp4")))
	assert.True(t, convert.FFmpegAllowed(fakeMediaFile("avc1", "/tmp/clip.mp4")))
}

// TestConvert_StartTypeCheck verifies that files whose content does not match their extension are
// not passed to converters.
func TestConvert_StartTypeCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	c := config.NewMinimalTestConfig(t.TempDir())
	oldCfg := Config()
	SetConfig(c)
	t.Cleanup(func() { SetConfig(oldCfg) })

	png, err := os.ReadFile("testdata/photoprism.png")
	require.NoError(t, err)
	require.NoError(t, fs.MkdirAll(c.OriginalsPath()))
	require.NoError(t, os.WriteFile(filepath.Join(c.OriginalsPath(), "image.jpg"), png, fs.ModeFile)) //nolint:gosec // G703: test-owned path

	require.NoError(t, NewConvert(c).Start(c.OriginalsPath(), nil, false))

	var created []string
	for _, dir := range []string{c.OriginalsPath(), c.SidecarPath()} {
		_ = filepath.Walk(dir, func(fileName string, info os.FileInfo, walkErr error) error {
			if walkErr == nil && !info.IsDir() && strings.HasPrefix(filepath.Base(fileName), "image.jpg.") {
				created = append(created, fileName)
			}
			return nil
		})
	}
	assert.Empty(t, created)
}

// TestConvert_StartAvcMetadata verifies that the type check does not keep the ExifTool metadata of a
// video from the converter, so that a video in a playable format is not transcoded.
func TestConvert_StartAvcMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	c := config.NewMinimalTestConfig(t.TempDir())
	oldCfg := Config()
	SetConfig(c)
	t.Cleanup(func() { SetConfig(oldCfg) })

	if !c.ExifToolEnabled() || !c.FFmpegEnabled() {
		t.Skip("ExifTool and FFmpeg must be available")
	}

	mp4, err := os.ReadFile(filepath.Join(c.SamplesPath(), "gopher-video.mp4"))
	require.NoError(t, err)
	require.NoError(t, fs.MkdirAll(c.OriginalsPath()))
	require.NoError(t, os.WriteFile(filepath.Join(c.OriginalsPath(), "clip.mp4"), mp4, fs.ModeFile)) //nolint:gosec // G703: test-owned path

	require.NoError(t, NewConvert(c).Start(c.OriginalsPath(), nil, false))

	var transcoded []string
	_ = filepath.Walk(c.SidecarPath(), func(fileName string, info os.FileInfo, walkErr error) error {
		if walkErr == nil && !info.IsDir() && strings.HasSuffix(fileName, fs.ExtAvc) {
			transcoded = append(transcoded, fileName)
		}
		return nil
	})
	assert.Empty(t, transcoded)
}
