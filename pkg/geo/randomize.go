package geo

import (
	"math/rand"
	"time"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

// Randomize adds a random offset to a value.
func Randomize(value, diameter float64) float64 {
	return value + (r.Float64()-0.5)*diameter
}
