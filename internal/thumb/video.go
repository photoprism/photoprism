package thumb

// VideoSizes contains all valid video output sizes sorted by size.
var VideoSizes = SizeList{
	Sizes[Fit7680],
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
		return Sizes[Fit7680] // maximum
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
	return Sizes[Fit7680]
}
