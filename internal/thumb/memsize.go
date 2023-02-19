package thumb

import (
	"image"
	"image/color"

	"github.com/dustin/go-humanize"
)

// Byte size factors.
const (
	KB = 1024
	MB = KB * 1024
	GB = MB * 1024
)

// Bytes represents memory usage in bytes.
type Bytes uint64

// KByte returns the size in kilobyte.
func (b Bytes) KByte() float64 {
	return float64(b) / KB
}

// MByte returns the size in megabyte.
func (b Bytes) MByte() float64 {
	return float64(b) / MB
}

// GByte returns the size in gigabyte.
func (b Bytes) GByte() float64 {
	return float64(b) / GB
}

// String returns a human-readable memory usage string.
func (b Bytes) String() string {
	return humanize.Bytes(uint64(b))
}

// MemSize returns the estimated size of the image in memory in bytes.
func MemSize(img image.Image) Bytes {
	r := img.Bounds()

	pixels := r.Dx() * r.Dy()
	bytesPerPixel := 4

	// Image representation in a computer memory:
	// https://medium.com/@oleg.shipitko/what-does-stride-mean-in-image-processing-bba158a72bcd
	switch img.ColorModel() {
	case color.AlphaModel, color.GrayModel:
		bytesPerPixel = 1
	case color.Alpha16Model, color.Gray16Model:
		bytesPerPixel = 2
	case color.RGBAModel, color.NRGBAModel:
		bytesPerPixel = 4
	case color.RGBA64Model, color.NRGBA64Model:
		bytesPerPixel = 8
	}

	return Bytes(pixels * bytesPerPixel)
}
