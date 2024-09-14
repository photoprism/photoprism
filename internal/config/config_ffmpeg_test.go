package config

import (
	"testing"

	"github.com/photoprism/photoprism/internal/ffmpeg"
	"github.com/photoprism/photoprism/internal/thumb"

	"github.com/stretchr/testify/assert"
)

func TestConfig_FFmpegEncoder(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, ffmpeg.SoftwareEncoder, c.FFmpegEncoder())
	c.options.FFmpegEncoder = "nvidia"
	assert.Equal(t, ffmpeg.NvidiaEncoder, c.FFmpegEncoder())
	c.options.FFmpegEncoder = "intel"
	assert.Equal(t, ffmpeg.IntelEncoder, c.FFmpegEncoder())
	c.options.FFmpegEncoder = "xxx"
	assert.Equal(t, ffmpeg.SoftwareEncoder, c.FFmpegEncoder())
	c.options.FFmpegEncoder = ""
	assert.Equal(t, ffmpeg.SoftwareEncoder, c.FFmpegEncoder())
}

func TestConfig_FFmpegEnabled(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, true, c.FFmpegEnabled())

	c.options.DisableFFmpeg = true
	assert.Equal(t, false, c.FFmpegEnabled())
}

func TestConfig_FFmpegBitrate(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, 50, c.FFmpegBitrate())

	c.options.FFmpegBitrate = 1000
	assert.Equal(t, 960, c.FFmpegBitrate())

	c.options.FFmpegBitrate = -5
	assert.Equal(t, 50, c.FFmpegBitrate())

	c.options.FFmpegBitrate = 800
	assert.Equal(t, 800, c.FFmpegBitrate())
}

func TestConfig_FFmpegSize(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, 4096, c.FFmpegSize())

	c.options.FFmpegSize = 0
	assert.Equal(t, 4096, c.FFmpegSize())

	c.options.FFmpegSize = -1
	assert.Equal(t, 7680, c.FFmpegSize())

	c.options.FFmpegSize = 10
	assert.Equal(t, 720, c.FFmpegSize())

	c.options.FFmpegSize = 720
	assert.Equal(t, 720, c.FFmpegSize())

	c.options.FFmpegSize = 1920
	assert.Equal(t, 1920, c.FFmpegSize())

	c.options.FFmpegSize = 4000
	assert.Equal(t, 3840, c.FFmpegSize())

	c.options.FFmpegSize = 8640
	assert.Equal(t, thumb.Sizes[thumb.Fit7680].Width, c.FFmpegSize())
}

func TestConfig_FFmpegBitrateExceeded(t *testing.T) {
	c := NewConfig(CliTestContext())
	c.options.FFmpegBitrate = 0
	assert.False(t, c.FFmpegBitrateExceeded(0.95))
	assert.False(t, c.FFmpegBitrateExceeded(1.05))
	assert.False(t, c.FFmpegBitrateExceeded(2.05))
	c.options.FFmpegBitrate = 1
	assert.False(t, c.FFmpegBitrateExceeded(0.95))
	assert.False(t, c.FFmpegBitrateExceeded(1.0))
	assert.True(t, c.FFmpegBitrateExceeded(1.05))
	assert.True(t, c.FFmpegBitrateExceeded(2.05))
	c.options.FFmpegBitrate = 50
	assert.False(t, c.FFmpegBitrateExceeded(0.95))
	assert.False(t, c.FFmpegBitrateExceeded(1.05))
	assert.False(t, c.FFmpegBitrateExceeded(2.05))
	c.options.FFmpegBitrate = -5
	assert.False(t, c.FFmpegBitrateExceeded(0.95))
	assert.False(t, c.FFmpegBitrateExceeded(1.05))
	assert.False(t, c.FFmpegBitrateExceeded(2.05))
}

func TestConfig_FFmpegMapVideo(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, ffmpeg.MapVideoDefault, c.FFmpegMapVideo())
}

func TestConfig_FFmpegMapAudio(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, ffmpeg.MapAudioDefault, c.FFmpegMapAudio())
}

func TestConfig_FFmpegOptions(t *testing.T) {
	c := NewConfig(CliTestContext())
	bitrate := "25M"
	opt, err := c.FFmpegOptions(ffmpeg.SoftwareEncoder, bitrate)
	assert.NoError(t, err)
	assert.Equal(t, c.FFmpegBin(), opt.Bin)
	assert.Equal(t, ffmpeg.SoftwareEncoder, opt.Encoder)
	assert.Equal(t, bitrate, opt.Bitrate)
	assert.Equal(t, ffmpeg.MapVideoDefault, opt.MapVideo)
	assert.Equal(t, ffmpeg.MapAudioDefault, opt.MapAudio)
	assert.Equal(t, c.FFmpegMapVideo(), opt.MapVideo)
	assert.Equal(t, c.FFmpegMapAudio(), opt.MapAudio)
}
