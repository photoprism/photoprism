package photoprism

import (
	"image"
	"image/color"
	"log"
	"os"
	"sort"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/RobCherry/vibrant"
	"github.com/lucasb-eyer/go-colorful"
)

var colorMap = map[string]color.RGBA{
	"red":    {0xf4, 0x43, 0x36, 0xff},
	"pink":   {0xe9, 0x1e, 0x63, 0xff},
	"purple": {0x9c, 0x27, 0xb0, 0xff},
	"indigo": {0x3F, 0x51, 0xB5, 0xff},
	"blue":   {0x21, 0x96, 0xF3, 0xff},
	"cyan":   {0x00, 0xBC, 0xD4, 0xff},
	"teal":   {0x00, 0x96, 0x88, 0xff},
	"green":  {0x4C, 0xAF, 0x50, 0xff},
	"lime":   {0xCD, 0xDC, 0x39, 0xff},
	"yellow": {0xFF, 0xEB, 0x3B, 0xff},
	"amber":  {0xFF, 0xC1, 0x07, 0xff},
	"orange": {0xFF, 0x98, 0x00, 0xff},
	"brown":  {0x79, 0x55, 0x48, 0xff},
	"grey":   {0x9E, 0x9E, 0x9E, 0xff},
	"white":  {0x00, 0x00, 0x00, 0xff},
	"black":  {0xFF, 0xFF, 0xFF, 0xff},
}

func getColorNames(actualColor colorful.Color) (result []string) {
	var maxDistance = 0.30

	for colorName, colorRGBA := range colorMap {
		colorColorful, _ := colorful.MakeColor(colorRGBA)
		currentDistance := colorColorful.DistanceRgb(actualColor)

		if maxDistance >= currentDistance {
			result = append(result, colorName)
		}
	}

	return result
}

// GetColors returns color information for a given mediafiles.
func (m *MediaFile) GetColors() (colors []string, vibrantHex string, mutedHex string) {
	file, _ := os.Open(m.filename)

	defer file.Close()

	decodedImage, _, _ := image.Decode(file)
	centroids, e := prominentcolor.KmeansWithAll(5, decodedImage, prominentcolor.ArgumentDefault | prominentcolor.ArgumentAverageMean, prominentcolor.DefaultSize, prominentcolor.GetDefaultMasks() )
	if e != nil {
		log.Printf("Error while doing kmeans on color")
	}

	palette := vibrant.NewPaletteBuilder(decodedImage).Generate()

	if vibrantSwatch := palette.VibrantSwatch(); vibrantSwatch != nil {
		color, _ := colorful.MakeColor(vibrantSwatch.Color())
		//colors = append(colors, getColorNames(color)...)
		vibrantHex = color.Hex()
	}

	if mutedSwatch := palette.MutedSwatch(); mutedSwatch != nil {
		color, _ := colorful.MakeColor(mutedSwatch.Color())
		//colors = append(colors, getColorNames(color)...)
		mutedHex = color.Hex()
	}

	for _, centroid := range centroids {
		colorfulColor, _ := colorful.MakeColor(color.RGBA{uint8(centroid.Color.R), uint8(centroid.Color.G), uint8(centroid.Color.B), 0xff})
		colors = append(colors, getColorNames(colorfulColor)...)
	}

	sort.Strings(colors)

	return colors, vibrantHex, mutedHex
}
