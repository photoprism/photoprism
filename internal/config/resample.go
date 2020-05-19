package config

import (
	"strings"

	"github.com/photoprism/photoprism/internal/thumb"
)

// JpegHidden returns true if JPEG files should be created in a .photoprism sub directory (hidden).
func (c *Config) JpegHidden() bool {
	return c.params.JpegHidden
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

// ThumbUncached returns true for on-demand rendering of default thumbnails (high memory and cpu usage).
func (c *Config) ThumbUncached() bool {
	return c.params.ThumbUncached
}

// ThumbSize returns the default thumbnail size limit in pixels (720-3840).
func (c *Config) ThumbSize() int {
	if c.params.ThumbSize > 3840 {
		return 3840
	}

	if c.params.ThumbSize < 720 {
		return 720
	}

	return c.params.ThumbSize
}

// ThumbLimit returns the on-demand thumbnail size limit in pixels (720-3840).
func (c *Config) ThumbLimit() int {
	if c.params.ThumbLimit > 3840 || c.params.ThumbLimit < 720 || c.ThumbSize() > c.params.ThumbLimit {
		return c.ThumbSize()
	}

	return c.params.ThumbLimit
}
