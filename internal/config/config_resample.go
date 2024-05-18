package config

import (
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/clean"
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

// PngSize returns the size limit for automatically converted files in `PIXELS` (720-30000).
func (c *Config) PngSize() int {
	if c.options.PngSize < 720 {
		return 720
	} else if c.options.PngSize > 30000 {
		return 30000
	}

	return c.options.PngSize
}

// JpegQuality returns the jpeg image quality as thumb.Quality (25-100).
func (c *Config) JpegQuality() thumb.Quality {
	return thumb.ParseQuality(c.options.JpegQuality)
}

// ThumbLibrary returns the name of the image processing library to be used for generating thumbnails.
func (c *Config) ThumbLibrary() string {
	switch clean.TypeLowerUnderscore(c.options.ThumbLibrary) {
	case thumb.LibImaging, "", "imagine", "internal":
		return thumb.LibImaging
	default:
		return thumb.LibVips
	}
}

// ThumbColor returns the color space for thumbnails.
func (c *Config) ThumbColor() thumb.ColorSpace {
	return thumb.ParseColor(c.options.ThumbColor, c.ThumbLibrary())
}

// ThumbFilter returns the thumbnail resample filter (best to worst: blackman, lanczos, cubic or linear).
func (c *Config) ThumbFilter() thumb.ResampleFilter {
	return thumb.ParseFilter(c.options.ThumbFilter, c.ThumbLibrary())
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
