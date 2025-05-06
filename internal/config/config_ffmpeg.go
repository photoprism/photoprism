package config

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// FFmpegBin returns the ffmpeg executable file name.
func (c *Config) FFmpegBin() string {
	return FindBin(c.options.FFmpegBin, encode.FFmpegBin)
}

// FFmpegEnabled checks if FFmpeg is enabled for video transcoding.
func (c *Config) FFmpegEnabled() bool {
	return !c.DisableFFmpeg()
}

// FFmpegEncoder returns the FFmpeg AVC encoder name.
func (c *Config) FFmpegEncoder() encode.Encoder {
	if c.options.FFmpegEncoder == encode.SoftwareAvc.String() {
		return encode.SoftwareAvc
	} else if c.options.FFmpegEncoder == "" {
		return encode.DefaultAvcEncoder()
	}

	return encode.FindEncoder(c.options.FFmpegEncoder)
}

// FFmpegSize returns the maximum ffmpeg video encoding size in pixels (720-7680).
func (c *Config) FFmpegSize() int {
	return thumb.VideoSize(c.options.FFmpegSize).Width
}

// FFmpegQuality returns the ffmpeg encoding quality from 1 to 100,
// with a default of 50 and where 100 is almost lossless.
func (c *Config) FFmpegQuality() int {
	switch {
	case c.options.FFmpegQuality <= 0:
		return encode.DefaultQuality
	case c.options.FFmpegQuality < encode.WorstQuality:
		return encode.WorstQuality
	case c.options.FFmpegQuality > 100:
		return encode.BestQuality
	default:
		return c.options.FFmpegQuality
	}
}

// FFmpegBitrate returns the ffmpeg bitrate limit in Mbps for non-AVC videos to be transcoded
// even if they could be played natively (optional).
func (c *Config) FFmpegBitrate() int {
	switch {
	case c.options.FFmpegBitrate < 0:
		return encode.NoBitrateLimit
	case c.options.FFmpegBitrate == 0:
		return encode.DefaultBitrateLimit
	case c.options.FFmpegBitrate < encode.MinBitrateLimit:
		return encode.MinBitrateLimit
	case c.options.FFmpegBitrate >= encode.MaxBitrateLimit:
		return encode.MaxBitrateLimit
	default:
		return c.options.FFmpegBitrate
	}
}

// FFmpegBitrateExceeded tests if the ffmpeg bitrate limit in Mbps is exceeded.
func (c *Config) FFmpegBitrateExceeded(bitrate float64) bool {
	if bitrate <= 0 {
		return false
	} else if limit := c.FFmpegBitrate(); limit <= 0 {
		return false
	} else {
		return bitrate > float64(limit)
	}
}

// FFmpegPreset returns the ffmpeg encoding preset from "ultrafast" to "veryslow",
// see https://trac.ffmpeg.org/wiki/Encode/H.264#Preset.
func (c *Config) FFmpegPreset() string {
	if c.options.FFmpegPreset == "" {
		return encode.PresetFast
	}

	return c.options.FFmpegPreset
}

// FFmpegDevice returns the ffmpeg device path for supported hardware encoders (optional).
func (c *Config) FFmpegDevice() string {
	if c.options.FFmpegDevice == "" {
		return ""
	} else if txt.IsUInt(c.options.FFmpegDevice) {
		return c.options.FFmpegDevice
	} else if fs.DeviceExists(c.options.FFmpegDevice) {
		return c.options.FFmpegDevice
	}

	return ""
}

// FFmpegMapVideo returns the video streams to be transcoded as string.
func (c *Config) FFmpegMapVideo() string {
	if c.options.FFmpegMapVideo == "" {
		return encode.DefaultMapVideo
	}

	return c.options.FFmpegMapVideo
}

// FFmpegMapAudio returns the audio streams to be transcoded as string.
func (c *Config) FFmpegMapAudio() string {
	if c.options.FFmpegMapAudio == "" {
		return encode.DefaultMapAudio
	}

	return c.options.FFmpegMapAudio
}

// FFmpegOptions returns the FFmpeg options to use for video transcoding.
func (c *Config) FFmpegOptions(encoder encode.Encoder, bitrate string) (encode.Options, error) {
	// Get options to transcode other formats with FFmpeg.
	opt := encode.NewVideoOptions(c.FFmpegBin(), encoder, c.FFmpegSize(), c.FFmpegQuality(), c.FFmpegPreset(), c.FFmpegDevice(), c.FFmpegMapVideo(), c.FFmpegMapAudio())

	// Check options and return error if invalid.
	if opt.Bin == "" {
		return opt, fmt.Errorf("ffmpeg is not installed")
	} else if c.DisableFFmpeg() {
		return opt, fmt.Errorf("ffmpeg is disabled")
	} else if bitrate == "" {
		return opt, fmt.Errorf("bitrate must not be empty")
	} else if encoder.String() == "" {
		return opt, fmt.Errorf("encoder must not be empty")
	}

	return opt, nil
}
