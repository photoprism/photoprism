package colors

import "github.com/lucasb-eyer/go-colorful"

// Colorful finds the Color most similar to the specified colorful.Color.
func Colorful(actualColor colorful.Color) (result Color) {
	var distance = 1.0

	for rgba, i := range ColorMap {
		colorColorful, _ := colorful.MakeColor(rgba)
		currentDistance := colorColorful.DistanceLab(actualColor)

		if distance >= currentDistance {
			distance = currentDistance
			result = i
		}
	}

	return result
}
