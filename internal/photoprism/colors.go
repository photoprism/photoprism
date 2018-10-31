package photoprism

import (
	"image"
	"os"
	"sort"

	"github.com/RobCherry/vibrant"
	"github.com/lucasb-eyer/go-colorful"
	"golang.org/x/image/colornames"
)

func getColorNames(actualColor colorful.Color) (result []string) {
	var maxDistance = 0.22

	for colorName, colorRGBA := range colornames.Map {
		colorColorful, _ := colorful.MakeColor(colorRGBA)
		currentDistance := colorColorful.DistanceRgb(actualColor)

		if maxDistance >= currentDistance {
			result = append(result, colorName)
		}
	}

	return result
}

func (m *MediaFile) GetColors() (colors []string, vibrantHex string, mutedHex string) {
	file, _ := os.Open(m.filename)

	defer file.Close()

	decodedImage, _, _ := image.Decode(file)
	palette := vibrant.NewPaletteBuilder(decodedImage).Generate()

	if vibrantSwatch := palette.VibrantSwatch(); vibrantSwatch != nil {
		color, _ := colorful.MakeColor(vibrantSwatch.Color())
		colors = append(colors, getColorNames(color)...)
		vibrantHex = color.Hex()
	}

	if mutedSwatch := palette.MutedSwatch(); mutedSwatch != nil {
		color, _ := colorful.MakeColor(mutedSwatch.Color())
		colors = append(colors, getColorNames(color)...)
		mutedHex = color.Hex()
	}

	sort.Strings(colors)

	return colors, vibrantHex, mutedHex
}
