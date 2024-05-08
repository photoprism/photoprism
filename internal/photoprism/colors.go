package photoprism

import (
	"fmt"
	"image/color"
	"math"

	"github.com/lucasb-eyer/go-colorful"

	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/colors"
)

// Colors returns the ColorPerception of an image (only JPEG supported).
func (m *MediaFile) Colors(thumbPath string) (perception colors.ColorPerception, err error) {
	if !m.IsPreviewImage() || m.IsThumb() {
		return perception, fmt.Errorf("%s is not a jpeg", clean.Log(m.BaseName()))
	}

	img, err := m.Resample(thumbPath, thumb.Colors)

	if err != nil {
		log.Debugf("colors: %s in %s (resample)", err, clean.Log(m.BaseName()))
		return perception, err
	}

	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// Enforce thumbnail width limit and warn if it is exceeded.
	if maxWidth := thumb.SizeColors.Width; width > maxWidth {
		log.Warnf("color: thumbnail width %d exceeds size limit of %d in %s", width, maxWidth, clean.Log(m.RootRelName()))
		width = maxWidth
	}

	// Enforce thumbnail height limit and warn if it is exceeded.
	if maxHeight := thumb.SizeColors.Height; height > maxHeight {
		log.Warnf("color: thumbnail height %d exceeds size limit of %d in %s", height, maxHeight, clean.Log(m.RootRelName()))
		height = maxHeight
	}

	pixels := float64(width * height)
	chromaSum := 0.0

	colorCount := make(map[colors.Color]uint16)
	var mainColorCount uint16

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			rgb, _ := colorful.MakeColor(color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)})
			i := colors.Colorful(rgb)
			perception.Colors = append(perception.Colors, i)

			if _, ok := colorCount[i]; ok == true {
				colorCount[i] += colors.Weights[i]
			} else {
				colorCount[i] = colors.Weights[i]
			}

			if colorCount[i] > mainColorCount {
				mainColorCount = colorCount[i]
				perception.MainColor = i
			}

			_, c, l := rgb.Hcl()

			chromaSum += c

			perception.Luminance = append(perception.Luminance, colors.Luminance(math.Round(l*15)))
		}
	}

	perception.Chroma = colors.Chroma(math.Round((chromaSum / pixels) * 100))

	return perception, nil
}
