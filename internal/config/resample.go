package config

import (
	"strings"

	"github.com/photoprism/photoprism/internal/thumb"
)

// JpegQuality returns the thumbnail jpeg quality setting (25-100).
func (c *Config) JpegQuality() int {
	if c.params.JpegQuality > 100 {
		return 100
	}

	if c.params.JpegQuality < 25 {
		return 25
	}

	return c.params.JpegQuality
}

// Size returns the pre-rendered thumbnail size limit in pixels (720-3840).
func (c *Config) ResampleSize() int {
	if c.params.ResampleSize > 3840 {
		return 3840
	}

	if c.params.ResampleSize < 720 {
		return 720
	}

	return c.params.ResampleSize
}

// Limit returns the on-demand thumbnail size limit in pixels (720-3840).
func (c *Config) ResampleLimit() int {
	if c.params.ResampleLimit > 3840 || c.params.ResampleLimit < 720 || c.ResampleSize() > c.params.ResampleLimit {
		return c.ResampleSize()
	}

	return c.params.ResampleLimit
}

// ResampleFilter returns the thumbnail resample filter (blackman, lanczos, cubic or linear).
func (c *Config) ResampleFilter() thumb.ResampleFilter {
	switch strings.ToLower(c.params.ResampleFilter) {
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

// ResampleUncached returns true for on-demand rendering of uncached thumbnails (high memory and cpu usage).
func (c *Config) ResampleUncached() bool {
	return c.params.ResampleUncached
}
