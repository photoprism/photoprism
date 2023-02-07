package thumb

import "image"

// Fitted contains only "fit" cropped thumbnail sizes from largest to smallest.
// Best for the viewer as proportional resizing maintains the aspect ratio.
var Fitted = []Size{
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
	j := len(Fitted) - 1

	for i := j; i >= 0; i-- {
		if size = Fitted[i]; w <= size.Width && h <= size.Height {
			return size
		}
	}

	return Fitted[0]
}

// FitBounds returns the largest thumbnail size fitting the rectangle.
func FitBounds(r image.Rectangle) (s Size) {
	return Fit(r.Dx(), r.Dy())
}
