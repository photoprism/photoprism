package frame

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

// cloneImage normalizes an image to an origin-based NRGBA image.
func cloneImage(img image.Image) *image.NRGBA {
	bounds := img.Bounds()
	clone := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(clone, clone.Bounds(), img, bounds.Min, draw.Src)
	return clone
}

// newCanvas allocates a solid-color NRGBA canvas.
func newCanvas(width, height int, fill color.Color) *image.NRGBA {
	canvas := image.NewNRGBA(image.Rect(0, 0, width, height))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: fill}, image.Point{}, draw.Src)
	return canvas
}

// overlayImage composites one image over another at the requested position.
func overlayImage(base image.Image, overlay image.Image, position image.Point) image.Image {
	dst := cloneImage(base)
	rect := image.Rectangle{Min: position, Max: position.Add(overlay.Bounds().Size())}
	draw.Draw(dst, rect, overlay, overlay.Bounds().Min, draw.Over)
	return dst
}

// rotateImage rotates an image by an arbitrary angle using bilinear sampling.
func rotateImage(img image.Image, angle float64, background color.NRGBA) image.Image {
	if angle == 0 {
		return cloneImage(img)
	}

	src := cloneImage(img)
	width := src.Bounds().Dx()
	height := src.Bounds().Dy()

	radians := angle * math.Pi / 180
	sin, cos := math.Sincos(radians)
	dstWidth := int(math.Ceil(math.Abs(float64(width)*cos) + math.Abs(float64(height)*sin)))
	dstHeight := int(math.Ceil(math.Abs(float64(width)*sin) + math.Abs(float64(height)*cos)))
	dst := newCanvas(dstWidth, dstHeight, background)

	srcCX := float64(width-1) / 2
	srcCY := float64(height-1) / 2
	dstCX := float64(dstWidth-1) / 2
	dstCY := float64(dstHeight-1) / 2

	for y := 0; y < dstHeight; y++ {
		for x := 0; x < dstWidth; x++ {
			dx := float64(x) - dstCX
			dy := float64(y) - dstCY
			srcX := cos*dx + sin*dy + srcCX
			srcY := -sin*dx + cos*dy + srcCY

			if srcX < 0 || srcX > float64(width-1) || srcY < 0 || srcY > float64(height-1) {
				continue
			}

			dst.SetNRGBA(x, y, sampleNRGBA(src, srcX, srcY))
		}
	}

	return dst
}

// sampleNRGBA bilinearly samples a pixel from an NRGBA image.
func sampleNRGBA(src *image.NRGBA, x, y float64) color.NRGBA {
	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	x1 := minInt(x0+1, src.Bounds().Dx()-1)
	y1 := minInt(y0+1, src.Bounds().Dy()-1)
	tx := x - float64(x0)
	ty := y - float64(y0)

	c00 := src.NRGBAAt(x0, y0)
	c10 := src.NRGBAAt(x1, y0)
	c01 := src.NRGBAAt(x0, y1)
	c11 := src.NRGBAAt(x1, y1)

	return color.NRGBA{
		R: interpolateChannel(c00.R, c10.R, c01.R, c11.R, tx, ty),
		G: interpolateChannel(c00.G, c10.G, c01.G, c11.G, tx, ty),
		B: interpolateChannel(c00.B, c10.B, c01.B, c11.B, tx, ty),
		A: interpolateChannel(c00.A, c10.A, c01.A, c11.A, tx, ty),
	}
}

// interpolateChannel bilinearly interpolates a single 8-bit channel.
func interpolateChannel(c00, c10, c01, c11 uint8, tx, ty float64) uint8 {
	top := (1-tx)*float64(c00) + tx*float64(c10)
	bottom := (1-tx)*float64(c01) + tx*float64(c11)
	value := (1-ty)*top + ty*bottom
	return uint8(math.Round(value))
}

// minInt returns the smaller of two ints.
func minInt(a, b int) int {
	if a < b {
		return a
	}

	return b
}
