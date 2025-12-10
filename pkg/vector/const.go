package vector

import "math"

const (
	// Epsilon is the smallest non-zero float used as a numerical tolerance.
	Epsilon = math.SmallestNonzeroFloat64
)

// NaN returns a quiet NaN value.
func NaN() float64 {
	return math.NaN()
}
