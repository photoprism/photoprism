package config

import (
	"testing"

	"github.com/photoprism/photoprism/internal/ffmpeg"

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
