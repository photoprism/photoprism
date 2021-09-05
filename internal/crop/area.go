package crop

import (
	"fmt"
	"image"
)

// Areas represents a list of relative crop areas.
type Areas []Area

// Area represents a relative crop area.
type Area struct {
	Name string  `json:"name,omitempty"`
	X    float32 `json:"x,omitempty"`
	Y    float32 `json:"y,omitempty"`
	W    float32 `json:"w,omitempty"`
	H    float32 `json:"h,omitempty"`
}

// String returns a string identifying the crop area.
func (a Area) String() string {
	return fmt.Sprintf("%03x%03x%03x%03x", int(a.X*1000), int(a.Y*1000), int(a.W*1000), int(a.H*1000))
}

// Bounds returns absolute coordinates and dimension.
func (a Area) Bounds(img image.Image) (min, max image.Point, dim int) {
	size := img.Bounds().Max

	min = image.Point{X: int(float32(size.X) * a.X), Y: int(float32(size.Y) * a.Y)}
	max = image.Point{X: int(float32(size.X) * (a.X + a.W)), Y: int(float32(size.Y) * (a.Y + a.H))}
	dim = int(float32(size.X) * a.W)

	return min, max, dim
}

// clipVal ensures the relative size is within a valid range.
func clipVal(f float32) float32 {
	if f > 1 {
		f = 1
	} else if f < 0 {
		f = 0
	}

	return f
}

// NewArea returns new relative image area.
func NewArea(name string, x, y, w, h float32) Area {
	return Area{
		Name: name,
		X:    clipVal(x),
		Y:    clipVal(y),
		W:    clipVal(w),
		H:    clipVal(h),
	}
}
