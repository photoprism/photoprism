package thumb

// VideoSizes contains all valid video output sizes sorted by size.
// The upper end matches the photo "fit" sizes (up to Fit15360) so high-resolution
// 360° footage is not downscaled; the codec/source still bound the actual output,
// since VideoFilter only ever scales down to min(SizeLimit, source). The 4K UHD
// (Fit3840) and 2K (Fit2048) steps are video-specific and kept to avoid changing
// the size a pre-existing ffmpeg-size config snaps to.
var VideoSizes = SizeList{
	Sizes[Fit15360],
	Sizes[Fit7680],
	Sizes[Fit5120],
	Sizes[Fit4096],
	Sizes[Fit3840],
	Sizes[Fit2560],
	Sizes[Fit2048],
	Sizes[Fit1920],
	Sizes[Fit1280],
	Sizes[Fit720],
}

// VideoSize returns the largest video size type for the given width limit.
func VideoSize(limit int) Size {
	switch {
	case limit < 0:
		return Sizes[Fit15360] // maximum
	case limit == 0:
		return Sizes[Fit4096] // default
	case limit <= 720:
		return Sizes[Fit720] // minimum
	}

	// Find match.
	for _, t := range VideoSizes {
		if t.Width <= limit {
			return t
		}
	}

	// Return maximum size.
	return Sizes[Fit15360]
}
