package config

import "github.com/photoprism/photoprism/internal/ffmpeg"

// FFmpegBin returns the ffmpeg executable file name.
func (c *Config) FFmpegBin() string {
	return findExecutable(c.options.FFmpegBin, "ffmpeg")
}

// FFmpegEnabled checks if FFmpeg is enabled for video transcoding.
func (c *Config) FFmpegEnabled() bool {
	return !c.DisableFFmpeg()
}

// FFmpegEncoder returns the FFmpeg AVC encoder name.
func (c *Config) FFmpegEncoder() ffmpeg.AvcEncoder {
	return ffmpeg.FindEncoder(c.options.FFmpegEncoder)
}

// FFmpegBitrate returns the ffmpeg bitrate limit in MBit/s.
func (c *Config) FFmpegBitrate() int {
	switch {
	case c.options.FFmpegBitrate <= 0:
		return 50
	case c.options.FFmpegBitrate >= 960:
		return 960
	default:
		return c.options.FFmpegBitrate
	}
}
