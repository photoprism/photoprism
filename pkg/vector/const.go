package vector

import "math"

const (
	Epsilon = math.SmallestNonzeroFloat64
)

func NaN() float64 {
	return math.NaN()
}
