package config

import (
	"strings"

	"github.com/photoprism/photoprism/internal/thumb"
)

// JpegSize returns the size limit for automatically converted files in `PIXELS` (720-30000).
func (c *Config) JpegSize() int {
	if c.options.JpegSize < 720 {
		return 720
	} else if c.options.JpegSize > 30000 {
		return 30000
	}

	return c.options.JpegSize
}

// JpegQuality returns the jpeg image quality as thumb.Quality (25-100).
func (c *Config) JpegQuality() thumb.Quality {
	return thumb.ParseQuality(c.options.JpegQuality)
}

// ThumbFilter returns the thumbnail resample filter (best to worst: blackman, lanczos, cubic or linear).
func (c *Config) ThumbFilter() thumb.ResampleFilter {
	switch strings.ToLower(c.options.ThumbFilter) {
	case "blackman":
		return thumb.ResampleBlackman
	case "lanczos":
		return thumb.ResampleLanczos
	case "cubic":
		return thumb.ResampleCubic
	case "linear":
		return thumb.ResampleLinear
	default:
		return thumb.ResampleCubic
	}
}

// ThumbPath returns the thumbnails directory.
func (c *Config) ThumbPath() string {
	return c.CachePath() + "/thumbnails"
}

// ThumbUncached checks if on-demand thumbnail rendering is enabled (high memory and cpu usage).
func (c *Config) ThumbUncached() bool {
	return c.options.ThumbUncached
}

// ThumbSizePrecached returns the pre-cached thumbnail size limit in pixels (720-7680).
func (c *Config) ThumbSizePrecached() int {
	size := c.options.ThumbSize

	if size < 720 {
		size = 720 // Mobile, TV
	} else if size > 7680 {
		size = 7680 // 8K Ultra HD
	}

	return size
}

// ThumbSizeUncached returns the on-demand rendering size limit in pixels (720-7680).
func (c *Config) ThumbSizeUncached() int {
	limit := c.options.ThumbSizeUncached

	if limit < 720 {
		limit = 720 // Mobile, TV
	} else if limit > 7680 {
		limit = 7680 // 8K Ultra HD
	}

	if c.ThumbSizePrecached() > limit {
		limit = c.ThumbSizePrecached()
	}

	return limit
}
