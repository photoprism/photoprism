package frame

import (
	"image"
	"math/rand/v2"
)

// RandomPoint returns a random image position within the specified range.
func RandomPoint(xMin, yMin, xMax, yMax int) image.Point {
	if xMin == 0 && yMin == 0 && xMax == 0 && yMax == 0 {
		return image.Pt(0, 0)
	}

	if xMin > xMax {
		xMin = xMax
	}

	xDiff := float64(xMax - xMin)
	x := xMin + int(rand.Float64()*xDiff) //nolint:gosec // visual randomness only

	if yMin > yMax {
		yMin = yMax
	}

	yDiff := float64(yMax - yMin)
	y := yMin + int(rand.Float64()*yDiff) //nolint:gosec // visual randomness only

	return image.Pt(x, y)
}
