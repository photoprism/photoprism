package geo

import (
	"math/rand/v2"
)

// Randomize adds a random offset to a value.
func Randomize(value, diameter float64) float64 {
	return value + (rand.Float64()-0.5)*diameter
}
