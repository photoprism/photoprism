package frame

import (
	"math/rand"
)

// RandomAngle returns a random angle between -max and max.
func RandomAngle(max float64) float64 {
	if max == 0 {
		return 0
	}

	if max < 0 {
		max = -1 * max
	}

	if max > 180 {
		max = 180
	}

	r := 2 * max

	return (rand.Float64() - 0.5) * r
}
