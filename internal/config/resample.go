package config

import (
	"strings"

	"github.com/photoprism/photoprism/internal/thumb"
)

// JpegSize returns the size limit for automatically converted files in `PIXELS` (720-30000).
func (c *Config) JpegSize() int {
	if c.params.JpegSize < 720 {
		return 720
	} else if c.params.JpegSize > 30000 {
		return 30000
	}

	return c.params.JpegSize
}

// JpegQuality returns the jpeg quality for resampling, use 95 for high-quality thumbs (25-100).
func (c *Config) JpegQuality() int {
	if c.params.JpegQuality > 100 {
		return 100
	}

	if c.params.JpegQuality < 25 {
		return 25
	}

	return c.params.JpegQuality
}

// ThumbFilter returns the thumbnail resample filter (best to worst: blackman, lanczos, cubic or linear).
func (c *Config) ThumbFilter() thumb.ResampleFilter {
	switch strings.ToLower(c.params.ThumbFilter) {
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
	return c.params.ThumbUncached
}

// ThumbSize returns the pre-rendered thumbnail size limit in pixels (720-7680).
func (c *Config) ThumbSize() int {
	size := c.params.ThumbSize

	if size < 720 {
		size = 720 // Mobile, TV
	} else if size > 7680 {
		size = 7680 // 8K Ultra HD
	}

	return size
}

// ThumbSizeUncached returns the on-demand rendering size limit in pixels (720-7680).
func (c *Config) ThumbSizeUncached() int {
	limit := c.params.ThumbSizeUncached

	if limit < 720 {
		limit = 720 // Mobile, TV
	} else if limit > 7680 {
		limit = 7680 // 8K Ultra HD
	}

	if c.ThumbSize() > limit {
		limit = c.ThumbSize()
	}

	return limit
}
