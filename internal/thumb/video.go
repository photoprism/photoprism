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
	if limit < 0 {
		// Return maximum size.
		return Sizes[Fit7680]
	} else if limit == 0 {
		// Return default size.
		return Sizes[Fit4096]
	} else if limit <= 720 {
		// Return minimum size.
		return Sizes[Fit720]
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
