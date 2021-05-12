package config

// FFmpegBin returns the ffmpeg executable file name.
func (c *Config) FFmpegBin() string {
	return findExecutable(c.options.FFmpegBin, "ffmpeg")
}

// FFmpegEnabled tests if FFmpeg is enabled for video transcoding.
func (c *Config) FFmpegEnabled() bool {
	return !c.DisableFFmpeg()
}

// FFmpegEncoder returns the ffmpeg AVC encoder name.
func (c *Config) FFmpegEncoder() string {
	if c.options.FFmpegEncoder == "" {
		return "libx264"
	}

	return c.options.FFmpegEncoder
}

// FFmpegBuffers returns the number of ffmpeg capture buffers.
func (c *Config) FFmpegBuffers() int {
	if c.options.FFmpegBuffers <= 8 {
		return 8
	}

	if c.options.FFmpegBuffers >= 2048 {
		return 2048
	}

	return c.options.FFmpegBuffers
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
