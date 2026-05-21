package config

import (
	"testing"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/internal/thumb"

	"github.com/stretchr/testify/assert"
)

func TestConfig_FFmpegEncoder(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, encode.DefaultAvcEncoder(), c.FFmpegEncoder())
	c.options.FFmpegEncoder = "nvidia"
	assert.Equal(t, encode.NvidiaAvc, c.FFmpegEncoder())
	c.options.FFmpegEncoder = "libx264"
	assert.Equal(t, encode.SoftwareAvc, c.FFmpegEncoder())
	c.options.FFmpegEncoder = "intel"
	assert.Equal(t, encode.IntelAvc, c.FFmpegEncoder())
	c.options.FFmpegEncoder = "vulkan"
	assert.Equal(t, encode.VulkanAvc, c.FFmpegEncoder())
	c.options.FFmpegEncoder = "xxx"
	assert.Equal(t, encode.SoftwareAvc, c.FFmpegEncoder())
	c.options.FFmpegEncoder = ""
	assert.Equal(t, encode.DefaultAvcEncoder(), c.FFmpegEncoder())
}

func TestConfig_FFmpegEnabled(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, true, c.FFmpegEnabled())

	c.options.DisableFFmpeg = true
	assert.Equal(t, false, c.FFmpegEnabled())
}

func TestConfig_FFmpegBitrate(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, encode.DefaultBitrateLimit, c.FFmpegBitrate())

	c.options.FFmpegBitrate = 1000
	assert.Equal(t, encode.MaxBitrateLimit, c.FFmpegBitrate())

	c.options.FFmpegBitrate = -5
	assert.Equal(t, encode.NoBitrateLimit, c.FFmpegBitrate())

	c.options.FFmpegBitrate = 1
	assert.Equal(t, encode.MinBitrateLimit, c.FFmpegBitrate())

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

func TestConfig_FFmpegQuality(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, encode.DefaultQuality, c.FFmpegQuality())

	c.options.FFmpegQuality = 103
	assert.Equal(t, encode.BestQuality, c.FFmpegQuality())

	c.options.FFmpegQuality = 1
	assert.Equal(t, 1, c.FFmpegQuality())

	c.options.FFmpegQuality = 0
	assert.Equal(t, encode.DefaultQuality, c.FFmpegQuality())
}

func TestConfig_FFmpegBitrateExceeded(t *testing.T) {
	c := NewConfig(CliTestContext())
	c.options.FFmpegBitrate = 0
	assert.False(t, c.FFmpegBitrateExceeded(0.95))
	assert.False(t, c.FFmpegBitrateExceeded(1.05))
	assert.False(t, c.FFmpegBitrateExceeded(2.05))
	assert.False(t, c.FFmpegBitrateExceeded(-1.02))
	c.options.FFmpegBitrate = 1
	assert.False(t, c.FFmpegBitrateExceeded(0.95))
	assert.False(t, c.FFmpegBitrateExceeded(1.0))
	assert.True(t, c.FFmpegBitrateExceeded(1.05))
	assert.True(t, c.FFmpegBitrateExceeded(6.05))
	c.options.FFmpegBitrate = 50
	assert.False(t, c.FFmpegBitrateExceeded(0.95))
	assert.False(t, c.FFmpegBitrateExceeded(1.05))
	assert.False(t, c.FFmpegBitrateExceeded(2.05))
	c.options.FFmpegBitrate = -5
	assert.False(t, c.FFmpegBitrateExceeded(0.95))
	assert.False(t, c.FFmpegBitrateExceeded(1.05))
	assert.False(t, c.FFmpegBitrateExceeded(2.05))
}

func TestConfig_FFmpegPreset(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, encode.PresetFast, c.FFmpegPreset())

	c.options.FFmpegPreset = "medium"
	assert.Equal(t, encode.PresetMedium, c.FFmpegPreset())

	c.options.FFmpegPreset = "fast"
	assert.Equal(t, encode.PresetFast, c.FFmpegPreset())

}

func TestConfig_FFmpegDevice(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "", c.FFmpegDevice())
	c.options.FFmpegDevice = "0"
	assert.Equal(t, "0", c.FFmpegDevice())
	c.options.FFmpegDevice = ""
	assert.Equal(t, "", c.FFmpegDevice())
}

func TestConfig_FFmpegMapVideo(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, encode.DefaultMapVideo, c.FFmpegMapVideo())
}

func TestConfig_FFmpegMapAudio(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, encode.DefaultMapAudio, c.FFmpegMapAudio())
}

func TestConfig_FFmpegOptions(t *testing.T) {
	c := NewConfig(CliTestContext())
	bitrate := "25M"
	opt, err := c.FFmpegOptions(encode.SoftwareAvc, bitrate)
	assert.NoError(t, err)
	assert.Equal(t, c.FFmpegBin(), opt.Bin)
	assert.Equal(t, encode.SoftwareAvc, opt.Encoder)
	assert.Equal(t, encode.DefaultMapVideo, opt.MapVideo)
	assert.Equal(t, encode.DefaultMapAudio, opt.MapAudio)
	assert.Equal(t, c.FFmpegMapVideo(), opt.MapVideo)
	assert.Equal(t, c.FFmpegMapAudio(), opt.MapAudio)
}
