package crop

import (
	"fmt"
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

// String returns a string identifying the approximate marker area.
func (m Area) String() string {
	return fmt.Sprintf("%03d%03d%03d%03d", int(m.X*100), int(m.Y*100), int(m.W*100), int(m.H*100))
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
