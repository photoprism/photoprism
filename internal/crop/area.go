package crop

import (
	"fmt"
	"image"
	"math"
	"strconv"
	"strings"

	"github.com/photoprism/photoprism/pkg/rnd"
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

// Empty tests if the area is empty.
func (a Area) Empty() bool {
	return a.X == 0 && a.Y == 0 && a.W == 0 && a.H == 0
}

// String returns a string identifying the crop area.
func (a Area) String() string {
	if a.Empty() {
		return ""
	}

	return fmt.Sprintf("%03x%03x%03x%03x", int(a.X*1000), int(a.Y*1000), int(a.W*1000), int(a.H*1000))
}

// Thumb returns a string identifying the file and crop area to create a thumb.
func (a Area) Thumb(fileHash string) string {
	if len(fileHash) < 40 {
		return a.String()
	}

	return fmt.Sprintf("%040s-%012s", fileHash, a.String())
}

// Bounds returns absolute coordinates and dimension.
func (a Area) Bounds(img image.Image) (min, max image.Point, dim int) {
	size := img.Bounds().Max

	min = image.Point{X: int(float32(size.X) * a.X), Y: int(float32(size.Y) * a.Y)}
	max = image.Point{X: int(float32(size.X) * (a.X + a.W)), Y: int(float32(size.Y) * (a.Y + a.H))}
	dim = int(float32(size.X) * a.W)

	return min, max, dim
}

// FileWidth returns the ideal file width based on the crop size.
func (a Area) FileWidth(size Size) int {
	return int(float32(size.Width) / a.W)
}

// Top returns the top Y coordinate as float64.
func (a Area) Top() float64 {
	return float64(a.Y)
}

// Left returns the left X coordinate as float64.
func (a Area) Left() float64 {
	return float64(a.X)
}

// Right returns the right X coordinate as float64.
func (a Area) Right() float64 {
	return float64(a.X + a.W)
}

// Bottom returns the bottom Y coordinate as float64.
func (a Area) Bottom() float64 {
	return float64(a.Y + a.H)
}

// Surface returns the surface area.
func (a Area) Surface() float64 {
	return float64(a.W * a.H)
}

// SurfaceRatio returns the surface ratio.
func (a Area) SurfaceRatio(area float64) float64 {
	if area <= 0 {
		return 0
	}

	if s := a.Surface(); s <= 0 {
		return 0
	} else if area > s {
		return s / area
	} else {
		return area / s
	}
}

// Overlap calculates the overlap of two areas.
func (a Area) Overlap(other Area) (x, y float64) {
	x = math.Max(0, math.Min(a.Right(), other.Right())-math.Max(a.Left(), other.Left()))
	y = math.Max(0, math.Min(a.Bottom(), other.Bottom())-math.Max(a.Top(), other.Top()))

	return x, y
}

// OverlapArea calculates the overlap area of two areas.
func (a Area) OverlapArea(other Area) (area float64) {
	x, y := a.Overlap(other)

	return x * y
}

// OverlapPercent calculates the overlap ratio of two areas in percent.
func (a Area) OverlapPercent(other Area) int {
	return int(math.Round(other.SurfaceRatio(a.OverlapArea(other)) * 100))
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

// AreaFromString returns an image area.
func AreaFromString(s string) Area {
	if len(s) != 12 || !rnd.IsHex(s) {
		return Area{}
	}

	x, _ := strconv.ParseInt(s[0:3], 16, 32)
	y, _ := strconv.ParseInt(s[3:6], 16, 32)
	w, _ := strconv.ParseInt(s[6:9], 16, 32)
	h, _ := strconv.ParseInt(s[9:12], 16, 32)

	return NewArea("crop", float32(x)/1000, float32(y)/1000, float32(w)/1000, float32(h)/1000)
}

// IsCroppedThumb tests if the string represents a cropped thumbnail and returns the split position if true.
func IsCroppedThumb(thumb string) int {
	if thumb == "" || len(thumb) < 41 {
		return -1
	}

	if i := strings.IndexRune(thumb, '-'); i >= 40 && i < len(thumb)-1 {
		return i
	}

	return -1
}

// ParseThumb splits a thumbnail string into the crop area and file hash.
func ParseThumb(thumb string) (fileHash, area string) {
	if len(thumb) == 12 {
		return "", thumb
	} else if len(thumb) < 41 {
		return thumb, ""
	}

	s := strings.SplitN(strings.Trim(thumb, "/ -"), "-", 2)

	fileHash = s[0]

	if len(s) < 2 {
		// Do nothing.
	} else if len(s[1]) >= 12 {
		area = s[1]
	}

	return fileHash, area
}
