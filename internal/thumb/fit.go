package thumb

import "image"

// FitSizes contains "fit" cropped thumbnail sizes from largest to smallest.
// Best for the viewer as proportional resizing maintains the aspect ratio.
var FitSizes = SizeList{
	Sizes[Fit7680],
	Sizes[Fit4096],
	Sizes[Fit3840],
	Sizes[Fit2560],
	Sizes[Fit2048],
	Sizes[Fit1920],
	Sizes[Fit1280],
	Sizes[Fit720],
}

// Fit returns the largest fitting thumbnail size.
func Fit(w, h int) (size Size) {
	j := len(FitSizes) - 1

	for i := j; i >= 0; i-- {
		if size = FitSizes[i]; w <= size.Width && h <= size.Height {
			return size
		}
	}

	return FitSizes[0]
}

// FitBounds returns the largest thumbnail size fitting the rectangle.
func FitBounds(r image.Rectangle) (s Size) {
	return Fit(r.Dx(), r.Dy())
}
